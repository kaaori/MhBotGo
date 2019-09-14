package commands

import (
	"fmt"
	"time"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/bot"
	"github.com/kaaori/MhBotGo/util"
	logging "github.com/kaaori/mhbotgo/log"
	config "github.com/spf13/viper"
)

var (
	prefix    string
	log       = logging.NewLog()
	session   *discordgo.Session
	authRoles []string

	defaultScreenshotW = int64(1920)
	defaultScreenshotH = int64(1)

	// BotInstance : The instance of the bot containing the discord session and all relevant DAOs
	BotInstance *bot.Instance
)

func refreshAuthRoles() {
	authRoles = config.GetStringSlice("botAuthRoles")
}

// InstallCommands : Installs the commands
func InstallCommands(instance *bot.Instance) {

	// TODO:
	//		- Command to allow Midnight Crew/Specific Role to create a voice channel
	//			- If the VC is empty for more than 5 minutes (no one joins or everyone leaves)
	//				* the bot will destroy the VC
	//
	//			- Birthdays
	//				- Users can assign or update their birthdate  	: !mh birthday set "mm/dd"
	//				- Users can unassign their birthday 			: !mh birthday reset
	//				- Bot will check for the week if a birthday is occurring in the set week
	//					* Add special event for user's birthday to append to schedule

	refreshAuthRoles()
	BotInstance = instance
	session = instance.ClientSession
	prefix = config.GetString("prefix")
	router := exrouter.New()

	// router command template
	// router.On("commandName", func(ctx *exrouter.Context) {
	// 	// Command code
	// 	ctx.Reply("Reply text here!")
	// }).Desc("Descriptive text")

	router.On("tss", func(ctx *exrouter.Context) {
		ctx.Reply("Okay, sending a targeted screenshot~ Standby<3")
		ParseTemplate(ctx.Msg.GuildID)
		go takeAndSendTargeted(ctx.Msg.ChannelID, ctx.Msg.GuildID, BotInstance)
	})

	router.Group(func(r *exrouter.Route) {
		r.Cat("events")

		r.On("events", nil).
			On("add", func(ctx *exrouter.Context) {
				if !AuthEventRunner(ctx) {
					return
				}
				if addEvent(ctx) {
					go parseAndSendSched(ctx)
				}
			})
		r.On("events", nil).
			On("remove", func(ctx *exrouter.Context) {
				if !AuthEventRunner(ctx) {
					return
				}
				if removeEvent(ctx) {
					go parseAndSendSched(ctx)
				}
			})
		r.On("events", nil).
			On("stats", func(ctx *exrouter.Context) {
				go postEventStats(ctx)
			}).Alias("next")
		r.On("events", nil).On("refresh", func(ctx *exrouter.Context) {
			if !AuthEventRunner(ctx) {
				return
			}
			go parseAndSendSched(ctx)
		})
		r.On("events", nil).On("clear", func(ctx *exrouter.Context) {
			if !AuthEventRunner(ctx) {
				return
			}
			err := BotInstance.EventDao.DeleteAllEventsForServer(ctx.Msg.GuildID)
			if err != nil {
				ctx.Reply("There was an error deleting all events ;-;")
				return
			}
			ctx.Reply("Ok! All events for this server have been cleared")
		})
	})

	router.On("servertime", func(ctx *exrouter.Context) {
		if !AuthEventRunner(ctx) {
			return
		}
		ctx.Reply("According to my watch, it is " + time.Now().In(util.ServerLoc).Format("Mon Jan 2 15:04:05 -0700 MST 2006") + " <3")
	})

	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		defer func() {
			if err := recover(); err != nil {
				// if we're in here, we had a panic and have caught it
				s.ChannelMessageSend(m.ChannelID, "Sorry, something when wrong running your command. "+
					"Please check your command usage or try again later.")
				fmt.Printf("Panic deferred in command [%s]: %s\n", m.Content, err)
			}
		}()
		router.FindAndExecute(session, prefix, session.State.User.ID, m.Message)
	})

	log.Info("Commands installed.")
}
