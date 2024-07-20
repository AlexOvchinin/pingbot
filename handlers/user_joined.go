package handlers

import (
	"fm/pingbot/model"

	tele "gopkg.in/telebot.v3"
)

func HandleUserJoined(ctx tele.Context) error {
	userJoined := ctx.Message().UserJoined
	if userJoined.IsBot {
		return nil
	}

	user := createUser(userJoined.ID, userJoined.Username, userJoined.FirstName)
	Storage.AddUserToMention(ctx.Chat().ID, model.MentionEveryoneName, user)
	return nil
}
