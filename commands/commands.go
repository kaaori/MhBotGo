package commands

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/bot"
	"github.com/kaaori/MhBotGo/chrome"
	"github.com/kaaori/MhBotGo/util"
	logging "github.com/kaaori/mhbotgo/log"
	config "github.com/spf13/viper"
)

var (
	prefix    string
	log       = logging.NewLog()
	session   *discordgo.Session
	authRoles []string

	defaultScreenshotW = int64(880)
	defaultScreenshotH = int64(1000)

	// BotInstance : The instance of the bot containing the discord session and all relevant DAOs
	BotInstance *bot.Instance
)

func refreshAuthRoles() {
	authRoles = config.GetStringSlice("botAuthRoles")
}

// InstallCommands : Installs the commands
func InstallCommands(instance *bot.Instance) {

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

	router.On("testgamestatus", func(ctx *exrouter.Context) {
		evt, _ := BotInstance.EventDao.GetNextEventOrDefault(ctx.Msg.GuildID)
		go bot.CycleEventParamsAsStatus(evt, BotInstance)
		ctx.Reply("Okay~")
	})

	router.On("factref", func(ctx *exrouter.Context) {
		BotInstance.CurrentFactTitle, BotInstance.CurrentFact = GetNewFact()
	})

	router.On("testss", func(ctx *exrouter.Context) {
		userW, _ := strconv.ParseInt(ctx.Args.Get(1), 10, 64)
		userH, _ := strconv.ParseInt(ctx.Args.Get(2), 10, 64)
		ctx.Reply("Okay! Sending a screenshot with Width: " + strconv.Itoa(int(userW)) + " and Height: " + strconv.Itoa(int(userH)) + "... Standby<3")
		ParseTemplate(ctx.Msg.GuildID)

		chrome.TakeScreenshot(userW, userH)
		f, err := os.Open("schedule.png")
		if err != nil {
			log.Error("Error getting schedule image", err)
			return
		}
		defer f.Close()

		ms := &discordgo.MessageSend{
			// Embed: &discordgo.MessageEmbed{
			// 	Title: "Click the schedule below to see more info!",
			// 	Color: 0x9400d3,
			// 	Image: &discordgo.MessageEmbedImage{
			// 		URL: "attachment://" + "schedule.png",
			// 	},
			// },
			Files: []*discordgo.File{
				&discordgo.File{
					Name:   "schedule.png",
					Reader: f,
				},
			},
		}

		BotInstance.ClientSession.ChannelMessageSendComplex(ctx.Msg.ChannelID, ms)
	})

	router.On("gctemplate", func(ctx *exrouter.Context) {
		ParseTemplate(ctx.Msg.GuildID)

		chrome.TakeScreenshot(defaultScreenshotW, defaultScreenshotH)

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
				postEventStats(ctx)
			})
		r.On("events", nil).On("refresh", func(ctx *exrouter.Context) {
			if !AuthEventRunner(ctx) {
				return
			}
			go parseAndSendSched(ctx)
		})
	})

	router.On("servertime", func(ctx *exrouter.Context) {
		if !AuthEventRunner(ctx) {
			return
		}
		fmt.Println(time.Now().Location())
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
