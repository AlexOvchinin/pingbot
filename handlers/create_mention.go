package handlers

import (
	"fmt"
	"regexp"
	"strings"

	tele "gopkg.in/telebot.v3"
)

var re = regexp.MustCompile(`[=&]`)

const (
	ErrorReplyCreateMention = "Could not create mention. Example: /create_mention mention_name. Maximum mention length is 20 symbols"
	ErrorReplyCreateMentionForbiddenSymbols = "Could not create mention. Remove symbols & and = and try again"
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

	if re.MatchString(mentionName) {
		return ctx.Send(ErrorReplyCreateMentionForbiddenSymbols)
	}

	result := Storage.AddMention(ctx.Chat().ID, mentionName)
	if result != nil {
		return ctx.Send(mapStorageErrorToBotError(result, mentionName))
	}

	return ctx.Send(fmt.Sprintf("Mention %v was successfully created", mentionName))
}
