// Modified template from https://github.com/2Bot/2Bot-Discord-Bot
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/spf13/viper"

	"github.com/bwmarrin/discordgo"

	_ "github.com/mattn/go-sqlite3"
)

var (
	Token        string
	log          = newLog()
	emojiRegex   = regexp.MustCompile("<(a)?:.*?:(.*?)>")
	userIDRegex  = regexp.MustCompile("<@!?([0-9]+)>")
	channelRegex = regexp.MustCompile("<#([0-9]+)>")
	status       = map[discordgo.Status]string{"dnd": "busy", "online": "online", "idle": "idle", "offline": "offline"}
	footer       = new(discordgo.MessageEmbedFooter)
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func initViper(configFilePath string) {
	viper.New()
	viper.SetConfigType("json")
	viper.SetConfigFile(configFilePath)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
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
	initViper("../configs/config.json")
	fmt.Println(viper.GetString("game"))
	log.Info("======================/ MH Bot Starting \\======================")
	// database, _ := sql.Open("sqlite3", "./configs/mh.db")
	// statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS people (id INTEGER PRIMARY KEY, firstname TEXT, lastname TEXT)")
	// statement.Exec()
	// statement, _ = database.Prepare("INSERT INTO people (firstname, lastname) VALUES (?, ?)")
	// statement.Exec("Nic", "Raboy")
	// rows, _ := database.Query("SELECT id, firstname, lastname FROM people")
	// var id int
	// var firstname string
	// var lastname string
	// for rows.Next() {
	// 	rows.Scan(&id, &firstname, &lastname)
	// 	fmt.Println(strconv.Itoa(id) + ": " + firstname + " " + lastname)
	// }

	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Error("Error creating session\n", err)
		return
	}

	discord.AddHandler(messageCreateEvent)
	discord.AddHandler(readyEvent)
	discord.AddHandler(guildJoinEvent)

	err = discord.Open()
	if err != nil {
		log.Error("Error opening connection\n", err)
		return
	}
	// Defer the session cleanup until the application is closed
	defer discord.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}

func setBotGame(s *discordgo.Session) {
	if err := s.UpdateStatus(0, viper.GetString("game")); err != nil {
		log.Error("Update status err:", err)
		return
	}
	log.Info("set initial game to", viper.GetString("game"))
}
