package handlers

import (
	"testing"

	"fm/pingbot/model"
	tele "gopkg.in/telebot.v3"
)

func TestParseAddMessage(t *testing.T) {
	tests := []struct {
		text  string
		name  string
		users []string
	}{
		{"/add @alice @bob", model.MentionEveryoneName, []string{"alice", "bob"}},
		{"/add team @alice @bob", "team", []string{"alice", "bob"}},
		{"/add @alice team @bob", "team", []string{"alice", "bob"}},
		{"/add @alice @bob team", "team", []string{"alice", "bob"}},
		{"/add", model.MentionEveryoneName, nil},
		{"/add new team", "new team", nil},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			message := &tele.Message{Text: tt.text}
			message.Entities = append(message.Entities, tele.MessageEntity{Type: tele.EntityCommand, Offset: 0, Length: 4})
			for _, user := range tt.users {
				start := -1
				// Find each unique tagged user in the test message.
				for i := 0; i+len(user)+1 <= len(tt.text); i++ {
					if tt.text[i:i+len(user)+1] == "@"+user {
						start = i
						break
					}
				}
				message.Entities = append(message.Entities, tele.MessageEntity{Type: tele.EntityMention, Offset: start, Length: len(user) + 1})
			}
			name, users := parseAddMessage(message)
			if name != tt.name || len(users) != len(tt.users) {
				t.Fatalf("got name %q, users %v; want %q, %v", name, users, tt.name, tt.users)
			}
			for i, user := range users {
				if user.Username != tt.users[i] {
					t.Fatalf("user %d: got %q, want %q", i, user.Username, tt.users[i])
				}
			}
		})
	}
}
