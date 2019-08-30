/**************************************************************************
* A majority of this will be refactored and broken up into packages later *
***************************************************************************/

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/Necroforger/dgrouter"
	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/fsnotify/fsnotify"
	"github.com/kaaori/mhbotgo/bot"
	"github.com/kaaori/mhbotgo/domain"
	"github.com/kaaori/mhbotgo/util"
	config "github.com/spf13/viper"

	"github.com/bwmarrin/discordgo"

	logging "github.com/kaaori/mhbotgo/log"
)

var (
	// Token is the Bot token for discord auth
	Token        string
	log          = logging.NewLog()
	emojiRegex   = regexp.MustCompile("<(a)?:.*?:(.*?)>")
	userIDRegex  = regexp.MustCompile("<@!?([0-9]+)>")
	channelRegex = regexp.MustCompile("<#([0-9]+)>")
	status       = map[discordgo.Status]string{"dnd": "busy", "online": "online", "idle": "idle", "offline": "offline"}
	footer       = new(discordgo.MessageEmbedFooter)
	prefix       = "!mh "

	// BotInstance : The bot instance
	BotInstance *bot.Instance
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func initViper(configFilePath string) {
	config.New()
	config.SetConfigType("json")
	config.SetConfigFile(configFilePath)
	config.WatchConfig()

	config.OnConfigChange(func(e fsnotify.Event) {
		log.Info("Config has been updated.")
	})

	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(config.ConfigFileNotFoundError); ok {
			log.Error("Config file not found!\n\t\tEnsure file ./configs/config.json exists", err)
			os.Exit(404)
		} else {
			// Config file was found but another error was produced
			log.Error("Fatal error in loading config:\n\t\t", err)
			return
		}
	}
}

func main() {
	initViper("./configs/config.json")
	log.Info("======================/ MH Bot Starting \\======================")
	log.Info("TODO: Scan for guild mismatch in DB (added or removed to new guilds etc) ")

	if _, err := os.Stat(config.GetString("dbLocation")); err != nil {
		log.Error("DB Not found. Creating in " + config.GetString("dbLocation"))
		bot.ReadDML(config.GetString("dbLocation"))
	}

	BotInstance = bot.InitBot(Token, config.GetString("dbLocation"))

	BotInstance.ClientSession.AddHandler(readyEvent)
	BotInstance.ClientSession.AddHandler(guildJoinEvent)

	installCommands(BotInstance.ClientSession)

	err := BotInstance.ClientSession.Open()
	if err != nil {
		log.Error("Error opening connection\n", err)
		return
	}
	// Defer the session cleanup until the application is closed
	defer BotInstance.ClientSession.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}

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
			validateNewEventArgs(ctx.Args, ctx, event)
			BotInstance.EventDao.InsertEvent(event)
			break
		case "edit":
			break
		case "remove":
			break
		default:
			ctx.Reply(config.GetString("badSyntaxError"))
			break

		}
		// for _, _ := range ctx.Args {

		// }
		ctx.Reply("Args printed")
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
		outputEmbed := util.GetEmbed(fields, "Test Command!", "Footer!")
		BotInstance.ClientSession.ChannelMessageSendEmbed(ctx.Msg.ChannelID, outputEmbed)
		// ctx.ReplyEmbed(outputEmbed)
		// ctx.Reply("Currently serving " + strconv.Itoa(userCount) + " users")
	}).Desc("Test command")

	session.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		router.FindAndExecute(session, prefix, session.State.User.ID, m.Message)
	})

	log.Info("Commands installed.")
}

func validateNewEventArgs(args []string, ctx *exrouter.Context, event *domain.Event) {
	// TODO: Add descriptive examples to errors \/
	// TODO: Validate el as date
	if dateString := ctx.Args.Get(2); "" != dateString {
		event.StartTimestamp = time.Now().Unix() + 10000
	} else {
		ctx.Reply("Please check your date format and try again!")
		return
	}
	if hostName := ctx.Args.Get(3); "" != hostName {
		event.HostName = hostName
	} else {
		ctx.Reply("Please ensure you have included a host to your event")
		return
	}
	if name := ctx.Args.Get(4); "" != name {
		event.Name = name
	} else {
		ctx.Reply("Please ensure you have given the event a name")
		return
	}
	if location := ctx.Args.Get(5); "" != location {
		event.EventLocation = location
	} else {
		ctx.Reply("Please ensure you have given the event a location")
		return
	}
}
