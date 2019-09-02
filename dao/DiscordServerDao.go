package dao

import (
	"database/sql"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/domain"
	"github.com/kaaori/MhBotGo/util"
)

// DiscordServerDao : Contains data access methods for stored discord servers
type DiscordServerDao struct {
	Session *discordgo.Session
}

// GetServerByID : Gets a guild by its ID
func (d *DiscordServerDao) GetServerByID(ID string) (*domain.DiscordServer, error) {
	query := "select * from Servers where ServerID = ?"
	// server := new(domain.DiscordSer  ver)

	db := get()
	defer db.Close()

	statement, _ := db.Prepare(query)

	rows, err := queryForRowsWithParams(statement, db, ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		log.Error("No guild by id " + ID + " found")
		return nil, err
	}

	server, err := mapRowToServer(*rows, d.Session)
	if err != nil {
		return nil, err
	}

	return &server, err
}

// GetAllServers : Gets all the servers in the database
func (d *DiscordServerDao) GetAllServers() ([]domain.DiscordServer, error) {
	query := "select * from Servers"
	servers := make([]domain.DiscordServer, 0)

	db := get()
	defer db.Close()

	rows, err := queryForRows(query, db)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		server, err := mapRowToServer(*rows, d.Session)
		if err != nil {
			return nil, err
		}

		servers = append(servers, server)
	}

	return servers, err
}

// InsertNewServer : Sets up the initial data for a guild
func (d *DiscordServerDao) InsertNewServer(serverID string) int64 {
	query := "insert into Servers (ServerID, JoinTimeUnix) values (?,?)"
	db := get()
	defer db.Close()

	statement, _ := db.Prepare(query)
	statementResult := executeQueryWithParams(statement, db, serverID, time.Now().Unix()-util.EstLocOffset)

	if rowsAffected, _ := statementResult.RowsAffected(); rowsAffected < 0 {
		log.Error("Error inserting server")
		return -1
	}
	lastID, _ := statementResult.LastInsertId()
	return lastID
}

func mapRowToServer(rows sql.Rows, s *discordgo.Session) (domain.DiscordServer, error) {
	var server domain.DiscordServer
	err := rows.Scan(&server.ServerID, &server.JoinTimeUnix)
	if err != nil {
		return server, err
	}
	// server.Guild = s.Guild(server.ServerID)
	server.Guild, err = s.State.Guild(server.ServerID)
	if err != nil {
		return server, err
	}

	return server, err
}
