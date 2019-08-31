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
	//• 10:00 PM (Eastern Standard Time) ── BigScreen Hangout ── (Hosted by Haughty:uwu_cowboy: in BigScreen Beta)
	return "• *" + e.StartTime.UTC().Format(time.Kitchen) + " (Eastern Standard Time)* ── **" + e.EventName + "** ── (Hosted  by ***" + e.HostName + "*** in " + e.EventLocation + ")"
}
