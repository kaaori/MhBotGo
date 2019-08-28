package domain

import (
	"github.com/bwmarrin/discordgo"
)

// Event : The model for the Events table
type Event struct {
	ServerID                  string
	EventID                   int32
	CreatorID                 string
	EventLocation             string
	HostName                  string
	CreationTimestamp         int32
	StartTimestamp            int32
	LastAnnouncementTimestamp int32
	DurationMinutes           int32

	// ORM Fields
	Creator discordgo.User
	Server  DiscordServer
}
