package handlers

import (
	"fm/pingbot/model"
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

const (
	// commmand
	JOIN_COMMAND_NAME = "join"
	// keyboard
	MAX_BUTTONS_PER_KEYBOARD_ROW = 2
)

func HandleJoinCommand(ctx tele.Context) error {
	message := ctx.Message()
	if len(message.Entities) > 1 {
		return ctx.Send(ErrorReplyCreateMention)
	}

	mentionName := strings.TrimSpace(message.Payload)

	if len(mentionName) > 0 {
		user := getSenderUser(ctx)
		return addSenderToMention(ctx, user, mentionName)
	} else {
		return replyWithMentionKeyboard(ctx, "Please choose which mention to join", JOIN_COMMAND_NAME)
	}
}

func handleJoinCallback(ctx tele.Context, arguments map[string]string) error {
	mentionName, ok := arguments[MENTION_ARGUMENT_NAME]
	if !ok {
		return ctx.Send("Unknown mention")
	}

	return addSenderToMention(ctx, getCallbackUser(ctx), mentionName)
}

func addSenderToMention(ctx tele.Context, user *model.User, mentionName string) error {
	addResult := Storage.AddUserToMention(ctx.Chat().ID, mentionName, user)
	if addResult != nil {
		return ctx.EditOrReply(fmt.Sprintf("Failed to add user %v to mention %v", getUserMention(user), mentionName), tele.ModeMarkdownV2)
	}
	return ctx.EditOrReply(fmt.Sprintf("Sucessfully added user %v to mention %v", getUserMention(user), mentionName), tele.ModeMarkdownV2)
}
