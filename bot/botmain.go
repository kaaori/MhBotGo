package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/dao"
	"github.com/spf13/viper"
)

// Instance : The current instance of the bot and its session
type Instance struct {
	ClientSession       *discordgo.Session
	ServerDao           dao.DiscordServerDao
	EventDao            dao.EventDao
	AnnouncementChannel string
	ScheduleChannel     string
	HasClearedSchedule  bool
	EventRunnerRoleName string
	CurrentFactTitle    string
	CurrentFact         string
}

// InitBot : Initialise a bot instance
func InitBot(token string, dbLocation string) *Instance {
	inst := new(Instance)

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Error("Error creating session\n", err)
		return nil
	}

	inst.ClientSession = discord
	inst.ServerDao = dao.DiscordServerDao{Session: inst.ClientSession}
	inst.EventDao = dao.EventDao{Session: inst.ClientSession}
	inst.HasClearedSchedule = false
	inst.EventRunnerRoleName = viper.GetString("EventRunnerRole")

	return inst
}
