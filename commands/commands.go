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
	defaultScreenshotH = int64(1080)

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

	router.On("refresh", nil).On("fact", func(ctx *exrouter.Context) {
		if !AuthEventRunner(ctx) {
			return
		}
		BotInstance.CurrentFactTitle, BotInstance.CurrentFact = GetNewFact(BotInstance.CurrentFact, false)
		ctx.Reply("Ok, fact has been updated if a newer one is available <3\n`" + BotInstance.CurrentFact + "`")
	})

	router.On("refresh", nil).On("userfact", func(ctx *exrouter.Context) {
		if !AuthEventRunner(ctx) {
			return
		}
		BotInstance.CurrentFactTitle, BotInstance.CurrentFact = GetNewFact(BotInstance.CurrentFact, true)
		ctx.Reply("Ok, fact has been updated if a newer one is available <3\n`" + BotInstance.CurrentFact + "`")
	})

	router.On("refresh", nil).On("help", func(ctx *exrouter.Context) {
		if !AuthEventRunner(ctx) {
			return
		}
		deleteHelpImage()
		ctx.Reply("Ok! The help command image will be updated the next time the command is used<3 | Server-time: " + time.Now().In(util.ServerLoc).Format("Mon Jan 2 15:04:05 -0700 MST 2006"))
	})

	router.On("fact", func(ctx *exrouter.Context) {
		fact := insertFact(ctx)
		if fact != nil {
			ctx.Reply("Ok, your fact has been inserted! <3\n`" + fact.FactContent + "`")
		}
	})

	router.On("fact", nil).On("reset", func(ctx *exrouter.Context) {
		if deleteFact(ctx) {
			ctx.Reply("Ok, your set fact has been deleted!")
		} else {
			ctx.Reply("Hmm, I couldn't find any facts by you.")
		}
	})

	router.On("help", func(ctx *exrouter.Context) {
		go sendHelpMessage(ctx)
	})

	router.Group(func(r *exrouter.Route) {
		r.Cat("birthdays")

		r.On("birthday", nil).
			On("set", func(ctx *exrouter.Context) {
				if setBirthday(ctx) {
					go parseAndSendSched(ctx)
				}
			})
		r.On("birthday", nil).
			On("reset", func(ctx *exrouter.Context) {
				if resetBirthday(ctx) {
					go parseAndSendSched(ctx)
				}
			})
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
			parseAndSendSched(ctx)
			ctx.Reply("Ok, I have refreshed the current schedule<3 | Server-time: " + time.Now().In(util.ServerLoc).Format("Mon Jan 2 15:04:05 -0700 MST 2006"))
		})
		// r.On("events", nil).On("clear", func(ctx *exrouter.Context) {
		// 	if !AuthEventRunner(ctx) {
		// 		return
		// 	}
		// 	err := BotInstance.EventDao.DeleteAllEventsForServer(ctx.Msg.GuildID)
		// 	if err != nil {
		// 		ctx.Reply("There was an error deleting all events ;-;")
		// 		return
		// 	}
		// 	ctx.Reply("Ok! All events for this server have been cleared")
		// })
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
		// We dont want our bot to respond to itself or other bots
		if !m.Message.Author.Bot {
			router.FindAndExecute(session, prefix, session.State.User.ID, m.Message)
		}
	})

	log.Info("Commands installed.")
}
