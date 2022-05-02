package dao

import (
	"log"
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"github.com/bwmarrin/discordgo"
	"mhbotgo.com/domain"
	"mhbotgo.com/profiler"
	"mhbotgo.com/util"
)

// EventDao : Contains data access methods for stored server events
type EventDao struct {
	Session *discordgo.Session
}

// DeleteAllEventsForServer : Nukes everything for a server
func (d *EventDao) DeleteAllEventsForServer(serverID string) error {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "delete from Events where ServerID = ?"

	stmt, err := queryForRows(query, DB, serverID)
	if err != nil {
		return err
	}
	defer stmt.Close()

	stmt.Exec()
	return nil
}

// GetAllEventsForServerForWeek : Gets a server's events within a week range
func (d *EventDao) GetAllEventsForServerForWeek(serverID string, weekTime time.Time, g *discordgo.Guild) ([]*domain.Event, error) {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select * from Events where ServerID = ? and StartTimestamp between ? AND ?"

	profiler.Start()
	stmt, err := queryForRows(query, DB, serverID, weekTime.Unix(), weekTime.AddDate(0, 0, 7).Unix())
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	profiler.StopAndPrintSeconds("Querying for rows")

	return processStmtForEventArray(stmt, d, g)
}

func getEventFromStmt(stmt *sqlite3.Stmt, d *EventDao, g *discordgo.Guild) (*domain.Event, error) {
	hasRow, err := stmt.Step()
	if err != nil {
		return nil, err
	}

	if !hasRow {
		return nil, nil
	}
	event, err := mapRowToEvent(stmt, d.Session, g)
	if err != nil {
		return nil, err
	}
	return event, err
}

// GetEventsCountForServerForWeek : Gets the # of a server's events within a week range
func (d *EventDao) GetEventsCountForServerForWeek(serverID string, weekTime time.Time) int {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select Count(*) cnt from Events where ServerID = ? and StartTimestamp between ? AND ?"
	// -2*3600 for 2 hr buffer between midnight & an events start
	stmt, err := queryForRows(query, DB, serverID, weekTime.Unix()+util.ServerLocOffset-(2*3600), weekTime.AddDate(0, 0, 6).Unix()+util.ServerLocOffset+(2*3600))
	if err != nil {
		return -1
	}
	defer stmt.Close()

	return getCountFromStmt(stmt)
}

// GetEventsCountForWeek : Gets the # of all events within a week range
func (d *EventDao) GetEventsCountForWeek(weekTime time.Time) int {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select Count(*) from Events where StartTimestamp between ? AND ?;"
	// -2*3600 for 2 hr buffer between midnight & an events start
	stmt, err := queryForRows(query, DB, weekTime.Unix()+util.ServerLocOffset-(2*3600), weekTime.AddDate(0, 0, 6).Unix()+util.ServerLocOffset+(2*3600))
	if err != nil {
		return -1
	}
	defer stmt.Close()

	return getCountFromStmt(stmt)
}

// GetAllEventCounts : Get a count of all events globally
func (d *EventDao) GetAllEventCounts() int {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select Count(*) from Events"

	stmt, err := queryForRows(query, DB)
	if err != nil {
		return -1
	}
	defer stmt.Close()

	return getCountFromStmt(stmt)
}

// GetEventCountForServer : Gets the total count of events for a server
func (d *EventDao) GetEventCountForServer(serverID string) int {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select Count(*) from Events where ServerID = ?"

	stmt, err := queryForRows(query, DB, serverID)
	if err != nil {
		return -1
	}
	defer stmt.Close()

	return getCountFromStmt(stmt)
}

// GetAllEventsForServer : Gets all the events by a given server ID
func (d *EventDao) GetAllEventsForServer(serverID string, g *discordgo.Guild) ([]*domain.Event, error) {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select * from Events where ServerID = ?"
	stmt, err := queryForRows(query, DB, serverID)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return processStmtForEventArray(stmt, d, g)
}

