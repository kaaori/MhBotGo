package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
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
			if t, isValid := validateDateString(ctx, ctx.Args.Get(2)); !isValid {
				return
			} else {
				referencedEvent, err := BotInstance.EventDao.GetEventByStartTime(t.Unix())
				if err != nil || referencedEvent == nil {
					ctx.Reply("Could not find that event, please try again")
					return
				}
				BotInstance.EventDao.DeleteEventByID(referencedEvent.EventID)
				embed := getEmbedFromEvent(referencedEvent, ctx, "deleted from ")
				BotInstance.ClientSession.ChannelMessageSendEmbed(ctx.Msg.ChannelID, embed)
			}
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
	baseField := util.GetField("Event "+eventEmbedText+timeObj, event.ToString(), false)
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
	if t, isValid := validateDateString(ctx, strconv.FormatInt(event.StartTimestamp, 10)); !isValid {
		return false
	} else {
		event.StartTimestamp = t.UTC().Unix()
	}

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
