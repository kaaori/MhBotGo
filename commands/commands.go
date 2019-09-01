package commands

import (
	"fmt"
	"os"
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

	router.On("gctemplate", func(ctx *exrouter.Context) {
		ParseTemplate(ctx.Msg.GuildID)

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

	router.Group(func(r *exrouter.Route) {
		r.Use(Auth)

		r.On("events", nil).
			On("add", Auth(func(ctx *exrouter.Context) {
				if addEvent(ctx) {
					parseAndSendSched(ctx)
				}
			}))
		r.On("events", nil).
			On("remove", func(ctx *exrouter.Context) {
				if removeEvent(ctx) {
					parseAndSendSched(ctx)
				}
			})
	})

	router.On("servertime", func(ctx *exrouter.Context) {
		fmt.Println(time.Now().Location())
		ctx.Reply("According to my watch, it is " + time.Now().In(util.EstLoc).Format("Mon Jan 2 15:04:05 -0700 EST! 2006") + " <3")
	})

	session.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		router.FindAndExecute(session, prefix, session.State.User.ID, m.Message)
	})

	log.Info("Commands installed.")
}

func test(ctx *exrouter.Context) {
	removeEvent(ctx)
	parseAndSendSched(ctx)
}
