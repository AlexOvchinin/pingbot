package handlers

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

const (
// Error
)

func HandleJoinCommand(ctx tele.Context) error {
	message := ctx.Message()
	if len(message.Entities) > 1 {
		return ctx.Send(ErrorReplyCreateMention)
	}

	mentionName := strings.TrimSpace(message.Payload)

	if len(mentionName) > 0 {
		return addSenderToMention(ctx, mentionName)
	} else {
		// run reply flow
	}

	return nil
}

func addSenderToMention(ctx tele.Context, mentionName string) error {
	user := getSenderUser(ctx)
	addResult := Storage.AddUserToMention(ctx.Chat().ID, mentionName, user)
	if addResult != nil {
		return ctx.Send(fmt.Sprintf("Failed to add user %v to mention %v", getUserMention(user), mentionName), tele.ModeMarkdownV2)
	}
	return ctx.Send(fmt.Sprintf("Sucessfully added user %v to mention %v", getUserMention(user), mentionName), tele.ModeMarkdownV2)
}
