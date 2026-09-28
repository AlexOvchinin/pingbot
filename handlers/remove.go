package handlers

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

// /remove requires a mention and at least one tagged user; it never removes
// somebody from every mention or guesses a user from plain text.
func HandleRemoveCommand(ctx tele.Context) error {
	name, users := parseUsersAndMention(ctx.Message())
	if name == "" || len(users) == 0 {
		return ctx.Send("Usage: /remove <mention> @user1 @user2")
	}
	removed, err := Storage.RemoveUsersFromMention(ctx.Chat().ID, name, users)
	if err != nil {
		return ctx.Send(mapStorageErrorToBotError(err, name))
	}
	if len(removed) == 0 {
		return ctx.Send(fmt.Sprintf("None of the listed users belong to %s", name))
	}
	return ctx.Send(fmt.Sprintf("Removed %s from %s", getMentionUsersString(removed), name), tele.ModeHTML)
}
