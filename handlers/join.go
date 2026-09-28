package handlers

import (
	"fm/pingbot/model"
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

const (
	// commmand
	JOIN_COMMAND_NAME = "join"
)

func HandleJoinCommand(ctx tele.Context) error {
	message := ctx.Message()
	names := parseMentionNames(message.Payload, Storage.GetChatMentions(ctx.Chat().ID))
	if len(names) > 0 {
		missing := []string{}
		for _, name := range names {
			if name != model.MentionEveryoneName && !Storage.IsMentionExists(ctx.Chat().ID, name) {
				missing = append(missing, name)
			}
		}
		if len(missing) > 0 {
			if len(names) == 1 {
				return suggestCreatingMention(ctx, missing[0], JOIN_COMMAND_NAME)
			}
			return suggestCreatingMentions(ctx, missing)
		}
		if len(names) == 1 {
			return addSenderToMention(ctx, getSenderUser(ctx), names[0])
		}
		return joinMentions(ctx, getSenderUser(ctx), names)
	} else {
		return replyWithMentionKeyboard(ctx, "Please choose which mention to join", JOIN_COMMAND_NAME)
	}
}

func joinMentions(ctx tele.Context, user *model.User, names []string) error {
	for _, name := range names {
		if err := Storage.AddUserToMention(ctx.Chat().ID, name, user); err != nil {
			return ctx.EditOrReply(mapStorageErrorToBotError(err, name))
		}
	}
	return ctx.EditOrReply(fmt.Sprintf("Successfully joined %s", strings.Join(names, ", ")))
}

func handleJoinCallback(ctx tele.Context, arguments map[string]string) error {
	mentionName, ok := arguments[MENTION_ARGUMENT_NAME]
	if !ok {
		return ctx.Send("Unknown mention")
	}

	return addSenderToMention(ctx, getCallbackUser(ctx), mentionName)
}

func addSenderToMention(ctx tele.Context, user *model.User, mentionName string) error {
	addResult := Storage.AddUserToMention(ctx.Chat().ID, mentionName, user)
	if addResult != nil {
		return ctx.EditOrReply(fmt.Sprintf("Failed to add user %v to mention %v", getUserMentionHtml(user), mentionName), tele.ModeHTML)
	}
	return ctx.EditOrReply(fmt.Sprintf("Sucessfully added user %v to mention %v", getUserMentionHtml(user), mentionName), tele.ModeHTML)
}
