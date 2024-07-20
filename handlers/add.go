package handlers

import (
	"fm/pingbot/model"
	"fmt"

	tele "gopkg.in/telebot.v3"
)

var Storage *model.ChatStorage

func HandleAddCommand(ctx tele.Context) error {
	users := make([]*model.User, 0)

	for _, entity := range ctx.Message().Entities {
		switch entity.Type {
		case tele.EntityMention:
			entityMention := ctx.Message().EntityText(entity)
			if len(entityMention) > 0 {
				users = append(users, createUserByUsername(entityMention[1:]))
			}
		case tele.EntityTMention:
			users = append(users, createUserByIdAndName(entity.User.ID, entity.User.FirstName))
		}
	}

	if len(users) == 0 {
		users = append(users, getSenderUser(ctx))
	}

	addedUsers := Storage.AddUsersToMention(ctx.Chat().ID, model.MentionEveryoneName, users)

	if len(addedUsers) == 0 {
		return ctx.Send("All transferred users already belong to the group")
	}

	return ctx.Send(fmt.Sprintf("Added %v to group %v", getMentionUsersString(addedUsers), model.MentionEveryoneName), tele.ModeMarkdownV2)
}
