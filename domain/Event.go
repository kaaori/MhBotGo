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
	TzOffset                  int64
	TzLoc                     *time.Location

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
	return "<em>" + e.StartTime.In(e.TzLoc).Format(time.Kitchen) +
		" (Eastern Standard Time)</em> ── <strong>" + e.EventName +
		"</strong> ── (Hosted  by <strong><em>" + e.HostName + "</em></strong> in " + e.EventLocation + ")"
}

// ToEmbedString : Provides a pretty-print string for the event in a discord embed
// Est Loc offset is calculated here as this is before it touches the DB and is adjusted
func (e *Event) ToEmbedString() string {
	return "• *" + time.Unix(e.StartTimestamp-e.TzOffset, 0).Format(time.Kitchen) +
		" (Eastern Standard Time)* ── **" + e.EventName + "** ── (Hosted  by ***" + e.HostName +
		"*** in " + e.EventLocation + ")"
}

func (e *Event) ToAnnounceString() string {
	return "**" + e.HostName + "** is about to start this event in " +
		e.EventLocation + " at **" + e.StartTime.Format("3:04 PM") +
		" (Eastern Standard Time)!**"
}

func (e *Event) ToStartingString() string {
	return "Join up on **" + e.HostName + "**! This event is taking place at " +
		e.StartTime.Format("3:04 PM") + ", and will last roughly 2 hours"
}
