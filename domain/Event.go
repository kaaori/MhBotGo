package domain

import (
	"github.com/bwmarrin/discordgo"
)

// Event : The model for the Events table
type Event struct {
	ServerID                  string
	EventID                   int64
	CreatorID                 string
	EventLocation             string
	HostName                  string
	CreationTimestamp         int64
	StartTimestamp            int64
	LastAnnouncementTimestamp int64
	DurationMinutes           int64
	Name                      string

	// ORM Fields
	Creator *discordgo.User
	Server  DiscordServer
}
