package commands

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/forestgiant/sliceutil"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/araddon/dateparse"
	"github.com/bwmarrin/discordgo"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/mmcdole/gofeed"
	config "github.com/spf13/viper"
	"mhbotgo.com/bot"
	"mhbotgo.com/chrome"
	"mhbotgo.com/domain"
	"mhbotgo.com/profiler"
	util "mhbotgo.com/util"
)

var (
	// ScheduleFileName : Name of schedule image
	ScheduleFileName = "schedule"

	// TodayFileName : The day's event image
	TodayFileName = "today"

	// BirthdayCooldownUsersMap : The map of users on setBirthday command cooldown (15 mins)
	BirthdayCooldownUsersMap = make(map[string]time.Time)
)

// MemberHasPermission : Checks if the member has a given perm
func MemberHasPermission(s *discordgo.Session, guildID string, userID string, permission int64) bool {
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		if member, err = s.GuildMember(guildID, userID); err != nil {
			return false
		}
	}

	// Iterate through the role IDs stored in member.Roles
	// to check permissions
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			return false
		}
		if role.Permissions&permission != 0 || role.Name == "BotAdmin" {
			return true
		}
	}

	return false
}

// ParseTemplate : Parses a given guild's events
// Returns true if the schedule needs a refresh
func ParseTemplate(guildID string) {
	profiler.Start()
	tmpl, err := template.ParseFiles("./web/schedule-template.html")
	if err != nil {
		panic(err)
	}

	tmplToday, err := template.ParseFiles("./web/today-template.html")
	if err != nil {
		panic(err)
	}

	profiler.StopAndPrintSeconds("Template parsing")

	profiler.Start()
	g, _ := BotInstance.ClientSession.Guild(guildID)
	fSched, err := os.Create("./web/schedule-parsed" + guildID + ".html")
	defer fSched.Close()

	if err != nil {
		log.Println("create file: ", err)
		return
	}

	fToday, err := os.Create("./web/today-parsed" + guildID + ".html")
	defer fToday.Close()

	if err != nil {
		log.Println("create file: ", err)
		return
	}
	profiler.StopAndPrintSeconds("File creation")
	weekTime := util.GetCurrentWeekFromMondayAsTime()

	events, err := BotInstance.EventDao.GetAllEventsForServerForWeek(guildID, weekTime, g)
	if err != nil {
		log.Println("", err)
		return
	}

	birthdays, err := BotInstance.BirthdayDao.GetAllBirthdaysForServerForWeek(guildID, weekTime, g)
	if err != nil {
		log.Println("", err)
		return
	}

	// if len(evts) <= 0 && !contains(guildsWithNoEvents, g.ID) {
	// 	guildsWithNoEvents = append(guildsWithNoEvents, g.ID)
	// 	ParseTemplate(g.ID)
	// 	go SendSchedule(schedChannel.ID, g.ID, inst)
	// 	continue
	// } else if len(evts) > 0 && !contains(guildsWithNoEvents, g.ID) {
	// 	guildsWithNoEvents = remove(guildsWithNoEvents, g.ID)
	// }

	profiler.StopAndPrintSeconds("Getting all events")
	// allEvents := make([][]*domain.EventView, 0)

	monEvts := make([]*domain.EventView, 0)
	tuesEvts := make([]*domain.EventView, 0)
	wedEvts := make([]*domain.EventView, 0)
	thursEvts := make([]*domain.EventView, 0)
	friEvts := make([]*domain.EventView, 0)
	satEvts := make([]*domain.EventView, 0)
	sunEvts := make([]*domain.EventView, 0)
	curDayEvts := make([]*domain.EventView, 0)

	monBirthdays := make([]*domain.BirthdayView, 0)
	tuesBirthdays := make([]*domain.BirthdayView, 0)
	wedBirthdays := make([]*domain.BirthdayView, 0)
	thursBirthdays := make([]*domain.BirthdayView, 0)
	friBirthdays := make([]*domain.BirthdayView, 0)
	satBirthdays := make([]*domain.BirthdayView, 0)
	sunBirthdays := make([]*domain.BirthdayView, 0)
	curDayBirthdays := make([]*domain.BirthdayView, 0)

	// TODO: to array accessed by (0-6)/(1-7) based on current day int
	// Fixes this ugly shit lol
	for _, el := range events {
		dayOfWeek := el.StartTime.Weekday()
		switch dayOfWeek {
		case time.Monday:
			monEvts = appendEventToList(monEvts, el)
			break
		case time.Tuesday:
			tuesEvts = appendEventToList(tuesEvts, el)
			break
		case time.Wednesday:
			wedEvts = appendEventToList(wedEvts, el)
			break
		case time.Thursday:
			thursEvts = appendEventToList(thursEvts, el)
			break
		case time.Friday:
			friEvts = appendEventToList(friEvts, el)
			break
		case time.Saturday:
			satEvts = appendEventToList(satEvts, el)
			break
		case time.Sunday:
			sunEvts = appendEventToList(sunEvts, el)
			break
		}
	}

	for _, birthday := range birthdays {
		dayOfWeek := birthday.GetTimeFromBirthday().Weekday()
		switch dayOfWeek {
		case time.Monday:
			monBirthdays = appendBirthdayToList(monBirthdays, birthday)
			break
		case time.Tuesday:
			tuesBirthdays = appendBirthdayToList(tuesBirthdays, birthday)
			break
		case time.Wednesday:
			wedBirthdays = appendBirthdayToList(wedBirthdays, birthday)
			break
		case time.Thursday:
			thursBirthdays = appendBirthdayToList(thursBirthdays, birthday)
			break
		case time.Friday:
			friBirthdays = appendBirthdayToList(friBirthdays, birthday)
			break
		case time.Saturday:
			satBirthdays = appendBirthdayToList(satBirthdays, birthday)
			break
		case time.Sunday:
			sunBirthdays = appendBirthdayToList(sunBirthdays, birthday)
			break
		}
	}

	switch time.Now().Weekday() {
	case time.Monday:
		curDayEvts = monEvts
		curDayBirthdays = monBirthdays
		break
	case time.Tuesday:
		curDayEvts = tuesEvts
		curDayBirthdays = tuesBirthdays
		break
	case time.Wednesday:
		curDayEvts = wedEvts
		curDayBirthdays = wedBirthdays
		break
	case time.Thursday:
		curDayEvts = thursEvts
		curDayBirthdays = thursBirthdays
		break
	case time.Friday:
		curDayEvts = friEvts
		curDayBirthdays = friBirthdays
		break
	case time.Saturday:
		curDayEvts = satEvts
		curDayBirthdays = satBirthdays
		break
	case time.Sunday:
		curDayEvts = sunEvts
		curDayBirthdays = sunBirthdays
		break
	}

	// _, isoWeek := t.ISOWeek()
	firstDayOfWeek := util.GetCurrentWeekFromMondayAsTime()

	// GOD ITS SO UGLY PLEASE MAKE IT STOP
	weekEvts := make([][]*domain.EventView, 0)
	weekEvts = append(weekEvts, sortEventList(monEvts))
	weekEvts = append(weekEvts, sortEventList(tuesEvts))
	weekEvts = append(weekEvts, sortEventList(wedEvts))
	weekEvts = append(weekEvts, sortEventList(thursEvts))
	weekEvts = append(weekEvts, sortEventList(friEvts))
	weekEvts = append(weekEvts, sortEventList(satEvts))
	weekEvts = append(weekEvts, sortEventList(sunEvts))

	birthdayEvts := make([][]*domain.BirthdayView, 0)
	birthdayEvts = append(birthdayEvts, monBirthdays)
	birthdayEvts = append(birthdayEvts, tuesBirthdays)
	birthdayEvts = append(birthdayEvts, wedBirthdays)
	birthdayEvts = append(birthdayEvts, thursBirthdays)
	birthdayEvts = append(birthdayEvts, friBirthdays)
	birthdayEvts = append(birthdayEvts, satBirthdays)
	birthdayEvts = append(birthdayEvts, sunBirthdays)

	days := buildDayViews(firstDayOfWeek, weekEvts, birthdayEvts)

	curDayEvts = sortEventList(curDayEvts)

	profiler.StopAndPrintSeconds("Parsing events to views")
	data := domain.ScheduleView{
		ServerName:        g.Name,
		CurrentWeekString: string(firstDayOfWeek.Format("January 2, 2006") + " ── " + firstDayOfWeek.AddDate(0, 0, 6).Format("January 2, 2006")),
		Tz:                "<strong>Eastern Standard Time</strong>",
		CurrentDayString:  time.Now().Format("Monday January 2, 2006"),
		CurrentDay:        curDayEvts,
		CurrentBirthdays:  curDayBirthdays,
		Week: &domain.WeekView{
			Days: days},

		FactTitle: BotInstance.CurrentFactTitle,
		Fact:      BotInstance.CurrentFact}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		err = tmpl.Execute(fSched, data)
		if err != nil {
			log.Println("Error executing template", err)
		}
	}()

	go func() {
		defer wg.Done()
		err = tmplToday.Execute(fToday, data)
		if err != nil {
			log.Println("Error executing template", err)
		}
	}()

	// Wait for both templates to be done processing
	wg.Wait()
	// fSched.Close()
	// fToday.Close()
	profiler.StopAndPrintSeconds("Executing templates")
}