// GetNextEventOrDefault : Gets the next occuring event or nil
func (d *EventDao) GetNextEventOrDefault(guildID string, g *discordgo.Guild) (*domain.Event, error) {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select * from Events where ServerID = ? ORDER by abs(StartTimestamp- strftime('%s', 'now') ) limit 1"

	stmt, err := queryForRows(query, DB, guildID)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return processStmtForEvent(stmt, d, g)
}

// GetEventByID : Gets an event by its ID
func (d *EventDao) GetEventByID(ID string, g *discordgo.Guild, u *discordgo.User) (*domain.Event, error) {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select * from Events where EventID = ?"
	stmt, err := queryForRows(query, DB, ID)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return getEventFromStmt(stmt, d, g)
}

// GetEventByStartTime : Gets an event by its start time and server
func (d *EventDao) GetEventByStartTime(guildID string, startTime int64, g *discordgo.Guild) (*domain.Event, error) {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select * from Events where StartTimestamp = ? and ServerID = ? limit 1"

	stmt, err := queryForRows(query, DB, startTime, guildID)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return getEventFromStmt(stmt, d, g)
}

// UpdateEvent : Updates an event by object
// Returns ID of new event
// We need to offset the last announcement timestamp, as the rest will already have been offset upon insert
func (d *EventDao) UpdateEvent(event *domain.Event) int64 {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "UPDATE Events " +
		"	SET " +
		"		ServerID = ?," +
		"		CreatorID = ?," +
		"		EventLocation = ?," +
		"		HostName = ?," +
		"		CreationTimestamp = ?," +
		"		StartTimestamp = ?," +
		"		LastAnnouncementTimestamp = ?," +
		"		DurationMinutes = ?" +
		"	WHERE " +
		"		EventID = ?"

	stmt, err := DB.Prepare(query,
		event.ServerID, event.CreatorID, event.EventLocation, event.HostName,
		event.CreationTimestamp, event.StartTimestamp, event.LastAnnouncementTimestamp, event.DurationMinutes,
		event.EventID)
	if err != nil {
		log.Fatal("Error updating event", err)
		return -1
	}
	defer stmt.Close()

	stmt.Exec()

	return DB.LastInsertRowID()
}

// DeleteEventByID : Deletes an event by ID
// Returns ID of deleted event
func (d *EventDao) DeleteEventByID(ID int64) int64 {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := `	DELETE FROM Events  
				WHERE EventID = ?`

	stmt, err := DB.Prepare(query, ID)
	if err != nil {
		log.Fatal("Error deleting event")
		return -1
	}
	stmt.Exec()

	return DB.LastInsertRowID()
}

// DeleteEventByStartTime : Deletes an event by given start time
// Returns ID of deleted event
func (d *EventDao) DeleteEventByStartTime(startTime int64) int64 {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := `	DELETE FROM Events  
				WHERE StartTimestamp = ?`

	stmt, err := DB.Prepare(query, startTime)
	if err != nil {
		return -1
	}

	defer stmt.Close()

	stmt.Exec()
	return DB.LastInsertRowID()
}

// DeleteEventByStartTimeAndHost : Deletes an event by given start time and host
// Returns ID of deleted event
func (d *EventDao) DeleteEventByStartTimeAndHost(startTime int64, hostName string) int64 {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := `	DELETE FROM Events  
				WHERE StartTimestamp = ? 
				AND   HostName = ?`

	stmt, err := DB.Prepare(query, startTime, hostName)
	if err != nil {
		return -1
	}

	defer stmt.Close()

	stmt.Exec()
	return DB.LastInsertRowID()
}

