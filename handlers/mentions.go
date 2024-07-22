package handlers

import (
	"fm/pingbot/model"
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func mention(ctx tele.Context, mentionName string) error {
	users, e := Storage.GetMentionUsers(ctx.Chat().ID, mentionName)
	if e != nil {
		return ctx.Send(mapStorageErrorToBotError(e, mentionName))
	}

	users = model.RemoveUser(getSenderUser(ctx), users)

	mentionMessage := getMentionUsersString(users)
	if len(mentionMessage) == 0 {
		return ctx.Send("Noone to mention. Please use /add to add users to mention manually")
	}

	return ctx.Send(mentionMessage, tele.ModeMarkdownV2)
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
	if len(user.Username) > 0 {
		return fmt.Sprintf("@%v", user.Username)
	} else {
		return fmt.Sprintf("[%v](tg://user?id=%v)", user.FirstName, user.ID)
	}
}
