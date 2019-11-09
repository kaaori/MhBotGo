package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/dao"
	"github.com/spf13/viper"
)

// Instance : The current instance of the bot and its session
type Instance struct {
	ClientSession *discordgo.Session
	ServerDao     dao.DiscordServerDao
	EventDao      dao.EventDao
	BirthdayDao   dao.BirthdayDao

	AnnouncementChannel       string
	ScheduleChannel           string
	HasClearedSchedule        bool
	EventRunnerRoleName       string
	EventAttendeeRoleName     string
	CurrentFactTitle          string
	CurrentFact               string
	ScheduleMessagesByGuildID map[string]string
}

// InitBot : Initialise a bot instance
func InitBot(token string, dbLocation string) *Instance {
	inst := new(Instance)

	inst.ScheduleMessagesByGuildID = make(map[string]string)

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Error("Error creating session\n", err)
		return nil
	}

	inst.ClientSession = discord
	inst.ServerDao = dao.DiscordServerDao{Session: inst.ClientSession}
	inst.EventDao = dao.EventDao{Session: inst.ClientSession}
	inst.BirthdayDao = dao.BirthdayDao{Session: inst.ClientSession}
	inst.HasClearedSchedule = false
	inst.EventRunnerRoleName = viper.GetString("eventRunnerRole")
	inst.EventAttendeeRoleName = viper.GetString("eventAttendeeRole")

	return inst
}
