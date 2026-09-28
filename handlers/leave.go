package handlers

import (
	"fmt"
	"strings"

	"fm/pingbot/model"
	tele "gopkg.in/telebot.v3"
)

const LEAVE_COMMAND_NAME = "leave"

func HandleLeaveCommand(ctx tele.Context) error {
	names := parseMentionNames(ctx.Message().Payload, Storage.GetChatMentions(ctx.Chat().ID))
	if len(names) == 0 {
		return replyWithMentionKeyboard(ctx, "Please choose which mention to leave", LEAVE_COMMAND_NAME)
	}
	if len(names) == 1 {
		return leaveMention(ctx, getSenderUser(ctx), names[0])
	}
	results := []string{}
	for _, name := range names {
		removed, err := Storage.RemoveUserFromMention(ctx.Chat().ID, name, getSenderUser(ctx))
		if err != nil {
			results = append(results, mapStorageErrorToBotError(err, name))
		} else if removed {
			results = append(results, fmt.Sprintf("Left %s", name))
		} else {
			results = append(results, fmt.Sprintf("Not in %s", name))
		}
	}
	return ctx.Send(strings.Join(results, "\n"))
}

func handleLeaveCallback(ctx tele.Context, arguments map[string]string) error {
	mentionName := arguments[MENTION_ARGUMENT_NAME]
	if mentionName == "" {
		return ctx.EditOrReply("Unknown mention")
	}
	return leaveMention(ctx, getCallbackUser(ctx), mentionName)
}

func leaveMention(ctx tele.Context, user *model.User, mentionName string) error {
	removed, err := Storage.RemoveUserFromMention(ctx.Chat().ID, mentionName, user)
	if err != nil {
		return ctx.EditOrReply(mapStorageErrorToBotError(err, mentionName))
	}
	if !removed {
		return ctx.EditOrReply(fmt.Sprintf("You are not in mention %s", mentionName))
	}
	return ctx.EditOrReply(fmt.Sprintf("Successfully left mention %s", mentionName))
}