func buildDayViews(firstDayOfWeek time.Time, events [][]*domain.EventView, birthdays [][]*domain.BirthdayView) []domain.DayView {
	monday := domain.DayView{
		DayName:            "Monday (" + firstDayOfWeek.Format("1/2") + ")",
		IsCurrentDayString: util.GetCurrentDayForSchedule(time.Monday),
		Events:             events[0],
		Birthdays:          birthdays[0]}
	tuesday := domain.DayView{
		DayName:            "Tuesday (" + firstDayOfWeek.AddDate(0, 0, 1).Format("1/2") + ")",
		IsCurrentDayString: util.GetCurrentDayForSchedule(time.Tuesday),
		Events:             events[1],
		Birthdays:          birthdays[1]}
	wednesday := domain.DayView{
		DayName:            "Wednesday (" + firstDayOfWeek.AddDate(0, 0, 2).Format("1/2") + ")",
		IsCurrentDayString: util.GetCurrentDayForSchedule(time.Wednesday),
		Events:             events[2],
		Birthdays:          birthdays[2]}
	thursday := domain.DayView{
		DayName:            "Thursday (" + firstDayOfWeek.AddDate(0, 0, 3).Format("1/2") + ")",
		IsCurrentDayString: util.GetCurrentDayForSchedule(time.Thursday),
		Events:             events[3],
		Birthdays:          birthdays[3]}
	friday := domain.DayView{
		DayName:            "Friday (" + firstDayOfWeek.AddDate(0, 0, 4).Format("1/2") + ")",
		IsCurrentDayString: util.GetCurrentDayForSchedule(time.Friday),
		Events:             events[4],
		Birthdays:          birthdays[4]}
	saturday := domain.DayView{
		DayName:            "Saturday (" + firstDayOfWeek.AddDate(0, 0, 5).Format("1/2") + ")",
		IsCurrentDayString: util.GetCurrentDayForSchedule(time.Saturday),
		Events:             events[5],
		Birthdays:          birthdays[5]}
	sunday := domain.DayView{
		DayName:            "Sunday (" + firstDayOfWeek.AddDate(0, 0, 6).Format("1/2") + ")",
		IsCurrentDayString: util.GetCurrentDayForSchedule(time.Sunday),
		Events:             events[6],
		Birthdays:          birthdays[6]}

	days := []domain.DayView{
		monday,
		tuesday,
		wednesday,
		thursday,
		friday,
		saturday,
		sunday,
	}
	return days
}

