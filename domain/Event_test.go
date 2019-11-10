package domain

import (
	"testing"
	"time"
)

func TestEvent_ToString(t *testing.T) {
	tests := []struct {
		name string
		e    Event
		want string
	}{
		{"Test with emoji", Event{EventName: "Test<a:diiLick:489912538021101570>Name",
			EventLocation: "TestGame<a:diiLick:489912538021101570>",
			HostName:      "TestUser<a:diiLick:489912538021101570>",
			StartTime:     time.Date(2019, 11, 10, 23, 0, 0, 0, time.Now().Location())},
			"<strong><em>11:00PM - TestName</strong></em> (Hosted  by <strong><em>TestUser</em></strong> in TestGame)"},
		{"Test without emoji", Event{EventName: "TestName",
			EventLocation: "TestGame",
			HostName:      "TestUser",
			StartTime:     time.Date(2019, 11, 10, 23, 0, 0, 0, time.Now().Location())},
			"<strong><em>11:00PM - TestName</strong></em> (Hosted  by <strong><em>TestUser</em></strong> in TestGame)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.ToString(); got != tt.want {
				t.Errorf("Event.ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvent_ToEmbedString(t *testing.T) {
	tests := []struct {
		name string
		e    Event
		want string
	}{
		{"Test with emoji", Event{EventName: "Test<a:diiLick:489912538021101570>Name",
			EventLocation: "TestGame<a:diiLick:489912538021101570>",
			HostName:      "TestUser<a:diiLick:489912538021101570>",
			StartTime:     time.Date(2019, 11, 10, 23, 0, 0, 0, time.Now().Location())},
			"• *11:00PM (Eastern Standard Time)* ── **TestName** ── (Hosted  by ***TestUser*** in TestGame)"},
		{"Test without emoji", Event{EventName: "TestName",
			EventLocation: "TestGame",
			HostName:      "TestUser",
			StartTime:     time.Date(2019, 11, 10, 23, 0, 0, 0, time.Now().Location())},
			"• *11:00PM (Eastern Standard Time)* ── **TestName** ── (Hosted  by ***TestUser*** in TestGame)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.ToEmbedString(); got != tt.want {
				t.Errorf("Event.ToEmbedString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvent_ToAnnounceString(t *testing.T) {
	tests := []struct {
		name string
		e    Event
		want string
	}{
		{"Test with emoji", Event{EventName: "Test<a:diiLick:489912538021101570>Name",
			EventLocation: "TestGame<a:diiLick:489912538021101570>",
			HostName:      "TestUser<a:diiLick:489912538021101570>",
			StartTime:     time.Date(2019, 11, 10, 23, 0, 0, 0, time.Now().Location())},
			"**TestUser** is about to start this event in TestGame at **11:00 PM (Eastern Standard Time)!**"},
		{"Test without emoji", Event{EventName: "TestName",
			EventLocation: "TestGame",
			HostName:      "TestUser",
			StartTime:     time.Date(2019, 11, 10, 23, 0, 0, 0, time.Now().Location())},
			"**TestUser** is about to start this event in TestGame at **11:00 PM (Eastern Standard Time)!**"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.ToAnnounceString(); got != tt.want {
				t.Errorf("Event.ToAnnounceString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvent_ToStartingString(t *testing.T) {
	tests := []struct {
		name string
		e    Event
		want string
	}{
		{"Test with emoji", Event{EventName: "Test<a:diiLick:489912538021101570>Name",
			EventLocation: "TestGame<a:diiLick:489912538021101570>",
			HostName:      "TestUser<a:diiLick:489912538021101570>",
			StartTime:     time.Date(2019, 11, 10, 23, 0, 0, 0, time.Now().Location())},
			"Join up on **TestUser**! This event is taking place at **11:00 PM**, and will last roughly 2 hours"},
		{"Test without emoji", Event{EventName: "TestName",
			EventLocation: "TestGame",
			HostName:      "TestUser",
			StartTime:     time.Date(2019, 11, 10, 23, 0, 0, 0, time.Now().Location())},
			"Join up on **TestUser**! This event is taking place at **11:00 PM**, and will last roughly 2 hours"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.ToStartingString(); got != tt.want {
				t.Errorf("Event.ToStartingString() = %v, want %v", got, tt.want)
			}
		})
	}
}
