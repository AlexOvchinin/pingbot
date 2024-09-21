package handlers

import (
	tele "gopkg.in/telebot.v3"
)

func HandleText(ctx tele.Context) error {
	possibleMentions := []string{}

	for _, entity := range ctx.Message().Entities {
		switch entity.Type {
		case tele.EntityCommand:
			possibleMentions = append(possibleMentions, ctx.Message().EntityText(entity))
		}
	}

	for _, mention := range possibleMentions {
		if len(mention) > 1 {
			tryMention(ctx, getSenderUser(ctx), mention[1:])
		}
	}

	return nil
}