func sortEventList(evts []*domain.EventView) []*domain.EventView {
	sort.Slice(evts, func(i, j int) bool {
		return evts[i].StartTimestamp < evts[j].StartTimestamp
	})
	return evts
}

func appendEventToList(targetList []*domain.EventView, e *domain.Event) []*domain.EventView {
	return append(targetList, &domain.EventView{
		PrettyPrint:    e.ToString(),
		StartTimestamp: e.StartTimestamp,
		HasPassed:      time.Now().In(util.ServerLoc).After(e.StartTime),
		DayOfWeek:      e.StartTime.Weekday().String(),
		HostName:       e.HostName,
		HostLocation:   e.EventLocation,
		EventName:      e.EventName})
}

func appendBirthdayToList(targetList []*domain.BirthdayView, e *domain.Birthday) []*domain.BirthdayView {
	t := e.GetTimeFromBirthday()
	return append(targetList, &domain.BirthdayView{
		PrettyPrint: e.ToString(),
		DayOfWeek:   t.Weekday().String(),
		HostName:    e.GuildUser.Username})
}

// TODO: check if another refresh is in progress?
func parseAndSendSched(ctx *exrouter.Context) {
	ParseTemplate(ctx.Msg.GuildID)
	channel := FindSchedChannel(BotInstance, ctx.Msg.GuildID)

	SendSchedule(channel.ID, ctx.Msg.GuildID, BotInstance)
}

func sendHelpMessage(ctx *exrouter.Context) {
	dmChannel, err := ctx.Ses.UserChannelCreate(ctx.Msg.Author.ID)
	if err != nil {
		log.Println("Error sending DM: ", err)
		ctx.Reply("I couldn't send you a DM for some reason, sorry!")
		return
	}
	msg, _ := ctx.Reply("Ok! You should receive a help message in your DMs shortly<3")

	SendHelpImage(dmChannel)

	// Wait 5 seconds and delete the confirmation message + the user message if possible
	go func() {
		time.Sleep(5 * time.Second)
		ctx.Ses.ChannelMessageDelete(msg.ChannelID, msg.ID)
		ctx.Ses.ChannelMessageDelete(msg.ChannelID, ctx.Msg.ID)
	}()

}

/* NOTE:
 * All set/reset add/remove functions return a bool
 * which tells the bot whether or not to update the schedule
 *
 * All replies should be handled in their respective functions
 */

func resetBirthday(ctx *exrouter.Context) bool {
	guild, _ := BotInstance.ClientSession.Guild(ctx.Msg.GuildID)
	birthday, _ := BotInstance.BirthdayDao.GetBirthdayByUser(guild, ctx.Msg.Author)

	if birthday != nil {
		// If the birthday exists, we need to see if we need to refresh the schedule
		scheduleNeedsReset := birthday.IsBirthdayInCurrentWeek()
		// Add them to our cooldown cache just in case they are attempting to spam
		BirthdayCooldownUsersMap[ctx.Msg.Author.ID] = birthday.LastSetTime

		// Then delete their birthday
		BotInstance.BirthdayDao.DeleteBirthdayByUserID(birthday.GuildUserID)

		// Provide confirmation and return the needsReset value
		ctx.Reply("Ok, your birthday has been reset!")
		return scheduleNeedsReset
	}

	// Otherwise don't refresh
	ctx.Reply("Hmm, I couldn't find your birthday, maybe you didn't set it first?")

	return false
}

