package scheduler

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jasonlvhit/gocron"
	"github.com/kaaori/MhBotGo/bot"
	"github.com/kaaori/MhBotGo/commands"
	"github.com/kaaori/MhBotGo/util"

	logging "github.com/kaaori/mhbotgo/log"
)

var (
	log                = logging.NewLog()
	GuildsWithNoEvents = make([]string, 0)
)

// Scheduler : The scheduler process for the bot
type Scheduler struct{}

// Init : Initialises the schedule timer tasks
func Init(inst *bot.Instance) {
	log.Info("Scheduler tasks starting.")

	// Check birthdays here as well
	gocron.Every(1).Monday().At("12:00").Do(WeeklyEvents, inst)

	gocron.Every(1).Day().At("00:30").Do(UpdateFact, inst)
	gocron.Every(1).Day().At("2:00").Do(UpdateSchedule, inst)
	gocron.Start()

	go procEventLoop(inst)
}

func procEventLoop(inst *bot.Instance) {
	var i int64

	for t := range time.NewTicker(10 * time.Second).C {
		defer func() {
			if err := recover(); err != nil {
				// if we're in here, we had a panic and have caught it
				fmt.Printf("Panic deferred in scheduler: %s\n", err)
			}
		}()

		// Every other tick cycle event stats
		if i%2 == 0 {
			go bot.CycleEventStatsAsStatus(inst)

		}
		checkEvents(t, inst)
		i++
	}
}

// UpdateSchedule : Refreshes the schedule for the current day
func UpdateSchedule(inst *bot.Instance) {
	for _, g := range inst.ClientSession.State.Guilds {
		schedChannel := commands.FindSchedChannel(inst, g.ID)
		if schedChannel == nil {
			log.Error("Couldn't find schedule channel")
			continue
		}

		commands.SendSchedule(schedChannel.ID, g.ID, inst)
	}
}

func WeeklyEvents(inst *bot.Instance) {
	ClearSchedulesAndMakeEveryoneMad(inst)
}

// PingEveryoneInScheduleChannel : Ping everyone in schedule channel
func PingEveryoneInScheduleChannel(guildID string, channelID string, s *discordgo.Session) {
	_, err := s.ChannelMessageSend(channelID, "@everyone")
	if err != nil {
		log.Error("Error pinging everyone", err)
		return
	}
}

// ClearSchedulesAndMakeEveryoneMad : Deletes any open schedules that may be floating around and refreshes them
// Ran every Monday to reset the current week at noon
func ClearSchedulesAndMakeEveryoneMad(inst *bot.Instance) {
	log.Info("Clearing schedules.")
	for _, g := range inst.ClientSession.State.Guilds {

		schedChannel := commands.FindSchedChannel(inst, g.ID)
		if schedChannel == nil {
			log.Error("Couldn't find schedule channel")
			continue
		}

		if schedChannel != nil {
			msgs, err := inst.ClientSession.ChannelMessages(schedChannel.ID, 100, "", "", "")
			if err != nil {
				log.Error("Couldn't find schedule channel messages")
				continue
			}

			// Delete all messages in schedule channel
			var msgIDsToDelete []string
			for _, msg := range msgs {
				msgIDsToDelete = append(msgIDsToDelete, msg.ID)
			}
			if len(msgIDsToDelete) > 0 {
				inst.ClientSession.ChannelMessagesBulkDelete(schedChannel.ID, msgIDsToDelete)
				log.Trace("Cleared messages from schedule channel")
			} else {
				log.Trace("Could not find any messages")
			}
			commands.ParseTemplate(g.ID)
			PingEveryoneInScheduleChannel(schedChannel.GuildID, schedChannel.ID, inst.ClientSession)
			commands.SendSchedule(schedChannel.ID, g.ID, inst, true)
		}
	}
}

