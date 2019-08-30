package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/dao"
)

// Instance : The current instance of the bot and its session
type Instance struct {
	ClientSession *discordgo.Session
	ServerDao     dao.DiscordServerDao
	db            *DB
}

// InitBot : Initialise a bot instance
func InitBot(token string) *Instance {
	inst := new(Instance)

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Error("Error creating session\n", err)
		return nil
	}

	newDb, err := initDB("./data/MHBot.db", "sqlite3")
	if err != nil {
		log.Error("Error loading DB", err)
		return nil
	}

	inst.ClientSession = discord
	inst.db = newDb
	inst.ServerDao = dao.DiscordServerDao{Session: inst.ClientSession}

	return inst
}