func setBirthday(ctx *exrouter.Context) bool {
	guild, _ := BotInstance.ClientSession.Guild(ctx.Msg.GuildID)
	birthday, _ := BotInstance.BirthdayDao.GetBirthdayByUser(guild, ctx.Msg.Author)

	// In case the user resets their birthday, we want a cache of previous cooldowns
	lastSetTime, isInCooldownMap := BirthdayCooldownUsersMap[ctx.Msg.Author.ID]
	isOnCooldown := false
	// Default expiry time to now
	expiryTime := time.Now()
	if isInCooldownMap {
		// But set it to the proper time if we have it
		expiryTime = lastSetTime.Add(15 * time.Minute)
	} else if birthday != nil {
		expiryTime = birthday.LastSetTime.Add(15 * time.Minute)
	}

	// If they are on cooldown, error out
	if isInCooldownMap && expiryTime.After(time.Now()) {
		isOnCooldown = true
	} else if birthday != nil && birthday.LastSetTime.Add(15*time.Minute).After(time.Now()) {
		// Or if the birthday obj exists and that says they are on cooldown, add them to our cache and error out
		BirthdayCooldownUsersMap[ctx.Msg.Author.ID] = birthday.LastSetTime
		isOnCooldown = true
	} else if isInCooldownMap && expiryTime.Before(time.Now()) {
		// If they are no longer on cooldown, remove them from our cache
		delete(BirthdayCooldownUsersMap, ctx.Msg.Author.ID)
	}

	if isOnCooldown {
		ctx.Reply("Sorry, you are on cooldown from this command! You may only set your birthday once every 15 minutes.")
		return false
	}

	if dateString := ctx.Args.Get(0); "" != dateString {
		t, isValid := validateDateString(ctx, dateString)
		// Make sure the year is current? Shouldn't really matter as the year never touches the DB...
		t = time.Date(time.Now().Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		if !isValid {
			return false
		}

		embedText := "set to "

		guild, _ := BotInstance.ClientSession.Guild(ctx.Msg.GuildID)
		if birthday != nil {
			embedText = "updated to "
			birthday.BirthdayDay = t.Day()
			birthday.BirthdayMonth = int(t.Month())
			birthdayID := BotInstance.BirthdayDao.UpdateBirthdayByUser(birthday, ctx.Msg.Author)
			birthday, _ = BotInstance.BirthdayDao.GetBirthdayByID(birthdayID, guild, ctx.Msg.Author)
		} else {
			birthday = new(domain.Birthday)
			birthday.ServerID = ctx.Msg.GuildID
			birthday.GuildUserID = ctx.Msg.Author.ID
			birthday.BirthdayDay = t.Day()
			birthday.BirthdayMonth = int(t.Month())
			birthday = BotInstance.BirthdayDao.InsertBirthday(birthday, BotInstance.ClientSession, guild)
		}
		if birthday == nil {
			log.Println("Error getting birthday after insert")
			return false
		}

		embed := GetEmbedFromBirthday(birthday, embedText, ctx.Msg.Author)
		BotInstance.ClientSession.ChannelMessageSendEmbed(ctx.Msg.ChannelID, embed)

		// Return false if we shouldn't update the schedule
		return birthday.IsBirthdayInCurrentWeek()
	}

	return false
}

func postEventStats(ctx *exrouter.Context) {
	count := BotInstance.EventDao.GetEventCountForServer(ctx.Msg.GuildID)
	weekTime := util.GetCurrentWeekFromMondayAsTime()
	weekCount := BotInstance.EventDao.GetEventsCountForServerForWeek(ctx.Msg.GuildID, weekTime)
	if count < 0 || weekCount < 0 {
		ctx.Reply("Error retrieving stats, please try again later.")
		return
	}

	guild, err := ctx.Guild(ctx.Msg.GuildID)
	if err != nil {
		ctx.Reply("Error retrieving stats, please try again later.")
		return
	}

	nearestEvent, err := BotInstance.EventDao.GetNextEventOrDefault(guild.ID, guild)
	nearestEventStr := ""
	if err != nil {
		log.Println("Error retrieving next event")
		ctx.Reply("Error retrieving stats, please try again later.")
		return
	} else if nearestEvent != nil {
		nearestEventStr = getMinutesOrHoursInRelationToClosestEvent(nearestEvent)
	} else {
		ctx.Reply("It looks like this server hasn't had any events yet!")
		return
	}
	statField := util.GetField("Event stats for *"+guild.Name+"*",
		"Events held in this server - **"+strconv.Itoa(count)+"** *("+strconv.Itoa(weekCount)+" this week)*"+
			nearestEventStr, false)
	emb := util.GetEmbed("", "", true, "", statField)
	ms := &discordgo.MessageSend{
		Embed: emb}
	BotInstance.ClientSession.ChannelMessageSendComplex(ctx.Msg.ChannelID, ms)
}

func addEvent(ctx *exrouter.Context) bool {
	if len(ctx.Args) < 4 {
		ctx.Reply("Please check your command format and try again.")
		return false
	}
	event := new(domain.Event)
	event.ServerID = ctx.Msg.GuildID
	event.CreatorID = ctx.Msg.Author.ID
	event.CreationTimestamp = time.Now().Unix()
	event.DurationMinutes = 120
	if !validateNewEventArgs(ctx, event) {
		log.Println("Error validating event args")
		return false
	}

	log.Println(event.StartTime)

	guildEvent := createGuildEvent(ctx, event)
	event.EventID = guildEvent.ID

	guild, _ := BotInstance.ClientSession.Guild(ctx.Msg.GuildID)
	event = BotInstance.EventDao.InsertEvent(event, BotInstance.ClientSession, guild)
	if event == nil {
		log.Println("Error getting event after insert")
		return false
	}

	embed := GetEmbedFromEvent(event, "scheduled for ")
	BotInstance.ClientSession.ChannelMessageSendEmbed(ctx.Msg.ChannelID, embed)

	// If our event is outside of the current week period, dont refresh the schedule
	if event.StartTime.After(util.GetCurrentWeekFromMondayAsTime().AddDate(0, 0, 7)) {
		return false
	}
	return true
}

func removeEvent(ctx *exrouter.Context) bool {
	if len(ctx.Args) < 4 {
		ctx.Reply("Please check your command format and try again.")
		return false
	}
	t, isValid := validateDateString(ctx, ctx.Args.Get(2))
	if !isValid {
		// The method call above handles outputting the error to the user and console.
		return false
	}

	guild, _ := BotInstance.ClientSession.Guild(ctx.Msg.GuildID)
	referencedEvent, err := BotInstance.EventDao.GetEventByStartTime(ctx.Msg.GuildID, t.Unix(), guild)
	if err != nil || referencedEvent == nil {
		ctx.Reply("Could not find that event, please try again")
		return false
	}

	if removeGuildEvent(ctx, referencedEvent) {
		BotInstance.EventDao.DeleteEventByID(referencedEvent.EventID)
		embed := GetEmbedFromEvent(referencedEvent, "deleted from ")
		BotInstance.ClientSession.ChannelMessageSendEmbed(ctx.Msg.ChannelID, embed)

		// If our event is outside of the current week period, dont refresh the schedule
		if referencedEvent.StartTime.After(util.GetCurrentWeekFromMondayAsTime().AddDate(0, 0, 7)) {
			return false
		}
		return true
	}

	return false
}

func removeGuildEvent(ctx *exrouter.Context, referencedEvent *domain.Event) bool {
	err := ctx.Ses.GuildScheduledEventDelete(ctx.Msg.GuildID, referencedEvent.EventID)
	if err != nil {
		log.Printf("Couldn't remove event")
		ctx.Reply("Something went wrong removing the event. Please try again later.")
		return false
	}
	return true
}

// Create a guild event
func createGuildEvent(ctx *exrouter.Context, event *domain.Event) *discordgo.GuildScheduledEvent {
	// Create the event
	scheduledEvent, err := ctx.Ses.GuildScheduledEventCreate(ctx.Msg.GuildID, &discordgo.GuildScheduledEventParams{
		Name:               event.EventName,
		Description:        "Hosted by - " + event.HostName,
		ScheduledStartTime: &event.StartTime,
		ScheduledEndTime:   &event.EndTime,
		EntityType:         discordgo.GuildScheduledEventEntityTypeExternal,
		EntityMetadata:     &discordgo.GuildScheduledEventEntityMetadata{Location: event.EventLocation},
		PrivacyLevel:       discordgo.GuildScheduledEventPrivacyLevelGuildOnly,
	})

	if err != nil {
		log.Printf("Error creating scheduled event: %v", err)
		return nil
	}

	fmt.Println("Created scheduled event:", scheduledEvent.Name)
	return scheduledEvent
}

func validateNewEventArgs(ctx *exrouter.Context, event *domain.Event) bool {
	t, isValid := validateDateString(ctx, strconv.FormatInt(event.StartTimestamp, 10))
	if !isValid {
		// The method call above handles outputting the error to the user and console.
		return false
	}

	event.StartTimestamp = t.Unix()
	event.StartTime = t
	event.EndTime = t.Add(time.Hour * 2)

	if hostName := ctx.Args.Get(2); "" != hostName {
		event.HostName = hostName
	} else {
		ctx.Reply("Please ensure you have included a host to your event")
		return false
	}
	if name := ctx.Args.Get(3); "" != name {
		event.EventName = name
	} else {
		ctx.Reply("Please ensure you have given the event a name")
		return false
	}
	if location := ctx.Args.Get(4); "" != location {
		event.EventLocation = location
	} else {
		ctx.Reply("Please ensure you have given the event a location")
		return false
	}
	return true
}

func validateDateString(ctx *exrouter.Context, dateString string) (time.Time, bool) {
	if dateString := ctx.Args.Get(1); "" != dateString {
		t, err := dateparse.ParseLocal(dateString)
		if err != nil {
			log.Println("Invalid time format? ", err)
			ctx.Reply("Please check your date format and try again")
			return time.Now().In(util.ServerLoc), false
		}
		// Don't adjust to server time?
		if t.Before(time.Now()) {
			ctx.Reply("Make sure the scheduled event is not in the past!")
			return time.Now(), false
		}
		return t, true
	}

	return time.Now(), false
}

// GetEmbedFromEvent : Returns a discord embed with the relevant event details
func GetEmbedFromEvent(event *domain.Event, eventEmbedText string) *discordgo.MessageEmbed {
	baseField := util.GetField("Event "+eventEmbedText+event.StartTime.Format("January 2, 2006"), event.ToEmbedString(), false)
	baseEmbed := util.GetEmbed("", "", false, "", baseField)
	return baseEmbed
}

// GetEmbedFromBirthday : Returns a discord embed with the relevant birthday details
func GetEmbedFromBirthday(birthday *domain.Birthday, birthdayEmbedText string, user *discordgo.User) *discordgo.MessageEmbed {
	baseField := util.GetField(user.Username+"'s birthday has been set!",
		// "set to" or "updated to" depending on context
		"Birthday "+birthdayEmbedText+strconv.Itoa(birthday.BirthdayMonth)+"/"+strconv.Itoa(birthday.BirthdayDay),
		false)
	baseEmbed := util.GetEmbed("", "", false, "", baseField)
	return baseEmbed
}

// GetAnnounceEmbedFromBirthday : Returns a discord embed with the relevant birthday details
func GetAnnounceEmbedFromBirthday(birthday *domain.Birthday, user *discordgo.User) *discordgo.MessageEmbed {
	baseField := util.GetField(user.Username+"'s birthday is today!",
		// "set to" or "updated to" depending on context
		"🎂❤️Make sure to wish them a happy birthday❤️🎂",
		false)
	baseEmbed := util.GetEmbed("", "", true, user.AvatarURL("128"), baseField)
	return baseEmbed
}

// GetAnnounceEmbedFromEvent : Gets an announcement embed from the given event
func GetAnnounceEmbedFromEvent(event *domain.Event, eventEmbedText string, eventEmbedTitle string) *discordgo.MessageEmbed {
	baseField := util.GetField(eventEmbedTitle, eventEmbedText, false)
	baseEmbed := util.GetEmbed("", "", false, "", baseField)
	return baseEmbed
}

// AuthAdmin : Authenticates the user has admin perm
func AuthAdmin(ctx *exrouter.Context) bool {
	return MemberHasPermission(ctx.Ses, ctx.Msg.GuildID, ctx.Msg.Author.ID, 8)
}

// AuthEventRunner : Authenticates against bot roles in config
func AuthEventRunner(ctx *exrouter.Context) bool {
	return checkRoleByName(ctx, BotInstance.EventRunnerRoleName)
}

// GatedRole : Checks if the user has the gated role for commands
func GatedRole(ctx *exrouter.Context) bool {
	return checkRoleByName(ctx, BotInstance.GatedRoleName)
}

func checkRoleByName(ctx *exrouter.Context, roleName string) bool {
	member, err := ctx.Member(ctx.Msg.GuildID, ctx.Msg.Author.ID)
	if err != nil {
		ctx.Reply("Could not fetch member: ", err)
		return false
	}
	roleToCheck, err := FindRoleByName(ctx.Msg.GuildID, roleName)
	if err != nil {
		log.Println("Error getting role", err)
	}
	if roleToCheck == nil {
		ctx.Reply("Oops, your server has not been configured properly!\n" +
			"I can't find the role named `" + roleName + "`.")
		return false
	}

	// Admin or has role
	if MemberHasPermission(ctx.Ses, ctx.Msg.GuildID, ctx.Msg.Author.ID, 0x8) ||
		sliceutil.Contains(member.Roles, roleToCheck.ID) {
		return true
	}

	ctx.Reply("You don't have permission to use this command, you need `" + roleName + "` to do this.")
	return false
}

func findRoleByID(ctx *exrouter.Context, roleID string) (*discordgo.Role, error) {
	var foundRole *discordgo.Role
	guildRoles, _ := ctx.Ses.GuildRoles(ctx.Msg.GuildID)
	for _, r := range guildRoles {
		if r.ID == roleID {
			foundRole = r
			break
		}
	}
	if foundRole != nil {
		return foundRole, nil
	}
	return nil, errors.New("No role found")
}

// GetNewFact : Sets the in-memory fact
// If the fact has not changed since the last time it was checked, it will fallback to a user fact
// And if a user fact cannot be found, then it will default to an explanation to encourage users to enter their own facts
// about themselves
func GetNewFact(inst *bot.Instance, currentFact string, isUserFact bool) (string, string) {
	log.Println("Updating fact...")

	//build request
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(config.GetString("RssLink"))
	if err != nil || isUserFact {
		return fallbackGetUserFactOrDefault(inst, err)
	}

	if len(feed.Items) <= 0 {
		return fallbackGetUserFactOrDefault(inst, nil)
	}

	curItem := feed.Items[0]
	title := curItem.Title
	content := strip.StripTags(curItem.Description)

	// If the content is empty (video link) or is the same as the last acquired fact
	if len(content) <= 0 || content == currentFact {
		return fallbackGetUserFactOrDefault(inst, nil)
	}

	return title, content
}

func fallbackGetUserFactOrDefault(inst *bot.Instance, err error) (string, string) {
	log.Println("Getting user fact instead of normal fact", err)
	userFact := getUserFact(inst)
	if userFact == nil {
		return "There was an issue getting the fact.", "Sorry! There's been an issue getting the fact for today ;-;<br/>" +
			"If you want to enter in a fact about yourself, do so with the command !mh fact \"Fact about me\" to enter it into the rotation!<br/>" +
			"To keep things fresh, once a user fact has been shown it will not be shown for another 7 days"
	}
	return "Did you know this about " + userFact.User.Username + "?", userFact.User.Username + " says: " + userFact.FactContent
}

// FindRoleByName : Gets a role by name
func FindRoleByName(guildID string, roleName string) (*discordgo.Role, error) {
	var foundRole *discordgo.Role
	guildRoles, _ := BotInstance.ClientSession.GuildRoles(guildID)
	for _, r := range guildRoles {
		if r.Name == roleName {
			foundRole = r
			break
		}
	}
	if foundRole != nil {
		return foundRole, nil
	}
	return nil, errors.New("No role found")
}

// FindSchedChannel : Finds the schedule channel for a given guild
func FindSchedChannel(inst *bot.Instance, guildID string) *discordgo.Channel {
	guildChannels, err := inst.ClientSession.GuildChannels(guildID)
	if err != nil {
		log.Println("Error while fetching guild ", err)
	}

	var schedChannel *discordgo.Channel
	for _, schedChannel = range guildChannels {
		if schedChannel.Name == inst.ScheduleChannel {
			break
		}
		schedChannel = nil
	}
	if schedChannel == nil {
		log.Println("Sched channel not found")
		return nil
	}
	return schedChannel
}

//FindAnnouncementsChannel : Finds the announcements channel for a given guild
func FindAnnouncementsChannel(guild *discordgo.Guild, inst *bot.Instance) *discordgo.Channel {
	var announcementChannel *discordgo.Channel
	for _, announcementChannel = range guild.Channels {
		if announcementChannel.Name == inst.AnnouncementChannel {
			break
		}
		announcementChannel = nil
	}
	if announcementChannel == nil {
		log.Println("Announcement channel not found")
		return nil
	}
	return announcementChannel
}

// GetSchedMessage : Gets the current schedule image posted, if exists, else nil
func GetSchedMessage(schedChannelID string, inst *bot.Instance) (*discordgo.Message, error) {
	msgHistory, err := inst.ClientSession.ChannelMessages(schedChannelID, 100, "", "", "")
	if err != nil {
		log.Println("Couldn't find sched message")
		return nil, err
	}

	var schedMsg *discordgo.Message
	for _, msg := range msgHistory {
		// Delete any extra schedules that may have been posted
		if schedMsg != nil && msg.Content != "@everyone" {
			err = inst.ClientSession.ChannelMessageDelete(msg.ChannelID, msg.ID)
			if err != nil {
				log.Println("Error deleting message")
			}
		} else if msg != nil && msg.Author.ID == inst.ClientSession.State.User.ID && msg.Content != "@everyone" {
			schedMsg = msg
		}
	}
	if schedMsg == nil {
		return nil, nil
	}
	return schedMsg, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func deleteHelpImage() {
	path, _ := os.Getwd()
	helpFileName := "help.png"
	helpExists, _ := exists(path + helpFileName)
	if helpExists {
		os.Remove(helpFileName)
	}
}

// SendHelpImage : Send the screenshotted help.html file in a DM
func SendHelpImage(dmChannel *discordgo.Channel) {
	path, _ := os.Getwd()
	helpFileName := "help.png"

	helpExists, err := exists(path + helpFileName)
	if err != nil || !helpExists {
		// File doesn't exist... so make it
		chrome.TakeScreenshot(defaultScreenshotW, defaultScreenshotH, "#today", helpFileName, "file:///"+path+"/web/help.html", false)
	}

	fHelp, err := os.Open(helpFileName)
	if err != nil {
		log.Println("Error getting schedule banner", err)
		fHelp.Close()
		return
	}

	body := ""

	msHelp := &discordgo.MessageSend{
		Content: body,
		Files: []*discordgo.File{
			&discordgo.File{
				Name:   helpFileName,
				Reader: fHelp,
			},
		},
	}

	BotInstance.ClientSession.ChannelMessageSendComplex(dmChannel.ID, msHelp)
	fHelp.Close()
}

// SendSchedule : Parses and sends the schedule message to a given channel
func SendSchedule(schedChannelID string, guildID string, inst *bot.Instance, isFirstSchedOfWeek ...bool) {
	log.Println("Send sched fired")
	schedMsg, err := GetSchedMessage(schedChannelID, inst)
	if err != nil {
		return
	}

	if schedMsg != nil && schedMsg.Content != "@everyone" {
		inst.ClientSession.ChannelMessageDelete(schedMsg.ChannelID, schedMsg.ID)
	}
	ParseTemplate(guildID)

	firstSched := false
	if len(isFirstSchedOfWeek) > 0 {
		// firstSched = isFirstSchedOfWeek[0]
	}
	go takeAndSendTargeted(schedChannelID, guildID, inst, firstSched)
}

func takeAndSendTargeted(schedChannelID string, guildID string, inst *bot.Instance, isFirstSchedOfWeek bool) {
	path, _ := os.Getwd()

	var wg sync.WaitGroup

	// Add our two screenshots to the wait group
	wg.Add(2)

	scheduleFileName := "schedule" + guildID + ".png"
	todayFileName := "today" + guildID + ".png"

	isScheduleTargeted := config.GetBool("isScheduleTargeted")
	isTodayTargeted := config.GetBool("isTodayTargeted")

	go chrome.TakeScreenshot(defaultScreenshotW, defaultScreenshotH, "#main", scheduleFileName, "file:///"+path+"/web/schedule-parsed"+guildID+".html", isScheduleTargeted, &wg)

	// For some reason the non-targeted screenshot will add ~300px extra margin to the bottom, so take a targeted screenshot instead
	go chrome.TakeScreenshot(defaultScreenshotW, defaultScreenshotH, "#today", todayFileName, "file:///"+path+"/web/today-parsed"+guildID+".html", isTodayTargeted, &wg)

	// Wait for both screenshots to be finished
	wg.Wait()

	fSched, err := os.Open(scheduleFileName)
	if err != nil {
		log.Println("Error getting schedule banner", err)
		fSched.Close()
		return
	}

	fFacts, err := os.Open(todayFileName)
	if err != nil {
		log.Println("Error getting schedule banner", err)
		fFacts.Close()
		return
	}

	body := ""

	msSched := &discordgo.MessageSend{
		Content: body,
		Files: []*discordgo.File{
			&discordgo.File{
				Name:   scheduleFileName,
				Reader: fSched,
			},
			&discordgo.File{
				Name:   todayFileName,
				Reader: fFacts,
			},
		},
	}

	BotInstance.ClientSession.ChannelMessageSendComplex(schedChannelID, msSched)
	fFacts.Close()
	fSched.Close()
}

/* ------------- User Facts ------------- */
func getUserFact(inst *bot.Instance) *domain.Fact {
	fact, err := inst.FactDao.GetRandomFact()
	if err != nil || fact == nil {
		return nil
	}
	return fact
}

func deleteFact(ctx *exrouter.Context) bool {
	return BotInstance.FactDao.DeleteFactByUserID(ctx.Msg.Author.ID)
}

func insertFact(ctx *exrouter.Context) *domain.Fact {
	// guild, _ := BotInstance.ClientSession.Guild(ctx.Msg.GuildID)
	originalFact := parseFact(ctx)
	// Ensure users can't inject HTML
	originalFact.FactContent = strip.StripTags(originalFact.FactContent)

	// And any emoji are clipped from the content
	emojiRegex := regexp.MustCompile("<(a)?:.*?:(.*?)>")
	originalFact.FactContent = strip.StripTags(emojiRegex.ReplaceAllString(originalFact.FactContent, ""))

	if originalFact != nil && len(originalFact.FactContent) >= 500 {
		ctx.Reply("Sorry, facts cannot be longer than 500 characters")
		return nil
	}

	if originalFact.FactContent != "" {
		fact := BotInstance.FactDao.InsertFact(originalFact, ctx.Ses)
		return fact
	}
	ctx.Reply("You need to enter a fact!\nExample: `!mh fact My Fact Here`")
	return nil
}

func parseFact(ctx *exrouter.Context) *domain.Fact {
	// Everything after "Prefix Fact"
	if ctx.Args.After(1) == "" {
		return nil
	}
	return &domain.Fact{
		UserID:      ctx.Msg.Author.ID,
		FactContent: ctx.Args.After(1)}
}

/* -------------------------------------- */

func getMinutesOrHoursInRelationToClosestEvent(nearestEvent *domain.Event) string {
	minutesUntilNext := time.Until(nearestEvent.StartTime)
	tilOrSinceStr := "until the next event!"
	if tilOrSince := minutesUntilNext.Minutes(); tilOrSince < 0 {
		tilOrSinceStr = "since the last event!"
	}
	minutesStr := ""
	if math.Abs(minutesUntilNext.Minutes()) >= 60 {
		hrs := math.Abs(math.Round(minutesUntilNext.Hours()))
		hrsPlural := ""
		if hrs <= 1 {
			hrsPlural = " hour "
		} else {
			hrsPlural = " hours "
		}
		minutesStr = strconv.Itoa(int(math.Abs(hrs))) + hrsPlural + tilOrSinceStr
	} else {
		minutesStr = strconv.Itoa(int(math.Abs(minutesUntilNext.Minutes()))) +
			" minutes " + tilOrSinceStr
	}
	return "\n" + minutesStr
}

func exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err != nil, err
}
