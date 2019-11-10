package domain

import (
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/kaaori/MhBotGo/util"
)

// Birthday : Holds the relevant information for a user's birthday
type Birthday struct {
	ServerID      string
	BirthdayID    int64
	GuildUserID   string
	BirthdayMonth int
	BirthdayDay   int
	LastSetUnix   int64

	// ORM Fields
	Server      DiscordServer
	GuildUser   *discordgo.User
	LastSetTime time.Time
}

// ToString : Provides a pretty-print string for the event
func (e Birthday) ToString() string {
	emojiRegex := regexp.MustCompile("<(a)?:.*?:(.*?)>")
	return "<strong><em> " + strip.StripTags(emojiRegex.ReplaceAllString(e.GuildUser.Username, "")) + "'s</strong></em> birthday!"
}

func (b Birthday) IsBirthdayInCurrentWeek() bool {
	t := time.Date(time.Now().Year(), time.Month(b.BirthdayMonth), b.BirthdayDay, 0, 0, 0, 0, time.Now().Location())

	// Return if our birthday&month is within our week
	// i.e not before the current week's monday and not after the curren week's monday + 7 days
	return !t.Before(util.GetCurrentWeekFromMondayAsTime()) && !t.After(util.GetCurrentWeekFromMondayAsTime().AddDate(0, 0, 7))
}

func (b Birthday) IsToday() bool {
	t := time.Date(time.Now().Year(), time.Month(b.BirthdayMonth), b.BirthdayDay, 0, 0, 0, 0, time.Now().Location())

	now := time.Now()
	// Return if our birthday & month match
	return now.Day() == t.Day() && now.Month() == t.Month()
}

func (b Birthday) GetTimeFromBirthday() time.Time {
	return time.Date(time.Now().Year(), time.Month(b.BirthdayMonth), int(time.Weekday(b.BirthdayDay)), time.Now().Hour(), 0, 0, 0, time.Now().Location())
}
