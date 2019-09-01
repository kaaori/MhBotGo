package commands

import (
	"os"
	"sort"
	"strconv"
	"text/template"
	"time"

	"github.com/forestgiant/sliceutil"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/araddon/dateparse"
	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/bot"
	"github.com/kaaori/MhBotGo/chrome"
	"github.com/kaaori/MhBotGo/util"
	"github.com/kaaori/mhbotgo/domain"
	"github.com/snabb/isoweek"
)

// ParseTemplate : Parses a given guild's events
// Returns true if the schedule needs a refresh
func ParseTemplate(guildID string) {
	tmpl, err := template.ParseFiles("./web/schedule-template.html")
	if err != nil {
		panic(err)
	}
	year, week := time.Now().In(util.EstLoc).ISOWeek()
	t := isoweek.StartTime(year, week, time.Now().In(util.EstLoc).Location())

	g, _ := BotInstance.ClientSession.Guild(guildID)
	f, err := os.Create("./web/schedule-parsed.html")
	if err != nil {
		log.Error("create file: ", err)
		return
	}
	defer f.Close()

	events, err := BotInstance.EventDao.GetAllEventsForServer(guildID)
	if err != nil {
		log.Error("", err)
		return
	}

	monEvts := make([]*domain.EventView, 0)
	tuesEvts := make([]*domain.EventView, 0)
	wedEvts := make([]*domain.EventView, 0)
	thursEvts := make([]*domain.EventView, 0)
	friEvts := make([]*domain.EventView, 0)
	satEvts := make([]*domain.EventView, 0)
	sunEvts := make([]*domain.EventView, 0)

	for _, el := range events {
		dayOfWeek := el.StartTime.Weekday()
		switch dayOfWeek {
		case time.Monday:
			monEvts = appendEventToList(monEvts, el)
			break
		case time.Tuesday:
			tuesEvts = appendEventToList(tuesEvts, el)
			break
		case time.Wednesday:
			wedEvts = appendEventToList(wedEvts, el)
			break
		case time.Thursday:
			thursEvts = appendEventToList(thursEvts, el)
			break
		case time.Friday:
			friEvts = appendEventToList(friEvts, el)
			break
		case time.Saturday:
			satEvts = appendEventToList(satEvts, el)
			break
		case time.Sunday:
			sunEvts = appendEventToList(sunEvts, el)
			break

		}
	}
	_, isoWeek := t.ISOWeek()
	firstDayOfWeek := util.FirstDayOfISOWeek(t.Year(), isoWeek, t.Location())
	data := domain.ScheduleView{
		ServerName:        g.Name,
		CurrentWeekString: string(firstDayOfWeek.Format("January 2, 2006") + " ── " + firstDayOfWeek.AddDate(0, 0, 6).Format("January 2, 2006")),
		Tz:                "<strong>Eastern Standard Time</strong>",
		MondayEvents:      sortEventList(monEvts),
		TuesdayEvents:     sortEventList(tuesEvts),
		WednesdayEvents:   sortEventList(wedEvts),
		ThursdayEvents:    sortEventList(thursEvts),
		FridayEvents:      sortEventList(friEvts),
		SaturdayEvents:    sortEventList(satEvts),
		SundayEvents:      sortEventList(sunEvts)}

	tmpl.Execute(f, data)
}

func sortEventList(evts []*domain.EventView) []*domain.EventView {
	sort.Slice(evts, func(i, j int) bool {
		return evts[i].StartTimestamp < evts[j].StartTimestamp
	})
	return evts
}

func appendEventToList(targetList []*domain.EventView, e *domain.Event) []*domain.EventView {
	return append(targetList, &domain.EventView{
		PrettyPrint:    e.ToString(),
		StartTimestamp: e.StartTimestamp,
		HasPassed:      time.Now().In(util.EstLoc).After(e.StartTime),
		DayOfWeek:      e.StartTime.Weekday().String()})
}

func parseAndSendSched(ctx *exrouter.Context) {
	ParseTemplate(ctx.Msg.GuildID)
	g, _ := ctx.Guild(ctx.Msg.GuildID)
	channel := FindSchedChannel(g, BotInstance)

	SendSchedule(channel.ID, BotInstance)
}

func addEvent(ctx *exrouter.Context) bool {
	event := new(domain.Event)
	event.ServerID = ctx.Msg.GuildID
	event.CreatorID = ctx.Msg.Author.ID
	event.CreationTimestamp = time.Now().Unix()
	event.DurationMinutes = 120
	if !validateNewEventArgs(ctx, event) {
		log.Error("Error validating event args")
		return false
	}

	event = BotInstance.EventDao.InsertEvent(event, BotInstance.ClientSession)
	if event == nil {
		log.Error("Error getting event after insert")
		return false
	}

	embed := GetEmbedFromEvent(event, "scheduled for ")
	BotInstance.ClientSession.ChannelMessageSendEmbed(ctx.Msg.ChannelID, embed)
	return true
}

func removeEvent(ctx *exrouter.Context) bool {
	t, isValid := validateDateString(ctx, ctx.Args.Get(2))
	if !isValid {
		// The method call above handles outputting the error to the user and console.
		return false
	}

	referencedEvent, err := BotInstance.EventDao.GetEventByStartTime(t.Unix())
	if err != nil || referencedEvent == nil {
		ctx.Reply("Could not find that event, please try again")
		return false
	}
	BotInstance.EventDao.DeleteEventByID(referencedEvent.EventID)
	embed := GetEmbedFromEvent(referencedEvent, "deleted from ")
	BotInstance.ClientSession.ChannelMessageSendEmbed(ctx.Msg.ChannelID, embed)
	return true
}

