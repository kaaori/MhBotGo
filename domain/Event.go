package domain

import (
	"regexp"
	"strconv"
	"time"

	strip "github.com/grokify/html-strip-tags-go"
)

// Event : The model for the Events table
type Event struct {
	ServerID                  string
	EventID                   string
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
	Server               DiscordServer
	CreationTime         time.Time
	StartTime            time.Time
	LastAnnouncementTime time.Time
	EndTime              time.Time
}

var (
	emojiRegex = regexp.MustCompile("<(a)?:.*?:(.*?)>")
)

// ToString : Provides a pretty-print string for the event
func (e *Event) ToString() string {
	return "<strong><em>" + e.StartTime.Format(time.Kitchen) +
		" - " + strip.StripTags(emojiRegex.ReplaceAllString(e.EventName, "")) +
		"</strong></em> (Hosted  by <strong><em>" + strip.StripTags(emojiRegex.ReplaceAllString(e.HostName, "")) + "</em></strong> in " + strip.StripTags(emojiRegex.ReplaceAllString(e.EventLocation, "")) + ")"
}

// ToEmbedString : Provides a pretty-print string for the event in a discord embed
// Server Loc offset is calculated here as this is before it touches the DB and is adjusted
func (e *Event) ToEmbedString() string {
	return "• *<t:" + strconv.FormatInt(e.StartTime.Unix(), 10) +
		">* ── **" + strip.StripTags(emojiRegex.ReplaceAllString(e.EventName, "")) + "** ── (Hosted  by ***" + strip.StripTags(emojiRegex.ReplaceAllString(e.HostName, "")) +
		"*** in " + strip.StripTags(emojiRegex.ReplaceAllString(e.EventLocation, "")) + ")"
}

// ToAnnounceString : Gets the string representing the pre-starting announcement for the given event
func (e *Event) ToAnnounceString() string {
	return "**" + strip.StripTags(emojiRegex.ReplaceAllString(e.HostName, "")) + "** is about to start this event in " +
		strip.StripTags(emojiRegex.ReplaceAllString(e.EventLocation, "")) + " at **<t:" + strconv.FormatInt(e.StartTime.Unix(), 10) +
		">!**"
}

// ToStartingString : Gets the string representing the starting announcement for the given event
func (e *Event) ToStartingString() string {
	return "Join up on **" + strip.StripTags(emojiRegex.ReplaceAllString(e.HostName, "")) + "**! This event is taking place at **<t:" +
		strconv.FormatInt(e.StartTime.Unix(), 10) + ">**, and will last roughly 2 hours"
}
