package handlers

import (
	"fm/pingbot/model"
	"fmt"
	"strings"

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

func getMentionUsersString(users []*model.User) string {
	var builder strings.Builder

	for _, user := range users {
		fmt.Fprintf(&builder, "%v", getUserMention(user))
		fmt.Fprintf(&builder, " ")
	}

	return strings.Trim(builder.String(), " ")
}

func getUserMention(user *model.User) string {
	if user.ID != 0 {
		return fmt.Sprintf("[%v](tg://user?id=%v)", user.FirstName, user.ID)
	} else {
		return fmt.Sprintf("@%v", user.Username)
	}
}

func getSenderUser(ctx tele.Context) *model.User {
	return createUser(ctx.Message().Sender.ID, ctx.Message().Sender.Username, ctx.Message().Sender.FirstName)
}