// InsertEvent : Insert a new event
func (d *EventDao) InsertEvent(event *domain.Event, s *discordgo.Session, g *discordgo.Guild) *domain.Event {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := `INSERT INTO Events 
	(ServerID, CreatorID, EventName, EventLocation, HostName, CreationTimestamp, StartTimestamp, LastAnnouncementTimestamp, DurationMinutes) 
	VALUES
	(?,?,?,?,?,?,?,?,?)`

	stmt, err := DB.Prepare(query)
	if err != nil {
		log.Fatal("Error inserting server", err)

		return nil
	}
	defer stmt.Close()

	err = stmt.Exec(event.ServerID, event.CreatorID, event.EventName, event.EventLocation, event.HostName,
		event.CreationTimestamp, event.StartTimestamp, -1, event.DurationMinutes)
	if err != nil {
		log.Fatal("Error inserting event", err)
		return nil
	}

	event, err = mapEventORMFields(event, s, g)
	if err != nil {
		log.Fatal("Error mapping ORM fields in new event", err)
		return nil
	}
	event.EventID = DB.LastInsertRowID()

	return event
}

func mapRowToEvent(rows *sqlite3.Stmt, s *discordgo.Session, g *discordgo.Guild) (*domain.Event, error) {
	event := new(domain.Event)

	err := rows.Scan(
		&event.EventID,
		&event.ServerID,
		&event.CreatorID,
		&event.EventName,
		&event.EventLocation,
		&event.HostName,
		&event.CreationTimestamp,
		&event.StartTimestamp,
		&event.LastAnnouncementTimestamp,
		&event.DurationMinutes)
	if err != nil {
		return nil, err
	}
	profiler.Start()

	event, err = mapEventORMFields(event, s, g)
	if err != nil {
		log.Fatal("Error mapping ORM Fields in event", err)
		return nil, err
	}
	profiler.StopAndPrintSeconds("Mapping ORM fields")
	return event, err
}

func mapEventORMFields(event *domain.Event, s *discordgo.Session, g *discordgo.Guild) (*domain.Event, error) {

	event.Server = domain.DiscordServer{Guild: g, ServerID: g.ID}

	// event.Creator = creator

	event.CreationTime = time.Unix(event.CreationTimestamp-util.ServerLocOffset, 0).In(util.ServerLoc)
	event.StartTime = time.Unix(event.StartTimestamp-util.ServerLocOffset, 0).In(util.ServerLoc)
	if event.LastAnnouncementTimestamp < 0 {
		event.LastAnnouncementTime = time.Unix(event.LastAnnouncementTimestamp, 0)
	} else {
		event.LastAnnouncementTime = time.Unix(event.LastAnnouncementTimestamp-util.ServerLocOffset, 0)
	}
	event.EndTime = event.StartTime.Add(time.Minute * time.Duration(event.DurationMinutes)).In(util.ServerLoc)
	event.TzOffset = util.ServerLocOffset
	event.TzLoc = util.ServerLoc

	return event, nil
}

func getCountFromStmt(stmt *sqlite3.Stmt) int {
	hasRow, err := stmt.Step()
	if err != nil {
		return -1
	}

	if !hasRow {
		log.Fatal("No events found")
		return 0
	}
	var count int
	err = stmt.Scan(&count)
	if err != nil {
		return -1
	}
	return count
}

func processStmtForEventArray(stmt *sqlite3.Stmt, d *EventDao, g *discordgo.Guild) ([]*domain.Event, error) {
	events := make([]*domain.Event, 0)
	for {
		event, err := getEventFromStmt(stmt, d, g)
		if err != nil {
			return nil, err
		}
		if event != nil {
			events = append(events, event)
		} else if len(events) > 0 {
			break
		} else {
			return nil, nil // errors.New("Couldn't find event")
		}
	}
	return events, nil
}

func processStmtForEvent(stmt *sqlite3.Stmt, d *EventDao, g *discordgo.Guild) (*domain.Event, error) {
	event, err := getEventFromStmt(stmt, d, g)
	if err != nil {
		stmt.Close()
		return nil, err
	}
	return event, err
}
