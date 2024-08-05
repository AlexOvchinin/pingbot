package handlers

import (
	"fm/pingbot/model"
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func mention(ctx tele.Context, currentUser *model.User, mentionName string) error {
	users, e := Storage.GetMentionUsers(ctx.Chat().ID, mentionName)
	if e != nil {
		return ctx.Send(mapStorageErrorToBotError(e, mentionName))
	}

	users = model.RemoveUser(currentUser, users)

	mentionContent := getMentionUsersString(users)
	if len(mentionContent) == 0 {
		return ctx.Send("Noone to mention. Please use /add to add users to mention manually")
	}

	return ctx.Send(fmt.Sprintf("Calling %v! %v", mentionName, mentionContent), tele.ModeMarkdownV2)
}

func tryMention(ctx tele.Context, currentUser *model.User, mentionName string) error {
	if Storage.IsMentionExists(ctx.Chat().ID, mentionName) {
		return mention(ctx, currentUser, mentionName)
	}

	return nil
}

func getMentionUsersString(users []*model.User) string {
	var builder strings.Builder

	for _, user := range users {
		fmt.Fprintf(&builder, "%v", getUserMention(user))
		fmt.Fprintf(&builder, " ")
	}

	return strings.TrimSpace(builder.String())
}

func getUserMention(user *model.User) string {
	if len(user.Username) > 0 {
		return fmt.Sprintf("@%v", user.Username)
	} else {
		return fmt.Sprintf("[%v](tg://user?id=%v)", user.FirstName, user.ID)
	}
}
