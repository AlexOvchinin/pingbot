package handlers

import (
	"strings"

	tele "gopkg.in/telebot.v3"
)

func HandleMention(ctx tele.Context) error {
	message := ctx.Message()
	if len(message.Entities) > 1 {
		return ctx.Send(ErrorReplyCreateMention)
	}

	mentionName := strings.TrimSpace(message.Payload)
	if len(mentionName) > 0 {
		return mention(ctx, mentionName)
	} else {
		// run reply flow
	}
	return nil
}
