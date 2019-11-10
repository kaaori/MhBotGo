package domain

import (
	"reflect"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestBirthday_IsToday(t *testing.T) {
	tests := []struct {
		name string
		b    Birthday
		want bool
	}{
		{name: "Birthday is not today", b: Birthday{BirthdayDay: time.Now().Add(24 * time.Hour).Day(), BirthdayMonth: int(time.Now().Add(24 * time.Hour).Month())}, want: false},
		{name: "Birthday is today", b: Birthday{BirthdayDay: time.Now().Day(), BirthdayMonth: int(time.Now().Month())}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.IsToday(); got != tt.want {
				t.Errorf("Birthday.IsToday() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBirthday_GetTimeFromBirthday(t *testing.T) {
	tests := []struct {
		name string
		b    Birthday
		want time.Time
	}{
		{name: "Birthday is Nov 3", b: Birthday{BirthdayDay: time.Date(2019, 11, 3, time.Now().Hour(), 0, 0, 0, time.Now().Location()).Day(),
			BirthdayMonth: int(time.Date(2019, 11, 3, time.Now().Hour(), 0, 0, 0, time.Now().Location()).Month())},
			want: time.Date(2019, 11, 3, time.Now().Hour(), 0, 0, 0, time.Now().Location())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.GetTimeFromBirthday(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Birthday.GetTimeFromBirthday() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBirthday_IsBirthdayInCurrentWeek(t *testing.T) {
	tests := []struct {
		name string
		b    Birthday
		want bool
	}{
		{name: "Birthday is in current week", b: Birthday{BirthdayDay: time.Now().Add(24 * time.Hour).Day(), BirthdayMonth: int(time.Now().Add(24 * time.Hour).Month())}, want: true},
		{name: "Birthday is not in current week", b: Birthday{BirthdayDay: time.Now().AddDate(0, 0, 8).Day(), BirthdayMonth: int(time.Now().AddDate(0, 0, 8).Month())}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.IsBirthdayInCurrentWeek(); got != tt.want {
				t.Errorf("Birthday.IsBirthdayInCurrentWeek() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBirthday_ToString(t *testing.T) {
	tests := []struct {
		name string
		b    Birthday
		want string
	}{
		//"<strong><em> " + strip.StripTags(emojiRegex.ReplaceAllString(b.GuildUser.Username, "")) + "'s</strong></em> birthday!"
		{name: "TestUser's Bday", b: Birthday{GuildUser: &discordgo.User{Username: "TestUser"}}, want: "<strong><em> TestUser's</strong></em> birthday!"},
		{name: "TestUser's Bday with stripped emoji", b: Birthday{GuildUser: &discordgo.User{Username: "TestUser<a:diiLick:489912538021101570>"}}, want: "<strong><em> TestUser's</strong></em> birthday!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.ToString(); got != tt.want {
				t.Errorf("Birthday.ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
