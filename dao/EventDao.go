package dao

import (
	"errors"
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/domain"
	"github.com/kaaori/MhBotGo/util"
)

// EventDao : Contains data access methods for stored server events
type EventDao struct {
	Session *discordgo.Session
}

// DeleteAllEventsForServer : Nukes everything for a server
func (d *EventDao) DeleteAllEventsForServer(serverID string) error {
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
func (d *EventDao) GetAllEventsForServerForWeek(serverID string, weekTime time.Time) ([]*domain.Event, error) {
	query := "select * from Events where ServerID = ? and StartTimestamp between ? AND ?"

	stmt, err := queryForRows(query, DB, serverID, weekTime.Unix(), weekTime.AddDate(0, 0, 7).Unix())
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return processStmtForEventArray(stmt, d)
}

func getEventFromStmt(stmt *sqlite3.Stmt, d *EventDao) (*domain.Event, error) {
	hasRow, err := stmt.Step()
	if err != nil {
		return nil, err
	}

	if !hasRow {
		return nil, nil
	}
	event, err := mapRowToEvent(stmt, d.Session)
	if err != nil {
		return nil, err
	}
	return event, err
}

// GetEventsCountForServerForWeek : Gets the # of a server's events within a week range
func (d *EventDao) GetEventsCountForServerForWeek(serverID string, weekTime time.Time) int {
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
	query := "select Count(*) from Events ORDER by (julianday(DATETIME('NOW')) - julianday(StartTimestamp)) desc LIMIT 1;"
	// -2*3600 for 2 hr buffer between midnight & an events start
	stmt, err := queryForRows(query, DB)
	if err != nil {
		return -1
	}
	defer stmt.Close()

	return getCountFromStmt(stmt)
}

// GetAllEventCounts : Get a count of all events globally
func (d *EventDao) GetAllEventCounts() int {
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
	query := "select Count(*) from Events where ServerID = ?"

	stmt, err := queryForRows(query, DB, serverID)
	if err != nil {
		return -1
	}
	defer stmt.Close()

	return getCountFromStmt(stmt)
}

// GetAllEventsForServer : Gets all the events by a given server ID
func (d *EventDao) GetAllEventsForServer(serverID string) ([]*domain.Event, error) {
	query := "select * from Events where ServerID = ?"
	stmt, err := queryForRows(query, DB, serverID)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return processStmtForEventArray(stmt, d)
}

// GetNextEventOrDefault : Gets the next occuring event or nil
func (d *EventDao) GetNextEventOrDefault(guildID string) (*domain.Event, error) {
	// unixNowInLoc := time.Now().Unix() - util.ServerLocOffset
	query := "select * from Events where ServerID = ? ORDER by abs(StartTimestamp- strftime('%s', 'now') ) limit 1"
	// query := "select * from Events where ServerID = ? and StartTimestamp < ? order by StartTimestamp desc"

	stmt, err := queryForRows(query, DB, guildID)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return processStmtForEvent(stmt, d)
}

// GetEventByID : Gets an event by its ID
func (d *EventDao) GetEventByID(ID string) (*domain.Event, error) {
	query := "select * from Events where EventID = ?"
	stmt, err := queryForRows(query, DB, ID)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return getEventFromStmt(stmt, d)
}

// GetEventByStartTime : Gets an event by its start time and server
func (d *EventDao) GetEventByStartTime(guildID string, startTime int64) (*domain.Event, error) {
	query := "select * from Events where StartTimestamp = ? and ServerID = ? limit 1"

	stmt, err := queryForRows(query, DB, startTime, guildID)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return getEventFromStmt(stmt, d)
}

// UpdateEvent : Updates an event by object
// Returns ID of new event
func (d *EventDao) UpdateEvent(event *domain.Event) int64 {
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
		log.Error("Error updating event", err)
		return -1
	}
	defer stmt.Close()

	stmt.Exec()

	return DB.LastInsertRowID()
}