func validateNewEventArgs(ctx *exrouter.Context, event *domain.Event) bool {
	// TODO: Add descriptive examples to errors \/
	// TODO: Validate el as date
	t, isValid := validateDateString(ctx, strconv.FormatInt(event.StartTimestamp, 10))
	if !isValid {
		// The method call above handles outputting the error to the user and console.
		return false
	}

	event.StartTimestamp = t.Unix()

	if hostName := ctx.Args.Get(2); "" != hostName {
		event.HostName = hostName
	} else {
		ctx.Reply("Please ensure you have included a host to your event")
		return false
	}
	if name := ctx.Args.Get(3); "" != name {
		event.EventName = name
	} else {
		ctx.Reply("Please ensure you have given the event a name")
		return false
	}
	if location := ctx.Args.Get(4); "" != location {
		event.EventLocation = location
	} else {
		ctx.Reply("Please ensure you have given the event a location")
		return false
	}
	return true
}

func GetEmbedFromEvent(event *domain.Event, eventEmbedText string) *discordgo.MessageEmbed {
	t := time.Unix(event.StartTimestamp, 0)
	timeObj := t.In(util.EstLoc).Format("January 2, 2006")
	baseField := util.GetField("Event "+eventEmbedText+timeObj, event.ToEmbedString(), false)
	baseEmbed := util.GetEmbed("", "Test footer", false, baseField)
	return baseEmbed
}

func GetAnnounceEmbedFromEvent(event *domain.Event, eventEmbedText string, eventEmbedTitle string) *discordgo.MessageEmbed {
	baseField := util.GetField(eventEmbedTitle, eventEmbedText, false)
	baseEmbed := util.GetEmbed("", "", false, baseField)
	return baseEmbed
}

func validateDateString(ctx *exrouter.Context, dateString string) (time.Time, bool) {
	if dateString := ctx.Args.Get(1); "" != dateString {
		t, err := dateparse.ParseAny(dateString)
		if err != nil {
			log.Error("Invalid time format? ", err)
			ctx.Reply("Please check your date format and try again")
			return time.Now().In(util.EstLoc), false
		}
		return t.In(util.EstLoc), true
	}
	return time.Now().In(util.EstLoc), false

}

// Auth : Authenticates against bot roles in config
func Auth(fn exrouter.HandlerFunc) exrouter.HandlerFunc {
	return func(ctx *exrouter.Context) {
		member, err := ctx.Member(ctx.Msg.GuildID, ctx.Msg.Author.ID)
		if err != nil {
			ctx.Reply("Could not fetch member: ", err)
		}
		var botAdminRole *discordgo.Role
		guildRoles, _ := ctx.Ses.GuildRoles(member.GuildID)
		for _, r := range guildRoles {
			if sliceutil.Contains(authRoles, r.Name) {
				botAdminRole = r
			}
		}
		if botAdminRole == nil {
			ctx.Reply("Oops, your server has not been configured properly!\n" +
				"I can't find the role named `EventRunner`.")
			return
		}

		if sliceutil.Contains(member.Roles, botAdminRole.ID) {
			ctx.Set("member", member)
			fn(ctx)
			return
		}

		ctx.Reply("You don't have permission to use this command")
	}
}

func FindSchedChannel(guild *discordgo.Guild, inst *bot.Instance) *discordgo.Channel {
	var schedChannel *discordgo.Channel
	for _, schedChannel = range guild.Channels {
		if schedChannel.Name == inst.ScheduleChannel {
			break
		}
		schedChannel = nil
	}
	if schedChannel == nil {
		log.Error("Sched channel not found")
		return nil
	}
	return schedChannel
}

func FindAnnouncementsChannel(guild *discordgo.Guild, inst *bot.Instance) *discordgo.Channel {
	var announcementChannel *discordgo.Channel
	for _, announcementChannel = range guild.Channels {
		if announcementChannel.Name == inst.AnnouncementChannel {
			break
		}
		announcementChannel = nil
	}
	if announcementChannel == nil {
		log.Error("Announcement channel not found")
		return nil
	}
	return announcementChannel
}

func GetSchedMessage(schedChannelID string, inst *bot.Instance) *discordgo.Message {
	msgHistory, _ := inst.ClientSession.ChannelMessages(schedChannelID, 100, "", "", "")

	var schedMsg *discordgo.Message
	for _, msg := range msgHistory {
		// Delete any extra schedules that may have been posted
		if schedMsg != nil {
			inst.ClientSession.ChannelMessageDelete(msg.ChannelID, msg.ID)
		} else if msg.Author.ID == inst.ClientSession.State.User.ID {
			log.Trace("Schedule msg found")
			schedMsg = msg
		}
	}
	return schedMsg
}

func SendSchedule(schedChannelID string, inst *bot.Instance) {
	schedMsg := GetSchedMessage(schedChannelID, inst)

	if schedMsg != nil {
		inst.ClientSession.ChannelMessageDelete(schedMsg.ChannelID, schedMsg.ID)
	}
	chrome.TakeScreenshot()

	f, err := os.Open("schedule.png")
	if err != nil {
		log.Error("Error getting schedule image", err)
		return
	}
	defer f.Close()

	ms := &discordgo.MessageSend{
		Files: []*discordgo.File{
			&discordgo.File{
				Name:   "schedule.png",
				Reader: f,
			},
		},
	}

	inst.ClientSession.ChannelMessageSendComplex(schedChannelID, ms)
}