// UpdateFact : Updates the fact of the day from an RSS feed
func UpdateFact(inst *bot.Instance) {
	inst.CurrentFactTitle, inst.CurrentFact = commands.GetNewFact()
	for _, g := range inst.ClientSession.State.Guilds {

		schedChannel := commands.FindSchedChannel(inst, g.ID)
		if schedChannel == nil {
			log.Error("Couldn't find schedule channel")
			continue
		}

		if schedChannel != nil {
			commands.ParseTemplate(g.ID)
			commands.SendSchedule(schedChannel.ID, g.ID, inst)
		}
	}
}

// Runs every 10 seconds
func checkEvents(t time.Time, inst *bot.Instance) {
	for _, g := range inst.ClientSession.State.Guilds {
		schedChannel := commands.FindSchedChannel(inst, g.ID)
		if schedChannel == nil {
			log.Error("Couldn't find schedule channel")
			continue
		}

		schedMsg, _ := commands.GetSchedMessage(schedChannel.ID, inst)
		if schedMsg == nil {
			commands.ParseTemplate(g.ID)
			commands.SendSchedule(schedChannel.ID, g.ID, inst)
		}

		weekTime := util.GetCurrentWeekFromMondayAsTime()
		evts, err := inst.EventDao.GetAllEventsForServerForWeek(g.ID, weekTime, g)
		if err != nil && err.Error() != "Couldn't find event" {
			log.Error("Error fetching events for guild "+g.Name, err)
			continue
		}

		if len(evts) <= 0 && !contains(GuildsWithNoEvents, g.ID) {
			GuildsWithNoEvents = append(GuildsWithNoEvents, g.ID)
			commands.ParseTemplate(g.ID)
			go commands.SendSchedule(schedChannel.ID, g.ID, inst)
			continue
		} else if len(evts) > 0 && !contains(GuildsWithNoEvents, g.ID) {
			GuildsWithNoEvents = remove(GuildsWithNoEvents, g.ID)
		}

		for _, evt := range evts {
			timeTilEvt := time.Until(evt.StartTime)
			timeSinceEvt := time.Since(evt.StartTime)
			announcementChannel := commands.FindAnnouncementsChannel(g, inst)

			if announcementChannel == nil {
				log.Error("Could not find announcement channel")
				break
			}
			announcement := ""
			body := ""
			content := ""

			// Unannounced = -1
			if timeTilEvt.Minutes() <= 20 &&
				evt.LastAnnouncementTimestamp < 0 &&
				timeSinceEvt < 0 {

				announcement = "**___" + evt.EventName + "___**" + " in *" + util.GetRoundedMinutesTilEvent(evt.StartTime) + " minutes!*"
				body = evt.ToAnnounceString()
				evt.LastAnnouncementTimestamp = 1

			} else if (timeSinceEvt.Nanoseconds() >= 0 || timeTilEvt < 0) &&
				(evt.LastAnnouncementTimestamp <= 1) &&
				timeSinceEvt.Hours() <= 2 {

				bot.EventRunning = true
				announcement = "**___" + evt.EventName + "___**" + " **has started!**"
				body = evt.ToStartingString()
				commands.ParseTemplate(g.ID)
				go commands.SendSchedule(schedChannel.ID, evt.ServerID, inst)
				go bot.CycleEventParamsAsStatus(evt, inst)
				role, err := commands.FindRoleByName(g.ID, inst.EventAttendeeRoleName)
				if err != nil {
					log.Error("Could not find attendee role")
				} else {
					content = role.Mention()
				}
				evt.LastAnnouncementTimestamp = 2
			} else {
				continue
			}

			inst.EventDao.UpdateEvent(evt)
			msg := &discordgo.MessageSend{
				Embed:   commands.GetAnnounceEmbedFromEvent(evt, body, announcement),
				Content: content}
			log.Trace("Updated event " + evt.EventName)

			go inst.ClientSession.ChannelMessageSendComplex(announcementChannel.ID, msg)
		}
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
