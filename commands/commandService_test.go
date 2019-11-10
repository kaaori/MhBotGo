package commands

import (
	"testing"
	"time"

	"github.com/kaaori/MhBotGo/dao"

	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/bot"
)

var (
	ses = initSession()
)

func TestMain(m *testing.M) {
	BotInstance = &bot.Instance{
		ClientSession: ses,
		ServerDao:     dao.DiscordServerDao{ses},
		EventDao:      dao.EventDao{ses},
		BirthdayDao:   dao.BirthdayDao{ses},
	}
	dao.ConnString = "file::memory:?cache=shared"
}

// initSession : Set up our mock server, user, and role
func initSession() *discordgo.Session {
	ses := &discordgo.Session{StateEnabled: true, State: discordgo.NewState()}

	ses.State.GuildAdd(&discordgo.Guild{ID: "guild"})
	ses.State.RoleAdd("guild", &discordgo.Role{
		ID:          "role",
		Name:        "BotAdmin",
		Permissions: 0x8,
		Mentionable: true,
	})
	user := &discordgo.User{
		ID:       "user",
		Username: "User Name",
	}
	user2 := &discordgo.User{
		ID:       "user2",
		Username: "User Name2",
	}

	ses.State.MemberAdd(&discordgo.Member{
		User:    user,
		Nick:    "User Nick",
		GuildID: "guild",
		Roles:   []string{"role"},
	})
	ses.State.MemberAdd(&discordgo.Member{
		User:    user2,
		Nick:    "User Nick2",
		GuildID: "guild",
		Roles:   []string{"role2"},
	})

	ses.State.ChannelAdd(&discordgo.Channel{
		Name:    "Channel Name",
		GuildID: "guild",
		ID:      "channel",
	})
	return ses
}

func TestMemberHasPermission(t *testing.T) {
	ses := initSession()

	type args struct {
		s          *discordgo.Session
		guildID    string
		userID     string
		permission int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Test with perms", args{ses, "guild", "user", 0}, true},
		{"Test with no perms", args{ses, "guild", "user2", 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MemberHasPermission(tt.args.s, tt.args.guildID, tt.args.userID, tt.args.permission); got != tt.want {
				t.Errorf("MemberHasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_contains(t *testing.T) {
	type args struct {
		s []string
		e string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Contains in slice", args{[]string{"haystack", "haystack", "needle"}, "needle"}, true},
		{"Containsn't in slice", args{[]string{"haystack", "haystack", "notaneedle"}, "needle"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.args.s, tt.args.e); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func Test_setBirthday(t *testing.T) {
// 	insertTestBirthday()

// 	type args struct {
// 		ctx *exrouter.Context
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want bool
// 	}{
// 		{"Test successful set", args{ctx: exrouter.NewContext(ses,
// 			&discordgo.Message{Author: &discordgo.User{
// 				ID:       "user",
// 				Username: "User Name",
// 			}}, nil, nil)}},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := setBirthday(tt.args.ctx); got != tt.want {
// 				t.Errorf("setBirthday() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func insertTestBirthday() {
	DB := dao.GetConnection(dao.ConnString)
	defer DB.Close()

	query := `INSERT INTO Birthdays 
	(ServerID, UserID, BirthMonthNum, BirthDayNum, LastSet) 
	VALUES
	(?,?,?,?,?)`

	stmt, err := DB.Prepare(query)
	if err != nil {
		log.Error("Error inserting server", err)
	}
	defer stmt.Close()

	// We need to increment the birthday by 1 due to how it's being parsed
	err = stmt.Exec("guild", "user", "11", "10", time.Now().Unix())
	if err != nil {
		log.Error("Error inserting birthday", err)
	}
}
