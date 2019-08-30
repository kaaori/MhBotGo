package dao

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/domain"
)

type DiscordServerDao struct {
	Session *discordgo.Session
}

// GetAllServers : Gets all the servers in the database
func (d *DiscordServerDao) GetAllServers() ([]domain.DiscordServer, error) {
	query := "select * from Servers"
	servers := make([]domain.DiscordServer, 0)

	db := get()
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		return servers, err
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

func mapRowToServer(rows sql.Rows, s *discordgo.Session) (domain.DiscordServer, error) {
	var currentServer domain.DiscordServer
	err := rows.Scan(&currentServer.ServerID, &currentServer.JoinTimeUnix)
	if err != nil {
		return currentServer, err
	}
	// currentServer.Guild = s.Guild(currentServer.ServerID)
	currentServer.Guild, err = s.State.Guild(currentServer.ServerID)
	if err != nil {
		return currentServer, err
	}

	return currentServer, err
}
