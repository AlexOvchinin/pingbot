package handlers

import (
	"strings"

	tele "gopkg.in/telebot.v3"
)

func HandleText(ctx tele.Context) error {
	payload := strings.TrimSpace(ctx.Message().Text)
	if len(payload) > 1 {
		tryMention(ctx, getSenderUser(ctx), payload[1:])
	}

	return nil
}
