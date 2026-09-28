package handlers

import (
	"fm/pingbot/model"
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func HandleAddCommand(ctx tele.Context) error {
	mentionName, users := parseAddMessage(ctx.Message())
	if mentionName != model.MentionEveryoneName && !Storage.IsMentionExists(ctx.Chat().ID, mentionName) {
		return suggestCreatingMention(ctx, mentionName, ADD_COMMAND_NAME)
	}
	return addUsersToMention(ctx, mentionName, users)
}

const ADD_COMMAND_NAME = "add"

func parseAddMessage(message *tele.Message) (string, []*model.User) {
	mentionName, users := parseUsersAndMention(message)
	if mentionName == "" {
		mentionName = model.MentionEveryoneName
	}
	return mentionName, users
}

func parseUsersAndMention(message *tele.Message) (string, []*model.User) {
	users := make([]*model.User, 0)
	// Remove entity text one occurrence at a time so names may appear before,
	// between, or after users, including Telegram text mentions.
	remaining := message.Text

	for _, entity := range message.Entities {
		switch entity.Type {
		case tele.EntityCommand:
			remaining = strings.Replace(remaining, message.EntityText(entity), "", 1)
		case tele.EntityMention:
			entityMention := message.EntityText(entity)
			if len(entityMention) > 0 {
				users = append(users, createUserByUsername(entityMention[1:]))
			}
			remaining = strings.Replace(remaining, entityMention, "", 1)
		case tele.EntityTMention:
			if entity.User != nil {
				users = append(users, createUserByIdAndName(entity.User.ID, entity.User.FirstName))
			}
			remaining = strings.Replace(remaining, message.EntityText(entity), "", 1)
		}
	}

	mentionName := strings.TrimSpace(remaining)
	return mentionName, users
}

func addUsersToMention(ctx tele.Context, mentionName string, users []*model.User) error {
	if len(users) == 0 {
		if ctx.Callback() != nil {
			users = append(users, getCallbackUser(ctx))
		} else {
			users = append(users, getSenderUser(ctx))
		}
	}

	addedUsers := Storage.AddUsersToMention(ctx.Chat().ID, mentionName, users)

	if len(addedUsers) == 0 {
		return ctx.Send("All transferred users already belong to the group")
	}

	return ctx.Send(fmt.Sprintf("Added %v to group %v", getMentionUsersString(addedUsers), mentionName), tele.ModeHTML)
}
