package scheduler

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jasonlvhit/gocron"
	"mhbotgo.com/bot"
	"mhbotgo.com/commands"
	"mhbotgo.com/domain"
	"mhbotgo.com/util"
)

var (
	GuildsWithNoEvents = make([]string, 0)
)

// Scheduler : The scheduler process for the bot
type Scheduler struct{}

// Init : Initialises the schedule timer tasks
func Init(inst *bot.Instance) {
	log.Println("Scheduler tasks starting.")

	gocron.Every(1).Monday().At("12:00").Do(WeeklyEvents, inst)

	gocron.Every(1).Day().At("00:30").Do(UpdateFact, inst)
	gocron.Every(1).Day().At("2:00").Do(UpdateSchedule, inst)
	gocron.Every(1).Day().At("10:00").Do(UpdateBirthdays, inst)
	gocron.Start()

	go procEventLoop(inst)
}

func procEventLoop(inst *bot.Instance) {
	var i int64

	for t := range time.NewTicker(10 * time.Second).C {
		defer func() {
			if err := recover(); err != nil {
				// if we're in here, we had a panic and have caught it
				fmt.Printf("Panic deferred in scheduler Event loop: %s\n%s", err, string(debug.Stack()))
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

// UpdateBirthdays : Once a day, go through every stored birthday for a given week and check if it is today
func UpdateBirthdays(inst *bot.Instance) {
	defer func() {
		if err := recover(); err != nil {
			// if we're in here, we had a panic and have caught it
			fmt.Printf("Panic deferred in Update Birthday: %s\n%s", err, string(debug.Stack()))
		}
	}()
	for _, g := range inst.ClientSession.State.Guilds {
		schedChannel := commands.FindSchedChannel(inst, g.ID)
		if schedChannel == nil {
			log.Println("Couldn't find schedule channel")
			continue
		}

		schedMsg, _ := commands.GetSchedMessage(schedChannel.ID, inst)
		if schedMsg == nil {
			commands.ParseTemplate(g.ID)
			commands.SendSchedule(schedChannel.ID, g.ID, inst)
		}

		weekTime := util.GetCurrentWeekFromMondayAsTime()
		birthdays, err := inst.BirthdayDao.GetAllBirthdaysForServerForWeek(g.ID, weekTime, g)
		if err != nil && err.Error() != "Couldn't find birthday" {
			log.Println("Error fetching birthdays for guild "+g.Name, err)
			continue
		}

		// If there are not events anymore (e.g had 1 but was removed)
		// Then we need to repost our schedule to reflect that change
		if len(birthdays) <= 0 && !contains(GuildsWithNoEvents, g.ID) {
			GuildsWithNoEvents = append(GuildsWithNoEvents, g.ID)
			commands.ParseTemplate(g.ID)
			go commands.SendSchedule(schedChannel.ID, g.ID, inst)

			// And go to the next server
			continue
		} else if len(birthdays) > 0 && !contains(GuildsWithNoEvents, g.ID) {
			// Otherwise if there are now events, remove that guild?
			// This is useless I believe
			GuildsWithNoEvents = remove(GuildsWithNoEvents, g.ID)
		}

		birthdaysToPost := make([]*domain.Birthday, 0)

		for _, birthday := range birthdays {
			announcementChannel := commands.FindAnnouncementsChannel(g, inst)

			if announcementChannel == nil {
				log.Println("Could not find announcement channel")
				break
			}

			content := ""

			// Unannounced = -1
			if birthday.IsToday() {
				birthdaysToPost = append(birthdaysToPost, birthday)
			} else {
				continue
			}

			birthdayUser, _ := inst.ClientSession.User(birthday.GuildUserID)
			msg := &discordgo.MessageSend{
				Embed:   commands.GetAnnounceEmbedFromBirthday(birthday, birthdayUser),
				Content: content}
			log.Println("Updated birthday for  " + birthdayUser.Username)

			// See if there's a fact for the user whos birthday it is
			userFact, err := inst.FactDao.GetFactByUser(birthdayUser)
			if userFact != nil && err == nil {
				// Update the fact and refresh the schedule if it exists
				log.Println("Birthday fact has been set, updating schedule")

				inst.CurrentFact = userFact.User.Username + " says: " + userFact.FactContent
				inst.CurrentFactTitle = "Did you know this about " + userFact.User.Username
				commands.SendSchedule(schedChannel.ID, g.ID, inst)
			}
			go inst.ClientSession.ChannelMessageSendComplex(announcementChannel.ID, msg)
		}
	}
}

// UpdateSchedule : Refreshes the schedule for the current day
func UpdateSchedule(inst *bot.Instance) {
	defer func() {
		if err := recover(); err != nil {
			// if we're in here, we had a panic and have caught it
			fmt.Printf("Panic deferred in Updating Scheduler: %s\n%s", err, string(debug.Stack()))
		}
	}()
	for _, g := range inst.ClientSession.State.Guilds {
		schedChannel := commands.FindSchedChannel(inst, g.ID)
		if schedChannel == nil {
			log.Println("Couldn't find schedule channel")
			continue
		}

		commands.SendSchedule(schedChannel.ID, g.ID, inst)
	}
}

func WeeklyEvents(inst *bot.Instance) {
	defer func() {
		if err := recover(); err != nil {
			// if we're in here, we had a panic and have caught it
			fmt.Printf("Panic deferred in Weekly Events post: %s\n%s", err, string(debug.Stack()))
		}
	}()
	ClearSchedulesAndMakeEveryoneMad(inst)
}

// PingEveryoneInScheduleChannel : Ping everyone in schedule channel
func PingEveryoneInScheduleChannel(guildID string, channelID string, s *discordgo.Session) {
	defer func() {
		if err := recover(); err != nil {
			// if we're in here, we had a panic and have caught it
			fmt.Printf("Panic deferred in Ping all for Schedule: %s\n%s", err, string(debug.Stack()))
		}
	}()
	_, err := s.ChannelMessageSend(channelID, "@everyone")
	if err != nil {
		log.Println("Error pinging everyone", err)
		return
	}
}

// ClearSchedulesAndMakeEveryoneMad : Deletes any open schedules that may be floating around and refreshes them
// Ran every Monday to reset the current week at noon
func ClearSchedulesAndMakeEveryoneMad(inst *bot.Instance) {
	defer func() {
		if err := recover(); err != nil {
			// if we're in here, we had a panic and have caught it
			fmt.Printf("Panic deferred in Clearing schedules and angering the masses: %s\n%s", err, string(debug.Stack()))
		}
	}()
	log.Println("Clearing schedules.")
	for _, g := range inst.ClientSession.State.Guilds {

		schedChannel := commands.FindSchedChannel(inst, g.ID)
		if schedChannel == nil {
			log.Println("Couldn't find schedule channel")
			continue
		}

		if schedChannel != nil {
			msgs, err := inst.ClientSession.ChannelMessages(schedChannel.ID, 100, "", "", "")
			if err != nil {
				log.Println("Couldn't find schedule channel messages")
				continue
			}

			// Delete all messages in schedule channel
			var msgIDsToDelete []string
			for _, msg := range msgs {
				msgIDsToDelete = append(msgIDsToDelete, msg.ID)
			}
			if len(msgIDsToDelete) > 0 {
				inst.ClientSession.ChannelMessagesBulkDelete(schedChannel.ID, msgIDsToDelete)
				log.Println("Cleared messages from schedule channel")
			} else {
				log.Println("Could not find any messages")
			}
			commands.ParseTemplate(g.ID)
			PingEveryoneInScheduleChannel(schedChannel.GuildID, schedChannel.ID, inst.ClientSession)
			commands.SendSchedule(schedChannel.ID, g.ID, inst, true)
		}
	}
}

// UpdateFact : Updates the fact of the day from an RSS feed
func UpdateFact(inst *bot.Instance) {
	defer func() {
		if err := recover(); err != nil {
			// if we're in here, we had a panic and have caught it
			fmt.Printf("Panic deferred in Updating fact: %s\n%s", err, string(debug.Stack()))
		}
	}()
	inst.CurrentFactTitle, inst.CurrentFact = commands.GetNewFact(inst, inst.CurrentFact, false)
	for _, g := range inst.ClientSession.State.Guilds {

		schedChannel := commands.FindSchedChannel(inst, g.ID)
		if schedChannel == nil {
			log.Println("Couldn't find schedule channel")
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
			log.Println("Couldn't find schedule channel")
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
			log.Println("Error fetching events for guild "+g.Name, err)
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
				log.Println("Could not find announcement channel")
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
					log.Println("Could not find attendee role")
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
			log.Println("Updated event " + evt.EventName)

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
