package handlers

import tele "gopkg.in/telebot.v3"

func HandleUserLeft(ctx tele.Context) error {
	userLeft := ctx.Message().UserLeft
	if userLeft.IsBot {
		return nil
	}

	user := createUser(userLeft.ID, userLeft.Username, userLeft.FirstName)
	Storage.RemoveUser(ctx.Chat().ID, user)
	return nil
}
