package dao

import (
	"errors"
	"log"
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"github.com/bwmarrin/discordgo"
	"mhbotgo.com/domain"
	"mhbotgo.com/util"
)

// DiscordServerDao : Contains data access methods for stored discord servers
type DiscordServerDao struct {
	Session *discordgo.Session
}

// GetServerByID : Gets a guild by its ID
func (d *DiscordServerDao) GetServerByID(ID string) (*domain.DiscordServer, error) {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select * from Servers where ServerID = ?"
	// server := new(domain.DiscordSer  ver)

	stmt, err := queryForRows(query, DB, ID)
	if err != nil {
		return nil, err
	}
	if stmt == nil {
		return nil, errors.New("statement was closed")
	}
	defer stmt.Close()

	server, err := getServerFromStmt(stmt, d)
	if err != nil {
		log.Println("Error getting server by ID", err)
		return nil, err
	}
	return server, err
}

// GetAllServers : Gets all the servers in the database
func (d *DiscordServerDao) GetAllServers() ([]*domain.DiscordServer, error) {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select * from Servers"
	servers := make([]*domain.DiscordServer, 0)

	stmt, err := queryForRows(query, DB)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for {
		server, err := getServerFromStmt(stmt, d)
		if err != nil {
			log.Println("Error getting server by ID", err)
			return nil, err
		}
		if server != nil {
			servers = append(servers, server)
		} else if len(servers) > 0 {
			break
		} else {
			return nil, errors.New("Couldn't find server")
		}
	}

	return servers, err
}

// InsertNewServer : Sets up the initial data for a guild
func (d *DiscordServerDao) InsertNewServer(serverID string) int64 {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "insert into Servers (ServerID, JoinTimeUnix) values (?,?)"
	stmt, err := DB.Prepare(query)
	if err != nil {
		log.Println("Error inserting guild", err)
		return -1
	}
	defer stmt.Close()
	err = stmt.Exec(serverID, time.Now().Unix()-util.ServerLocOffset)
	if err != nil {
		log.Println("Error inserting guild", err)
		return -1
	}
	return DB.LastInsertRowID()
}

func mapRowToServer(rows *sqlite3.Stmt, s *discordgo.Session) (domain.DiscordServer, error) {
	var server domain.DiscordServer
	err := rows.Scan(&server.ServerID, &server.JoinTimeUnix)
	if err != nil {
		return server, err
	}
	server.Guild, err = s.State.Guild(server.ServerID)
	if err != nil {
		log.Println("Error finding guild: ", err)
		return server, err
	}

	return server, err
}

func getServerFromStmt(stmt *sqlite3.Stmt, d *DiscordServerDao) (*domain.DiscordServer, error) {
	hasRow, err := stmt.Step()
	if err != nil {
		return nil, err
	}

	if !hasRow {
		return nil, nil
	}
	server, err := mapRowToServer(stmt, d.Session)
	if err != nil {
		return nil, err
	}
	return &server, err
}
