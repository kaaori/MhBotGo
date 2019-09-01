package main

import (
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/araddon/dateparse"
	"github.com/snabb/isoweek"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/chrome"
	"github.com/kaaori/mhbotgo/domain"
	"github.com/kaaori/mhbotgo/util"
	config "github.com/spf13/viper"
)

func installCommands(session *discordgo.Session) {
	prefix = config.GetString("prefix")
	router := exrouter.New()

	// router command template
	router.On("commandName", func(ctx *exrouter.Context) {
		// Command code
		ctx.Reply("Reply text here!")
	}).Desc("Descriptive text")

	router.On("gctest", func(ctx *exrouter.Context) {
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

		BotInstance.ClientSession.ChannelMessageSendComplex(ctx.Msg.ChannelID, ms)
	}).Desc("Descriptive text")

	router.On("gctemplate", func(ctx *exrouter.Context) {
		tmpl, err := template.ParseFiles("./web/schedule-template.html")
		if err != nil {
			panic(err)
		}
		year, week := time.Now().ISOWeek()
		t := isoweek.StartTime(year, week, time.Now().Location())

		g, _ := ctx.Guild(ctx.Msg.GuildID)
		f, err := os.Create("./web/schedule-parsed.html")
		if err != nil {
			log.Error("create file: ", err)
			return
		}
		events, err := BotInstance.EventDao.GetAllEventsForServer(ctx.Msg.GuildID)
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
				monEvts = append(monEvts, &domain.EventView{
					PrettyPrint:    el.ToString(),
					StartTimestamp: el.StartTimestamp,
					HasPassed:      time.Now().UTC().After(el.StartTime),
					DayOfWeek:      el.StartTime.Weekday().String()})
				break
			case time.Tuesday:
				tuesEvts = append(monEvts, &domain.EventView{
					PrettyPrint:    el.ToString(),
					StartTimestamp: el.StartTimestamp,
					HasPassed:      time.Now().UTC().After(el.StartTime),
					DayOfWeek:      el.StartTime.Weekday().String()})
				break
			case time.Wednesday:
				wedEvts = append(monEvts, &domain.EventView{
					PrettyPrint:    el.ToString(),
					StartTimestamp: el.StartTimestamp,
					HasPassed:      time.Now().UTC().After(el.StartTime),
					DayOfWeek:      el.StartTime.Weekday().String()})
				break
			case time.Thursday:
				thursEvts = append(monEvts, &domain.EventView{
					PrettyPrint:    el.ToString(),
					StartTimestamp: el.StartTimestamp,
					HasPassed:      time.Now().UTC().After(el.StartTime),
					DayOfWeek:      el.StartTime.Weekday().String()})
				break
			case time.Friday:
				friEvts = append(monEvts, &domain.EventView{
					PrettyPrint:    el.ToString(),
					StartTimestamp: el.StartTimestamp,
					HasPassed:      time.Now().UTC().After(el.StartTime),
					DayOfWeek:      el.StartTime.Weekday().String()})
				break
			case time.Saturday:
				satEvts = append(monEvts, &domain.EventView{
					PrettyPrint:    el.ToString(),
					StartTimestamp: el.StartTimestamp,
					HasPassed:      time.Now().UTC().After(el.StartTime),
					DayOfWeek:      el.StartTime.Weekday().String()})
				break
			case time.Sunday:
				sunEvts = append(monEvts, &domain.EventView{
					PrettyPrint:    el.ToString(),
					StartTimestamp: el.StartTimestamp,
					HasPassed:      time.Now().UTC().After(el.StartTime),
					DayOfWeek:      el.StartTime.Weekday().String()})
				break

			}
		}
		_, isoWeek := t.ISOWeek()
		firstDayOfWeek := util.FirstDayOfISOWeek(t.Year(), isoWeek, t.Location())
		data := domain.ScheduleView{
			ServerName:        g.Name,
			CurrentWeekString: string(firstDayOfWeek.Format("January 2, 2006") + " ── " + firstDayOfWeek.AddDate(0, 0, 7).Format("January 2, 2006")),
			Tz:                "<strong>Eastern Standard Time</strong>",
			MondayEvents:      monEvts,
			TuesdayEvents:     tuesEvts,
			WednesdayEvents:   wedEvts,
			ThursdayEvents:    thursEvts,
			FridayEvents:      friEvts,
			SaturdayEvents:    satEvts,
			SundayEvents:      sunEvts}

		tmpl.Execute(f, data)
		f.Close()

		chrome.TakeScreenshot()

		f, err = os.Open("schedule.png")
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

		BotInstance.ClientSession.ChannelMessageSendComplex(ctx.Msg.ChannelID, ms)
	}).Desc("Descriptive text")

	router.On("events", func(ctx *exrouter.Context) {
		// Matches all in quotes
		// (["'])(?:(?=(\\?))\2.)*?\1
		switch strings.ToLower(ctx.Args.Get(1)) {
		case "add":
			event := new(domain.Event)
			event.ServerID = ctx.Msg.GuildID
			event.CreatorID = ctx.Msg.Author.ID
			event.CreationTimestamp = time.Now().Unix()
			event.DurationMinutes = 120
			if !validateNewEventArgs(ctx, event) {
				log.Error("Error validating event args")
				return
			}

			event = BotInstance.EventDao.InsertEvent(event, BotInstance.ClientSession)
			if event == nil {
				log.Error("Error getting event after insert")
				return
			}

			embed := getEmbedFromEvent(event, ctx, "scheduled for ")
			BotInstance.ClientSession.ChannelMessageSendEmbed(ctx.Msg.ChannelID, embed)
			return
		case "remove":
			t, isValid := validateDateString(ctx, ctx.Args.Get(2))
			if !isValid {
				// The method call above handles outputting the error to the user and console.
				return
			}

			referencedEvent, err := BotInstance.EventDao.GetEventByStartTime(t.Unix())
			if err != nil || referencedEvent == nil {
				ctx.Reply("Could not find that event, please try again")
				return
			}
			BotInstance.EventDao.DeleteEventByID(referencedEvent.EventID)
			embed := getEmbedFromEvent(referencedEvent, ctx, "deleted from ")
			BotInstance.ClientSession.ChannelMessageSendEmbed(ctx.Msg.ChannelID, embed)

			break
		// case "edit":
		// 	break
		default:
			ctx.Reply(config.GetString("badSyntaxError"))
			break

		}
	})

	router.On("test", func(ctx *exrouter.Context) {
		servers, err := BotInstance.ServerDao.GetAllServers()
		if err != nil {
			log.Error("Error calling servers", err)
		}

		log.Trace("Attempting to find servers")
		userCount := 0
		guildCount := 0
		fields := make([]*discordgo.MessageEmbedField, 0)
		for _, element := range servers {
			if err != nil {
				log.Error("Error retrieving guild, probably doesn't exist?")
				continue
			}

			guildCount++
			userCount += element.Guild.MemberCount
		}
		fieldTest := util.GetField(
			"Currently Serving",
			strconv.Itoa(userCount)+" Members, and "+strconv.Itoa(guildCount)+" Guild(s)",
			false)

		fields = append(fields, fieldTest)
		outputEmbed := util.GetEmbed("Test Command!", "Footer!", false, fields...)
		BotInstance.ClientSession.ChannelMessageSendEmbed(ctx.Msg.ChannelID, outputEmbed)
		// ctx.ReplyEmbed(outputEmbed)
		// ctx.Reply("Currently serving " + strconv.Itoa(userCount) + " users")
	}).Desc("Test command")

	session.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		router.FindAndExecute(session, prefix, session.State.User.ID, m.Message)
	})

	log.Info("Commands installed.")
}

