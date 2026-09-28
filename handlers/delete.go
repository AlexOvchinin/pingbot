package handlers

import (
	"fmt"
	"strings"

	"fm/pingbot/model"
	tele "gopkg.in/telebot.v3"
)

const DELETE_COMMAND_NAME = "delete"
const CONFIRM_DELETE_COMMAND_NAME = "delete_confirm"

func HandleDeleteCommand(ctx tele.Context) error {
	name := strings.TrimSpace(ctx.Message().Payload)
	if name != "" && name != model.MentionEveryoneName && Storage.IsMentionExists(ctx.Chat().ID, name) {
		return confirmDeleteMention(ctx, name)
	}
	return chooseMentionToDelete(ctx, name)
}

func chooseMentionToDelete(ctx tele.Context, unmatched string) error {
	names := []string{}
	for _, name := range Storage.GetChatMentions(ctx.Chat().ID) {
		if name != model.MentionEveryoneName {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return ctx.EditOrReply("No mentions available to delete")
	}
	keyboard := buildReplyInlineKeyboard(sortMentions(names), func(name string) string {
		return buildCommandString(DELETE_COMMAND_NAME, map[string]string{MENTION_ARGUMENT_NAME: name})
	})
	markup := &tele.ReplyMarkup{InlineKeyboard: addCancelButton(keyboard, DELETE_COMMAND_NAME)}
	text := "Choose a mention to delete"
	if unmatched != "" {
		text = fmt.Sprintf("Mention %s not found or cannot be deleted. Choose a mention to delete", unmatched)
	}
	return ctx.EditOrReply(text, markup)
}

func handleDeleteCallback(ctx tele.Context, arguments map[string]string) error {
	if !originalDeleteSender(ctx) {
		return ctx.RespondAlert("Only the sender of /delete can delete a mention")
	}
	name := arguments[MENTION_ARGUMENT_NAME]
	if name == model.MentionEveryoneName || !Storage.IsMentionExists(ctx.Chat().ID, name) {
		return chooseMentionToDelete(ctx, name)
	}
	return confirmDeleteMention(ctx, name)
}

func confirmDeleteMention(ctx tele.Context, name string) error {
	markup := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{{
		Text: fmt.Sprintf("Yes, delete %s", name),
		Data: buildCommandString(CONFIRM_DELETE_COMMAND_NAME, map[string]string{MENTION_ARGUMENT_NAME: name}),
	}}}}
	markup.InlineKeyboard = addCancelButton(markup.InlineKeyboard, DELETE_COMMAND_NAME)
	return ctx.EditOrReply(fmt.Sprintf("Delete mention %s and remove all its users?", name), markup)
}

func handleConfirmDeleteCallback(ctx tele.Context, arguments map[string]string) error {
	if !originalDeleteSender(ctx) {
		return ctx.RespondAlert("Only the sender of /delete can delete a mention")
	}
	name := arguments[MENTION_ARGUMENT_NAME]
	if name == "" {
		return ctx.Edit("Unknown mention")
	}
	if err := Storage.DeleteMention(ctx.Chat().ID, name); err != nil {
		return ctx.Edit(mapStorageErrorToBotError(err, name))
	}
	return ctx.Edit(fmt.Sprintf("Mention %s deleted", name))
}

func originalDeleteSender(ctx tele.Context) bool {
	callback := ctx.Callback()
	if callback == nil || callback.Sender == nil || callback.Message == nil ||
		callback.Message.ReplyTo == nil || callback.Message.ReplyTo.Sender == nil {
		return false
	}
	original := callback.Message.ReplyTo
	command := strings.Fields(original.Text)
	return len(command) > 0 && (command[0] == "/delete" || strings.HasPrefix(command[0], "/delete@")) &&
		original.Sender.ID == callback.Sender.ID
}
