package bot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/domain"
	"github.com/kaaori/MhBotGo/util"
)

var (
	EventRunning = false
)

// SetBotGame : Sets the bot's status to playing a game when an event is passed, or just a game if not
func (inst *Instance) SetBotGame(s *discordgo.Session, game string) {
	if err := s.UpdateStatus(0, game); err != nil {
		log.Error("Error setting game", err)
	}
}

// CycleEventStatsAsStatus : Cycle through global stats info once
func CycleEventStatsAsStatus(inst *Instance) {
	for i := 0; i < 3; i++ {
		defer func() {
			if err := recover(); err != nil {
				// if we're in here, we had a panic and have caught it
				fmt.Printf("Panic deferred in scheduler: %s\n", err)
			}
		}()

		// If the event is over or still running
		if EventRunning {
			break
		}

		switch i % 3 {
		case 0:
			inst.SetBotGame(inst.ClientSession, strconv.Itoa(inst.EventDao.GetAllEventCounts())+" events served so far!")
			break
		case 1:
			inst.SetBotGame(inst.ClientSession, strconv.Itoa(inst.EventDao.GetEventsCountForWeek(util.GetCurrentWeekFromMondayAsTime()))+" events this week!")
			break
		case 2:
			inst.SetBotGame(inst.ClientSession, "<3 you all~")
			break
		}
		time.Sleep(6 * time.Second)
	}
}

// CycleEventParamsAsStatus : Cycle through the paramters of the passed event until the event has passed
func CycleEventParamsAsStatus(evt *domain.Event, inst *Instance) {
	i := 0

	for range time.NewTicker(4 * time.Second).C {
		// log.Info("Updating schedule")
		defer func() {
			if err := recover(); err != nil {
				// if we're in here, we had a panic and have caught it
				fmt.Printf("Panic deferred in scheduler: %s\n", err)
			}
		}()

		// If the event is over
		if time.Since(evt.StartTime).Minutes() >= 30 {
			EventRunning = false
			inst.SetBotGame(inst.ClientSession, "<3 event has ended~")
			break
		}

		switch i % 2 {
		case 0:
			inst.SetBotGame(inst.ClientSession, evt.EventName)
			break
		case 1:
			inst.SetBotGame(inst.ClientSession, evt.EventLocation+" with "+evt.HostName)
			break
		}
		i++
	}
}
