package handlers

import (
	"strings"

	tele "gopkg.in/telebot.v3"
)

const (
	MENTION_COMMAND_NAME = "mention"
)

func HandleMention(ctx tele.Context) error {
	message := ctx.Message()
	if len(message.Entities) > 1 {
		return ctx.Send(ErrorReplyCreateMention)
	}

	mentionName := strings.TrimSpace(message.Payload)
	if len(mentionName) > 0 {
		return sendMention(ctx, getSenderUser(ctx), mentionName)
	} else {
		return replyWithMentionKeyboard(ctx, "Please choose who to mention", MENTION_COMMAND_NAME)
	}
}

func handleMentionCallback(ctx tele.Context, arguments map[string]string) error {
	mentionName, ok := arguments[MENTION_ARGUMENT_NAME]
	if !ok {
		return ctx.EditOrReply("Unknown mention")
	}

	ctx.Delete()
	return sendMention(ctx, getCallbackUser(ctx), mentionName)
}
