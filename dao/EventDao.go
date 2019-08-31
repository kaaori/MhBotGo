package dao

import (
	"database/sql"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/mhbotgo/domain"
)

// EventDao : Contains data access methods for stored server events
type EventDao struct {
	Session *discordgo.Session
}

// GetEventByID : Gets an event by its ID
func (d *EventDao) GetEventByID(ID string) (*domain.Event, error) {
	query := "select * from Events where EventID = ?"

	db := get()
	defer db.Close()

	statement, _ := db.Prepare(query)

	rows, err := queryForRowsWithParams(statement, db, ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		log.Error("No event by id " + ID + " found")
		return nil, err
	}

	event, err := mapRowToEvent(rows, d.Session)
	if err != nil {
		return nil, err
	}

	return event, err
}

// GetEventByStartTime : Gets an event by its start time
func (d *EventDao) GetEventByStartTime(startTime int64) (*domain.Event, error) {
	query := "select * from Events where StartTimestamp = ? limit 1"

	db := get()
	defer db.Close()

	statement, _ := db.Prepare(query)

	rows, err := queryForRowsWithParams(statement, db, startTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		log.Error("No event with start time" + time.Unix(startTime, 0).Format(time.RFC1123) + " found")
		return nil, err
	}

	event, err := mapRowToEvent(rows, d.Session)
	if err != nil {
		return nil, err
	}

	return event, err
}

// UpdateEvent : Updates an event by object
// Returns ID of new event
func (d *EventDao) UpdateEvent(event *domain.Event) int64 {
	query := `	UDPATE Events 
				SET 
					ServerID = ?,
					CreatorID = ?,
					EventLocation = ?,
					HostName = ?,
					CreationTimestamp = ?
					StartTimestamp = ?
					LastAnnouncementTimestamp = ?
					DurationMinutes = ?
				WHERE 
					EventID = ?`
	db := get()
	defer db.Close()

	statement, _ := db.Prepare(query)
	statementResult := executeQueryWithParams(statement, db,
		event.ServerID, event.CreatorID, event.EventLocation, event.HostName,
		event.CreationTimestamp, event.StartTimestamp, event.LastAnnouncementTimestamp, event.DurationMinutes,
		event.EventID)

	if rowsAffected, _ := statementResult.RowsAffected(); rowsAffected < 0 {
		log.Error("Error inserting server")
		return -1
	}
	lastID, _ := statementResult.LastInsertId()
	return lastID
}

// DeleteEventByID : Deletes an event by ID
// Returns ID of deleted event
func (d *EventDao) DeleteEventByID(ID int64) int64 {
	query := `	DELETE FROM Events  
				WHERE EventID = ?`
	db := get()
	defer db.Close()

	statement, _ := db.Prepare(query)
	statementResult := executeQueryWithParams(statement, db,
		ID)

	if rowsAffected, _ := statementResult.RowsAffected(); rowsAffected < 0 {
		log.Error("Error deleting event")
		return -1
	}
	lastID, _ := statementResult.LastInsertId()
	return lastID
}

// DeleteEventByStartTime : Deletes an event by given start time
// Returns ID of deleted event
func (d *EventDao) DeleteEventByStartTime(startTime int64) int64 {
	query := `	DELETE FROM Events  
				WHERE StartTimestamp = ?`
	db := get()
	defer db.Close()

	statement, _ := db.Prepare(query)
	statementResult := executeQueryWithParams(statement, db,
		startTime)

	if rowsAffected, _ := statementResult.RowsAffected(); rowsAffected < 0 {
		log.Error("Error deleting event by start time")
		return -1
	}
	lastID, _ := statementResult.LastInsertId()
	return lastID
}

// DeleteEventByStartTimeAndHost : Deletes an event by given start time and host
// Returns ID of deleted event
func (d *EventDao) DeleteEventByStartTimeAndHost(startTime int64, hostName string) int64 {
	query := `	DELETE FROM Events  
				WHERE StartTimestamp = ? 
				AND   HostName = ?`
	db := get()
	defer db.Close()

	statement, _ := db.Prepare(query)
	statementResult := executeQueryWithParams(statement, db,
		startTime, hostName)

	if rowsAffected, _ := statementResult.RowsAffected(); rowsAffected < 0 {
		log.Error("Error deleting event by start time")
		return -1
	}
	lastID, _ := statementResult.LastInsertId()
	return lastID
}

// InsertEvent : Insert a new event
func (d *EventDao) InsertEvent(event *domain.Event, s *discordgo.Session) *domain.Event {

	query := `INSERT INTO Events 
	(ServerID, CreatorID, EventName, EventLocation, HostName, CreationTimestamp, StartTimestamp, LastAnnouncementTimestamp, DurationMinutes) 
	VALUES
	(?,?,?,?,?,?,?,?,?)`
	db := get()
	defer db.Close()

	statement, _ := db.Prepare(query)
	statementResult := executeQueryWithParams(statement, db,
		event.ServerID, event.CreatorID, event.EventName, event.EventLocation, event.HostName,
		event.CreationTimestamp, event.StartTimestamp, -1, event.DurationMinutes)

	if rowsAffected, _ := statementResult.RowsAffected(); rowsAffected < 0 {
		log.Error("Error inserting server")
		return nil
	}
	event, err := mapORMFields(event, s)
	if err != nil {
		log.Error("Error mapping ORM fields in new event", err)
		return nil
	}
	lastID, _ := statementResult.LastInsertId()
	event.EventID = lastID

	return event
}

func mapRowToEvent(rows *sql.Rows, s *discordgo.Session) (*domain.Event, error) {
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
		return event, err
	}

	event, err = mapORMFields(event, s)
	if err != nil {
		log.Error("Error mapping ORM Fields in event", err)
		return event, err
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

	// TODO FIX THESE
	// For now I will just use the unix timestamp directly
	event.CreationTime = time.Unix(event.CreationTimestamp, 0)
	event.StartTime = time.Unix(event.StartTimestamp, 0)
	event.LastAnnouncementTime = time.Unix(event.LastAnnouncementTimestamp, 0)
	event.EndTime = event.StartTime.Add(time.Minute * time.Duration(event.DurationMinutes))
	return event, err
}