package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/kaaori/MhBotGo/dao"

	"github.com/kaaori/MhBotGo/scheduler"

	"github.com/fsnotify/fsnotify"
	"github.com/kaaori/MhBotGo/bot"
	"github.com/kaaori/MhBotGo/commands"
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

func initDbForGuild(guild *discordgo.GuildCreate) {
	if rowsAffected := BotInstance.ServerDao.InsertNewServer(guild.ID); rowsAffected < 0 {
		return
	}
	log.Info("DB Initialised for guild " + guild.Name)
}

func main() {
	initViper("./configs/config.json")
	log.Info("======================/ MH Bot Starting \\======================")
	log.Info("TODO: Scan for guild mismatch in DB (added or removed to new guilds etc) ")

	log.Info("Db opened...")
	if _, err := os.Stat(config.GetString("dbLocation")); err != nil {
		log.Error("DB Not found. Creating in " + config.GetString("dbLocation"))
		bot.ReadDML(config.GetString("dbLocation"))
	}

	BotInstance = bot.InitBot(Token, config.GetString("dbLocation"))
	BotInstance.ClientSession.State.MaxMessageCount = 100
	BotInstance.AnnouncementChannel = config.GetString("announcements")
	BotInstance.ScheduleChannel = config.GetString("schedule")
	BotInstance.CurrentFactTitle, BotInstance.CurrentFact = commands.GetNewFact()
	BotInstance.ClientSession.AddHandler(readyEvent)
	BotInstance.ClientSession.AddHandler(guildJoinEvent)

	commands.InstallCommands(BotInstance)

	err := BotInstance.ClientSession.Open()
	if err != nil {
		log.Error("Error opening connection\n", err)
		return
	}
	// Defer the session cleanup until the application is closed
	defer BotInstance.ClientSession.Close()
	defer dao.DB.Close()
	scheduler.Init(BotInstance)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}
