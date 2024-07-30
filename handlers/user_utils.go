package handlers

import (
	"fm/pingbot/model"

	tele "gopkg.in/telebot.v3"
)

func createUserByUsername(username string) *model.User {
	return &model.User{
		Username: username,
	}
}

func createUserByIdAndName(id int64, firstName string) *model.User {
	return &model.User{
		ID:        id,
		FirstName: firstName,
	}
}

func createUser(id int64, username string, firstName string) *model.User {
	return &model.User{
		ID:        id,
		Username:  username,
		FirstName: firstName,
	}
}

func getSenderUser(ctx tele.Context) *model.User {
	return createUser(ctx.Message().Sender.ID, ctx.Message().Sender.Username, ctx.Message().Sender.FirstName)
}

func getCallbackUser(ctx tele.Context) *model.User {
	return createUser(ctx.Callback().Sender.ID, ctx.Callback().Sender.Username, ctx.Callback().Sender.FirstName)
}
