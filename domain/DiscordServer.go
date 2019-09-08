package domain

import (
	"github.com/bwmarrin/discordgo"
)

// DiscordServer : Model of a Discord Server
type DiscordServer struct {
	ServerID     string
	JoinTimeUnix int64

	// ORM Fields
	Guild *discordgo.Guild
}
