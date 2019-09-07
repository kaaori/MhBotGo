package commands

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"text/template"
	"time"

	"github.com/forestgiant/sliceutil"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/araddon/dateparse"
	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/bot"
	"github.com/kaaori/MhBotGo/chrome"
	"github.com/kaaori/MhBotGo/domain"
	"github.com/kaaori/MhBotGo/util"
	"github.com/snabb/isoweek"
)

// MemberHasPermission : Checks if the member has a given perm
func MemberHasPermission(s *discordgo.Session, guildID string, userID string, permission int) bool {
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
	tmpl, err := template.ParseFiles("./web/schedule-template.html")
	if err != nil {
		panic(err)
	}
	year, week := time.Now().In(util.ServerLoc).ISOWeek()
	t := isoweek.StartTime(year, week, time.Now().In(util.ServerLoc).Location())

	g, _ := BotInstance.ClientSession.Guild(guildID)
	f, err := os.Create("./web/schedule-parsed.html")
	if err != nil {
		log.Error("create file: ", err)
		return
	}
	defer f.Close()

	weekTime := util.GetCurrentWeekFromMondayAsTime()
	events, err := BotInstance.EventDao.GetAllEventsForServerForWeek(guildID, weekTime)
	if err != nil {
		log.Error("", err)
		return
	}

	monEvts := make([]*domain.EventView, 0)
	tuesEvts := make([]*domain.EventView, 0)
	wedEvts := make([]*domain.EventView, 0)
	thursEvts := make([]*domain.EventView, 0)
	friEvts := make([]*domain.EventView, 0)
	satEvts := make([]*domain.EventView, 0)
	sunEvts := make([]*domain.EventView, 0)

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
	_, isoWeek := t.ISOWeek()
	firstDayOfWeek := util.FirstDayOfISOWeek(t.Year(), isoWeek, t.Location())
	data := domain.ScheduleView{
		ServerName:        g.Name,
		CurrentWeekString: string(firstDayOfWeek.Format("January 2, 2006") + " ── " + firstDayOfWeek.AddDate(0, 0, 6).Format("January 2, 2006")),
		Tz:                "<strong>Eastern Standard Time</strong>",
		MondayEvents:      sortEventList(monEvts),
		TuesdayEvents:     sortEventList(tuesEvts),
		WednesdayEvents:   sortEventList(wedEvts),
		ThursdayEvents:    sortEventList(thursEvts),
		FridayEvents:      sortEventList(friEvts),
		SaturdayEvents:    sortEventList(satEvts),
		SundayEvents:      sortEventList(sunEvts),
		Fact:              BotInstance.CurrentFact}

	tmpl.Execute(f, data)
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
		HasPassed:      time.Now().In(util.ServerLoc).After(e.StartTime.In(util.ServerLoc)),
		DayOfWeek:      e.StartTime.Weekday().String()})
}

func parseAndSendSched(ctx *exrouter.Context) {
	ParseTemplate(ctx.Msg.GuildID)
	g, err := ctx.Guild(ctx.Msg.GuildID)
	if err != nil {
		panic("Couldn't find guild")
	}
	channel := FindSchedChannel(g, BotInstance)

	SendSchedule(channel.ID, BotInstance)
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

	nearestEvent, err := BotInstance.EventDao.GetNextEventOrDefault(guild.ID)
	nearestEventStr := ""
	if err != nil {
		log.Error("Error retrieving next event")
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
	emb := util.GetEmbed("", "", true, statField)
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
		log.Error("Error validating event args")
		return false
	}

	event = BotInstance.EventDao.InsertEvent(event, BotInstance.ClientSession)
	if event == nil {
		log.Error("Error getting event after insert")
		return false
	}

	embed := GetEmbedFromEvent(event, "scheduled for ")
	BotInstance.ClientSession.ChannelMessageSendEmbed(ctx.Msg.ChannelID, embed)
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

	referencedEvent, err := BotInstance.EventDao.GetEventByStartTime(ctx.Msg.GuildID, t.Unix()-util.ServerLocOffset)
	if err != nil || referencedEvent == nil {
		ctx.Reply("Could not find that event, please try again")
		return false
	}
	BotInstance.EventDao.DeleteEventByID(referencedEvent.EventID)
	embed := GetEmbedFromEvent(referencedEvent, "deleted from ")
	BotInstance.ClientSession.ChannelMessageSendEmbed(ctx.Msg.ChannelID, embed)
	return true
}

