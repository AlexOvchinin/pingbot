package handlers

import (
	"fm/pingbot/model"
	"fmt"
	"strings"
	"html"

	tele "gopkg.in/telebot.v3"
)

func mention(ctx tele.Context, currentUser *model.User, mentionName string, content string) error {
	users, e := Storage.GetMentionUsers(ctx.Chat().ID, mentionName)
	if e != nil {
		return ctx.Send(mapStorageErrorToBotError(e, mentionName))
	}

	users = model.RemoveUser(currentUser, users)

	usersMention := getMentionUsersString(users)
	if len(usersMention) == 0 {
		return ctx.Send(fmt.Sprintf("Noone to mention in %s. Please use /add to add users to mention manually or /join to join it yourself", mentionName))
	}

	message := fmt.Sprintf("%s calling %s", getUserName(currentUser), mentionName)
	if len(content) > 0 {
		message += ": " + html.EscapeString(content)
	}
	if len(message) > 0 {
		message += fmt.Sprintf("\n%v", usersMention)
	}

	return ctx.Send(message, tele.ModeHTML)
}

func tryMention(ctx tele.Context, currentUser *model.User, mentionName string, content string) error {
	if Storage.IsMentionExists(ctx.Chat().ID, mentionName) {
		return mention(ctx, currentUser, mentionName, content)
	}

	return nil
}

func getMentionUsersString(users []*model.User) string {
	var builder strings.Builder

	for _, user := range users {
		fmt.Fprintf(&builder, "%v", getUserMentionHtml(user))
		fmt.Fprintf(&builder, " ")
	}

	return strings.TrimSpace(builder.String())
}

func getUserMentionMarkdown(user *model.User) string {
	if len(user.Username) > 0 {
		return fmt.Sprintf("@%v", strings.Replace(user.Username, "_", "\\_", -1))
	} else {
		return fmt.Sprintf("[%v](tg://user?id=%v)", user.FirstName, user.ID)
	}
}

func getUserMentionHtml(user *model.User) string {
	return fmt.Sprintf("<a href=\"tg://user?id=%v\">%v</a>", user.ID, html.EscapeString(user.FirstName))
}

func getUserName(user *model.User) string {
	if len(user.Username) > 0 {
		return user.Username
	} else {
		return user.FirstName
	}
}
