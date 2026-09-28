package handlers

import (
	"strings"
	"testing"

	"fm/pingbot/model"
	tele "gopkg.in/telebot.v3"
)

func TestDeleteMentionConfirmation(t *testing.T) {
	c := setupHandlers(t)
	if err := Storage.AddMention(7, "team"); err != nil {
		t.Fatal(err)
	}
	c.message = testMessage("/delete team")
	c.message.Payload = "team"
	c.message.Chat = c.chat
	if err := HandleDeleteCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "Delete mention team and remove all its users?")
	if !Storage.IsMentionExists(7, "team") {
		t.Fatal("mention removed before confirmation")
	}
	markup := c.calls[0].opts[0].(*tele.ReplyMarkup)
	if !strings.Contains(markup.InlineKeyboard[0][0].Data, "command=delete_confirm") {
		t.Fatalf("confirmation callback: %v", markup.InlineKeyboard)
	}
	c.callback.Message.ReplyTo = c.message
	if err := handleConfirmDeleteCallback(c, map[string]string{MENTION_ARGUMENT_NAME: "team"}); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "edit", "Mention team deleted")
	if Storage.IsMentionExists(7, "team") {
		t.Fatal("mention still exists")
	}
	if !Storage.IsMentionExists(7, model.MentionEveryoneName) {
		t.Fatal("everyone was deleted")
	}
}

func TestDeleteShowsOptionsForMissingOrUnmatchedName(t *testing.T) {
	c := setupHandlers(t)
	_ = Storage.AddMention(7, "team")
	_ = Storage.AddMention(7, "ops")
	c.message = testMessage("/delete")
	c.message.Chat = c.chat
	if err := HandleDeleteCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "Choose a mention to delete")
	markup := c.calls[0].opts[0].(*tele.ReplyMarkup)
	if len(markup.InlineKeyboard) != 3 || markup.InlineKeyboard[0][0].Text != "ops" || markup.InlineKeyboard[1][0].Text != "team" {
		t.Fatalf("unexpected options: %v", markup.InlineKeyboard)
	}
	c.message.Payload = "unknown"
	if err := HandleDeleteCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "Mention unknown not found or cannot be deleted. Choose a mention to delete")
	c.callback.Message.ReplyTo = c.message
	if err := handleDeleteCallback(c, map[string]string{MENTION_ARGUMENT_NAME: "team"}); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "Delete mention team and remove all its users?")
}

func TestDeleteRejectsOtherSenderAndEveryone(t *testing.T) {
	c := setupHandlers(t)
	_ = Storage.AddMention(7, "team")
	c.message = testMessage("/delete team")
	c.message.Chat = c.chat
	c.callback.Message.ReplyTo = c.message
	c.callback.Sender = &tele.User{ID: 999}
	c.callback.Data = "command=delete_confirm&mention=team"
	if err := OnCallback(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "alert", "Only the sender of the original command can use this keyboard")
	if !Storage.IsMentionExists(7, "team") {
		t.Fatal("unauthorized deletion")
	}
	c.callback.Sender = c.message.Sender
	if err := handleConfirmDeleteCallback(c, map[string]string{MENTION_ARGUMENT_NAME: model.MentionEveryoneName}); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "edit", "The default everyone mention cannot be deleted")
}
