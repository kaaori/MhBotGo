package scheduler

import (
	"fmt"
	"math"
	"strconv"
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

	gocron.Every(1).Monday().At("3:45").Do(ClearSchedules, inst)
	gocron.Every(1).Day().At("00:01").Do(UpdateFact, inst)
	gocron.Start()

	var i int64
	for t := range time.NewTicker(10 * time.Second).C {
		// log.Info("Updating schedule")
		defer func() {
			if err := recover(); err != nil {
				// if we're in here, we had a panic and have caught it
				fmt.Printf("Panic deferred in scheduler: %s\n", err)
			}
		}()
		if i%6 == 0 {

		}
		checkEvents(t, inst)
		i++
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

func UpdateFact(inst *bot.Instance) {
	inst.CurrentFact = commands.GetNewFact()
}

// Runs every 10 seconds
func checkEvents(t time.Time, inst *bot.Instance) {
	for _, g := range inst.ClientSession.State.Guilds {

		schedChannel := commands.FindSchedChannel(g, inst)
		if schedChannel == nil {
			log.Error("Couldn't find schedule channel")
			continue
		}
		schedMsg, _ := commands.GetSchedMessage(schedChannel.ID, inst)
		if schedMsg == nil {
			commands.ParseTemplate(g.ID)
			commands.SendSchedule(schedChannel.ID, inst)
		}
		weekTime := util.GetCurrentWeekFromMondayAsTime()
		evts, err := inst.EventDao.GetAllEventsForServerForWeek(g.ID, weekTime)
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
				evt.LastAnnouncementTimestamp < 0 && timeSinceEvt < 0 {
				announcement = "**___" + evt.EventName + "___**" + " in *" + strconv.Itoa(int(math.Ceil(time.Until(evt.StartTime).Minutes()))) + " minutes!*"
				body = evt.ToAnnounceString()
			} else if (timeSinceEvt.Nanoseconds() >= 0 || timeTilEvt < 0) &&
				(evt.StartTime.After(evt.LastAnnouncementTime) || evt.LastAnnouncementTimestamp < 0) && timeSinceEvt.Hours() <= 2 {
				bot.EventRunning = true
				announcement = "**___" + evt.EventName + "___**" + " **has started!**"
				body = evt.ToStartingString()
				go inst.SetBotGame(inst.ClientSession, "Party Time!")
				commands.ParseTemplate(g.ID)
				commands.SendSchedule(schedChannel.ID, inst)
				go baitAndSwitchGame(inst)
			} else {
				continue
			}

			evt.LastAnnouncementTimestamp = time.Now().Unix()
			inst.EventDao.UpdateEvent(evt)
			inst.ClientSession.ChannelMessageSendEmbed(announcementChannel.ID, commands.GetAnnounceEmbedFromEvent(evt, body, announcement))
			log.Trace("Updated event " + evt.EventName)
		}
	}
}

// Wait an hour then change game back
func baitAndSwitchGame(inst *bot.Instance) {
	time.Sleep(1 * time.Hour)
	inst.SetBotGame(inst.ClientSession, viper.GetString("game"))
	bot.EventRunning = false
}
