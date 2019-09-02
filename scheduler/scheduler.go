package scheduler

import (
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/jasonlvhit/gocron"
	"github.com/kaaori/MhBotGo/bot"
	"github.com/kaaori/MhBotGo/commands"
	"github.com/kaaori/MhBotGo/util"

	logging "github.com/kaaori/mhbotgo/log"
)

var (
	log = logging.NewLog()
)

// Scheduler : The scheduler process for the bot
type Scheduler struct{}

// Init : Initialises the schedule timer tasks
func Init(inst *bot.Instance) {
	log.Info("Scheduler tasks starting.")

	gocron.Every(1).Sunday().At("3:45").Do(ClearSchedules, inst)
	gocron.Start()

	for t := range time.NewTicker(10 * time.Second).C {
		// log.Info("Updating schedule")
		defer func() {
			if err := recover(); err != nil {
				// if we're in here, we had a panic and have caught it
				fmt.Printf("Panic deferred in scheduler: %s\n", err)
			}
		}()
		checkEvents(t, inst)
	}
	// return &Scheduler{}
}

// ClearSchedules : Deletes any open schedules that may be floating around and refreshes them
// Ran every Monday to reset the current week
func ClearSchedules(inst *bot.Instance) {
	log.Info("Clearing schedules.")
	for _, g := range inst.ClientSession.State.Guilds {
		schedChannel := commands.FindSchedChannel(g, inst)
		if schedChannel == nil {
			log.Error("Couldn't find schedule channel")
			continue
		}

		if schedChannel != nil {
			commands.ParseTemplate(g.ID)
			commands.SendSchedule(schedChannel.ID, inst)
		}
	}
}

func checkEvents(t time.Time, inst *bot.Instance) {
	for _, g := range inst.ClientSession.State.Guilds {

		schedChannel := commands.FindSchedChannel(g, inst)
		if schedChannel == nil {
			log.Error("Couldn't find schedule channel")
			continue
		}
		schedMsg := commands.GetSchedMessage(schedChannel.ID, inst)
		if schedMsg == nil {
			commands.ParseTemplate(g.ID)
			commands.SendSchedule(schedChannel.ID, inst)
		}
		// TODO Write GetAllEventsForServerForWeek(g.ID, t.Now().ISOWeek() or .Weekday()))
		evts, err := inst.EventDao.GetAllEventsForServer(g.ID)
		if err != nil {
			log.Error("Error fetching events for guild "+g.Name, err)
			continue
		}
		for _, evt := range evts {
			timeTilEvt := time.Until(evt.StartTime)
			timeSinceEvt := time.Since(evt.StartTime)
			announcementChannel := commands.FindAnnouncementsChannel(g, inst)

			if announcementChannel == nil {
				log.Error("Could not find announcement channel")
				break
			}
			var announcement string
			body := ""

			// Unannounced = -1
			if timeTilEvt.Minutes() <= 20 &&
				evt.LastAnnouncementTimestamp < 0 {
				announcement = "**___" + evt.EventName + "___**" + " in *20 minutes!*"
				body = evt.ToAnnounceString()
			} else if timeSinceEvt.Seconds() >= 0 &&
				evt.StartTime.After(evt.LastAnnouncementTime) {
				announcement = "**___" + evt.EventName + "___**" + " **has started!**"
				body = evt.ToStartingString()
				util.SetBotGame(inst.ClientSession, "Party Time!", evt)
				commands.ParseTemplate(g.ID)
				commands.SendSchedule(schedChannel.ID, inst)
				go baitAndSwitchGame(inst)
			} else {
				continue
			}

			evt.LastAnnouncementTimestamp = time.Now().Unix()
			inst.EventDao.UpdateEvent(evt)
			inst.ClientSession.ChannelMessageSendEmbed(announcementChannel.ID, commands.GetAnnounceEmbedFromEvent(evt, body, announcement))

		}
	}
}

// Wait an hour then change game back
func baitAndSwitchGame(inst *bot.Instance) {
	time.Sleep(1 * time.Hour)
	util.SetBotGame(inst.ClientSession, viper.GetString("game"), nil)
}