func validateNewEventArgs(ctx *exrouter.Context, event *domain.Event) bool {
	t, isValid := validateDateString(ctx, strconv.FormatInt(event.StartTimestamp, 10))
	if !isValid {
		// The method call above handles outputting the error to the user and console.
		return false
	}

	event.StartTimestamp = t.Unix()

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

// GetEmbedFromEvent : Returns a discord embed with the relevant event details
func GetEmbedFromEvent(event *domain.Event, eventEmbedText string) *discordgo.MessageEmbed {
	t := time.Unix(event.StartTimestamp, 0)
	timeObj := t.In(util.ServerLoc).Format("January 2, 2006")
	baseField := util.GetField("Event "+eventEmbedText+timeObj, event.ToEmbedString(), false)
	baseEmbed := util.GetEmbed("", "", false, baseField)
	return baseEmbed
}

// GetAnnounceEmbedFromEvent : Gets an announcement embed from the given event
func GetAnnounceEmbedFromEvent(event *domain.Event, eventEmbedText string, eventEmbedTitle string) *discordgo.MessageEmbed {
	baseField := util.GetField(eventEmbedTitle, eventEmbedText, false)
	baseEmbed := util.GetEmbed("", "", false, baseField)
	return baseEmbed
}

func validateDateString(ctx *exrouter.Context, dateString string) (time.Time, bool) {
	if dateString := ctx.Args.Get(1); "" != dateString {
		t, err := dateparse.ParseAny(dateString)
		if err != nil {
			log.Error("Invalid time format? ", err)
			ctx.Reply("Please check your date format and try again")
			return time.Now().In(util.ServerLoc), false
		}
		return t.In(util.ServerLoc), true
	}
	return time.Now().In(util.ServerLoc), false

}

// AuthAdmin : Authenticates the user has admin perm
func AuthAdmin(ctx *exrouter.Context) bool {
	return MemberHasPermission(ctx.Ses, ctx.Msg.GuildID, ctx.Msg.Author.ID, 8)
}

// AuthEventRunner : Authenticates against bot roles in config
func AuthEventRunner(ctx *exrouter.Context) bool {
	member, err := ctx.Member(ctx.Msg.GuildID, ctx.Msg.Author.ID)
	if err != nil {
		ctx.Reply("Could not fetch member: ", err)
		return false
	}
	eventRunnerRole, err := findRoleByName(ctx, BotInstance.EventRunnerRoleName)
	if err != nil {
		log.Error("Error getting role", err)
	}
	if eventRunnerRole == nil {
		ctx.Reply("Oops, your server has not been configured properly!\n" +
			"I can't find the role named `EventRunner`.")
		return false
	}

	if MemberHasPermission(ctx.Ses, ctx.Msg.GuildID, ctx.Msg.Author.ID, 0x8) ||
		sliceutil.Contains(member.Roles, eventRunnerRole.ID) {
		return true
	}

	ctx.Reply("You don't have permission to use this command")
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
func GetNewFact() string {
	log.Info("Updating fact")

	// Build the request
	req, err := http.NewRequest("GET", "https://uselessfacts.jsph.pl/random.json?language=en", nil)
	if err != nil {
		log.Error("NewRequest: ", err)
		return "Error retrieving facts... Sorry ;-;"
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Do: ", err)
		return "Error retrieving facts... Sorry ;-;"
	}
	defer resp.Body.Close()

	var fact domain.Fact

	if err := json.NewDecoder(resp.Body).Decode(&fact); err != nil {
		log.Error("Error decoding fact: ", err)
	}
	return fact.Text
}
func findRoleByName(ctx *exrouter.Context, roleName string) (*discordgo.Role, error) {
	var foundRole *discordgo.Role
	guildRoles, _ := ctx.Ses.GuildRoles(ctx.Msg.GuildID)
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
func FindSchedChannel(guild *discordgo.Guild, inst *bot.Instance) *discordgo.Channel {
	var schedChannel *discordgo.Channel
	for _, schedChannel = range guild.Channels {
		if schedChannel.Name == inst.ScheduleChannel {
			break
		}
		schedChannel = nil
	}
	if schedChannel == nil {
		log.Error("Sched channel not found")
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
		log.Error("Announcement channel not found")
		return nil
	}
	return announcementChannel
}

// GetSchedMessage : Gets the current schedule image posted, if exists, else nil
func GetSchedMessage(schedChannelID string, inst *bot.Instance) (*discordgo.Message, error) {
	msgHistory, err := inst.ClientSession.ChannelMessages(schedChannelID, 100, "", "", "")
	if err != nil {
		log.Error("Couldn't find sched message")
		return nil, err
	}

	var schedMsg *discordgo.Message
	for _, msg := range msgHistory {
		// Delete any extra schedules that may have been posted
		if schedMsg != nil {
			inst.ClientSession.ChannelMessageDelete(msg.ChannelID, msg.ID)
		} else if msg.Author.ID == inst.ClientSession.State.User.ID {
			schedMsg = msg
		}
	}
	return schedMsg, nil
}

// SendSchedule : Parses and sends the schedule message to a given channel
func SendSchedule(schedChannelID string, inst *bot.Instance) {
	schedMsg, err := GetSchedMessage(schedChannelID, inst)
	if err != nil {
		return
	}

	if schedMsg != nil {
		inst.ClientSession.ChannelMessageDelete(schedMsg.ChannelID, schedMsg.ID)
	}

	go takeAndSend(schedChannelID, inst)
}

func takeAndSend(schedChannelID string, inst *bot.Instance) {
	chrome.TakeScreenshot(defaultScreenshotW, defaultScreenshotH)

	f, err := os.Open("schedule.png")
	if err != nil {
		log.Error("Error getting schedule image", err)
		return
	}
	defer f.Close()

	ms := &discordgo.MessageSend{
		// Embed: &discordgo.MessageEmbed{
		// 	Title: "Click the schedule below to see more info!",
		// 	Color: 0x9400d3,
		// 	Image: &discordgo.MessageEmbedImage{
		// 		URL: "attachment://" + "schedule.png",
		// 	},
		// },
		Files: []*discordgo.File{
			&discordgo.File{
				Name:   "schedule.png",
				Reader: f,
			},
		},
	}

	BotInstance.ClientSession.ChannelMessageSendComplex(schedChannelID, ms)
}

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
