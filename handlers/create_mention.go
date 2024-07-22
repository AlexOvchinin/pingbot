package handlers

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

const (
	ErrorReplyCreateMention = "Could not create mention. Example: /create_mention mention_name. Maximum mention length is 20 symbols"
	MaxMentionLength        = 20
)

func HandleCreateMention(ctx tele.Context) error {
	message := ctx.Message()
	if len(message.Entities) > 1 {
		return ctx.Send(ErrorReplyCreateMention)
	}

	mentionName := strings.TrimSpace(message.Payload)
	if len(mentionName) > MaxMentionLength || len(mentionName) < 1 {
		return ctx.Send(ErrorReplyCreateMention)
	}

	result := Storage.AddMention(ctx.Chat().ID, mentionName)
	if result != nil {
		return ctx.Send(mapStorageErrorToBotError(result, mentionName))
	}

	return ctx.Send(fmt.Sprintf("Mention %v was successfully created", mentionName))
}
