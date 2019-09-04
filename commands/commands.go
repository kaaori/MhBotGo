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

	router.On("testgamestatus", func(ctx *exrouter.Context) {
		evt, _ := BotInstance.EventDao.GetNextEventOrDefault(ctx.Msg.GuildID)
		util.SetBotGame(ctx.Ses, "Party Time!", evt)
	})

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
		r.On("events", nil).
			On("stats", func(ctx *exrouter.Context) {
				postEventStats(ctx)
			}).Alias("next", "last")
	})

	router.On("servertime", func(ctx *exrouter.Context) {
		fmt.Println(time.Now().Location())
		ctx.Reply("According to my watch, it is " + time.Now().In(util.EstLoc).Format("Mon Jan 2 15:04:05 -0700 EST! 2006") + " <3")
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
