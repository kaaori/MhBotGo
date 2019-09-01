package domain

import (
	"time"

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
	EventName                 string

	// ORM Fields
	Creator              *discordgo.User
	Server               DiscordServer
	CreationTime         time.Time
	StartTime            time.Time
	LastAnnouncementTime time.Time
	EndTime              time.Time
}

// ToString : Provides a pretty-print string for the event
func (e *Event) ToString() string {
	return "<em>" + e.StartTime.UTC().Format(time.Kitchen) +
		" (Eastern Standard Time)</em> ── <strong>" + e.EventName +
		"</strong> ── (Hosted  by <strong><em>" + e.HostName + "</em></strong> in " + e.EventLocation + ")"
}

// ToEmbedString : Provides a pretty-print string for the event in a discord embed
func (e *Event) ToEmbedString() string {
	return "• *" + e.StartTime.UTC().Format(time.Kitchen) +
		" (Eastern Standard Time)* ── **" + e.EventName + "** ── (Hosted  by ***" + e.HostName +
		"*** in " + e.EventLocation + ")"
}
