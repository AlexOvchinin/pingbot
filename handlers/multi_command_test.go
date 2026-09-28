package handlers

import (
	"strings"
	"testing"

	"fm/pingbot/model"
	tele "gopkg.in/telebot.v3"
)

func TestJoinAndLeaveMultipleMentions(t *testing.T) {
	c := setupHandlers(t)
	for _, name := range []string{"team", "ops", "new team"} {
		if err := Storage.AddMention(7, name); err != nil {
			t.Fatal(err)
		}
	}
	c.message.Payload = "new team ops"
	if err := HandleJoinCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "Successfully joined new team, ops")
	for _, name := range []string{"new team", "ops"} {
		users, _ := Storage.GetMentionUsers(7, name)
		if len(users) != 1 {
			t.Fatalf("%s has %v", name, users)
		}
	}
	if err := HandleLeaveCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "send", "Left new team\nLeft ops")
}

func TestCreateMissingMentionsThenJoinAll(t *testing.T) {
	c := setupHandlers(t)
	if err := Storage.AddMention(7, "team"); err != nil {
		t.Fatal(err)
	}
	c.message = testMessage("/join team, new group, ops")
	c.message.Payload = "team, new group, ops"
	c.message.Chat = c.chat
	if err := HandleJoinCommand(c); err != nil {
		t.Fatal(err)
	}
	if len(c.calls) != 1 || c.calls[0].method != "reply" ||
		!strings.Contains(c.calls[0].opts[0].(*tele.ReplyMarkup).InlineKeyboard[0][0].Data, CREATE_MENTIONS_COMMAND_NAME) {
		t.Fatalf("expected create prompt, got %+v", c.calls)
	}
	c.callback.Message.ReplyTo = c.message
	c.calls = nil
	if err := handleCreateMentionsCallback(c, nil); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "Successfully joined team, new group, ops")
	for _, name := range []string{"team", "new group", "ops"} {
		users, err := Storage.GetMentionUsers(7, name)
		if err != nil || len(users) != 1 {
			t.Fatalf("%s users=%v err=%v", name, users, err)
		}
	}
}

func TestRemoveCommandRequiresMentionAndUsers(t *testing.T) {
	c := setupHandlers(t)
	_ = Storage.AddMention(7, "team")
	user := &model.User{Username: "bob"}
	Storage.AddUserToMention(7, "team", user)
	Storage.AddUserToMention(7, model.MentionEveryoneName, user)
	c.message = testMessage("/remove @bob team")
	c.message.Entities[0].Length = 7
	c.message.Entities = append(c.message.Entities, tele.MessageEntity{Type: tele.EntityMention, Offset: 8, Length: 4})
	if err := HandleRemoveCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "send", "Removed @bob from team")
	team, _ := Storage.GetMentionUsers(7, "team")
	everyone, _ := Storage.GetMentionUsers(7, model.MentionEveryoneName)
	if len(team) != 0 || len(everyone) != 1 {
		t.Fatalf("team=%v everyone=%v", team, everyone)
	}
	c.message = testMessage("/remove @bob")
	c.message.Entities[0].Length = 7
	c.message.Entities = append(c.message.Entities, tele.MessageEntity{Type: tele.EntityMention, Offset: 8, Length: 4})
	if err := HandleRemoveCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "send", "Usage: /remove <mention> @user1 @user2")
}