func getEmbedFromEvent(event *domain.Event, ctx *exrouter.Context, eventEmbedText string) *discordgo.MessageEmbed {
	t := time.Unix(event.StartTimestamp, 0)
	timeObj := t.Format("January 2, 2006")
	baseField := util.GetField("Event "+eventEmbedText+timeObj, event.ToEmbedString(), false)
	baseEmbed := util.GetEmbed("", "Test footer", false, baseField)
	return baseEmbed
}

func validateDateString(ctx *exrouter.Context, dateString string) (time.Time, bool) {
	if dateString := ctx.Args.Get(2); "" != dateString {
		t, err := dateparse.ParseAny(dateString)
		if err != nil {
			log.Error("Invalid time format? ", err)
			ctx.Reply("Please check your date format and try again")
			return t, false
		}
		return t, true
	}
	return time.Now(), false

}

func validateNewEventArgs(ctx *exrouter.Context, event *domain.Event) bool {
	// TODO: Add descriptive examples to errors \/
	// TODO: Validate el as date
	t, isValid := validateDateString(ctx, strconv.FormatInt(event.StartTimestamp, 10))
	if !isValid {
		// The method call above handles outputting the error to the user and console.
		return false
	}

	event.StartTimestamp = t.UTC().Unix()

	if hostName := ctx.Args.Get(3); "" != hostName {
		event.HostName = hostName
	} else {
		ctx.Reply("Please ensure you have included a host to your event")
		return false
	}
	if name := ctx.Args.Get(4); "" != name {
		event.EventName = name
	} else {
		ctx.Reply("Please ensure you have given the event a name")
		return false
	}
	if location := ctx.Args.Get(5); "" != location {
		event.EventLocation = location
	} else {
		ctx.Reply("Please ensure you have given the event a location")
		return false
	}
	return true
}
