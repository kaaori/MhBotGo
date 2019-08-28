package domain

import (
	"github.com/bwmarrin/discordgo"
)

// Birthday : Holds the relevant information for a user's birthday
type Birthday struct {
	ServerID    string
	BirthdayID  int32
	GuildUserID string

	// ORM Fields
	Server    DiscordServer
	GuildUser discordgo.User
}