// DeleteEventByID : Deletes an event by ID
// Returns ID of deleted event
func (d *EventDao) DeleteEventByID(ID int64) int64 {
	query := `	DELETE FROM Events  
				WHERE EventID = ?`

	stmt, err := DB.Prepare(query, ID)
	if err != nil {
		log.Error("Error deleting event")
		return -1
	}
	stmt.Exec()

	return DB.LastInsertRowID()
}

// DeleteEventByStartTime : Deletes an event by given start time
// Returns ID of deleted event
func (d *EventDao) DeleteEventByStartTime(startTime int64) int64 {
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
func (d *EventDao) InsertEvent(event *domain.Event, s *discordgo.Session) *domain.Event {

	query := `INSERT INTO Events 
	(ServerID, CreatorID, EventName, EventLocation, HostName, CreationTimestamp, StartTimestamp, LastAnnouncementTimestamp, DurationMinutes) 
	VALUES
	(?,?,?,?,?,?,?,?,?)`

	stmt, err := DB.Prepare(query)
	if err != nil {
		log.Error("Error inserting server", err)

		return nil
	}
	defer stmt.Close()

	err = stmt.Exec(event.ServerID, event.CreatorID, event.EventName, event.EventLocation, event.HostName,
		event.CreationTimestamp-util.ServerLocOffset, event.StartTimestamp-util.ServerLocOffset, -1, event.DurationMinutes)
	if err != nil {
		log.Error("Error inserting event", err)
		return nil
	}

	event, err = mapORMFields(event, s)
	if err != nil {
		log.Error("Error mapping ORM fields in new event", err)
		return nil
	}
	event.EventID = DB.LastInsertRowID()

	return event
}

func mapRowToEvent(rows *sqlite3.Stmt, s *discordgo.Session) (*domain.Event, error) {
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

	event, err = mapORMFields(event, s)
	if err != nil {
		log.Error("Error mapping ORM Fields in event", err)
		return nil, err
	}
	return event, err
}

func mapORMFields(event *domain.Event, s *discordgo.Session) (*domain.Event, error) {
	guild, err := s.State.Guild(event.ServerID)
	if err != nil {
		log.Error("Could not find guild")
		return event, err
	}
	event.Server = domain.DiscordServer{Guild: guild, ServerID: guild.ID}

	creator, err := s.User(event.CreatorID)
	if err != nil {
		log.Error("Could not find user")
		return event, err
	}
	event.Creator = creator

	event.CreationTime = time.Unix(event.CreationTimestamp, 0).In(util.ServerLoc)
	event.StartTime = time.Unix(event.StartTimestamp, 0).In(util.ServerLoc)
	event.LastAnnouncementTime = time.Unix(event.LastAnnouncementTimestamp, 0)
	event.EndTime = event.StartTime.Add(time.Minute * time.Duration(event.DurationMinutes)).In(util.ServerLoc)
	event.TzOffset = util.ServerLocOffset
	event.TzLoc = util.ServerLoc
	return event, err
}

func getCountFromStmt(stmt *sqlite3.Stmt) int {
	hasRow, err := stmt.Step()
	if err != nil {
		return -1
	}

	if !hasRow {
		log.Error("No events found")
		return 0
	}
	var count int
	err = stmt.Scan(&count)
	if err != nil {
		return -1
	}
	return count
}

func processStmtForEventArray(stmt *sqlite3.Stmt, d *EventDao) ([]*domain.Event, error) {
	events := make([]*domain.Event, 0)
	for {
		event, err := getEventFromStmt(stmt, d)
		if err != nil {
			return nil, err
		}
		if event != nil {
			events = append(events, event)
		} else if len(events) > 0 {
			break
		} else {
			return nil, errors.New("Couldn't find event")
		}
	}
	return events, nil
}

func processStmtForEvent(stmt *sqlite3.Stmt, d *EventDao) (*domain.Event, error) {
	event, err := getEventFromStmt(stmt, d)
	if err != nil {
		return nil, err
	}
	return event, err
}
