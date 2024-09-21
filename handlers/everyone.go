package handlers

import (
	"fm/pingbot/model"

	tele "gopkg.in/telebot.v3"
)

func HandleEveryoneCommand(ctx tele.Context) error {
	users, error := Storage.GetMentionUsers(ctx.Chat().ID, model.MentionEveryoneName)
	if error != nil {
		if error.Error() == model.ErrorUnknownMention {
			return ctx.Send("Unknown mention, please add group and users before trying again")
		}
	}

	users = model.RemoveUser(getSenderUser(ctx), users)

	mentionMessage := getMentionUsersString(users)
	if len(mentionMessage) == 0 {
		return ctx.Send("Noone to mention. Please use /add or to add users to mention manually or /join to join it yourself")
	}

	return ctx.Send(mentionMessage, tele.ModeMarkdownV2)
}
