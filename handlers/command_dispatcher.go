package handlers

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

const (
	COMMAND_ARGUMENT = "command"
)

func dispatch(ctx tele.Context, argments map[string]string) error {
	command, ok := argments[COMMAND_ARGUMENT]
	if !ok {
		return ctx.EditOrReply("Could not handle command, try again later")
	}

	switch command {
	case JOIN_COMMAND_NAME:
		return handleJoinCallback(ctx, argments)
	case LEAVE_COMMAND_NAME:
		return handleLeaveCallback(ctx, argments)
	case DELETE_COMMAND_NAME:
		return handleDeleteCallback(ctx, argments)
	case CONFIRM_DELETE_COMMAND_NAME:
		return handleConfirmDeleteCallback(ctx, argments)
	case CREATE_AND_CONTINUE_COMMAND_NAME:
		return handleCreateAndContinueCallback(ctx, argments)
	case CREATE_MENTIONS_COMMAND_NAME:
		return handleCreateMentionsCallback(ctx, argments)
	case MENTION_COMMAND_NAME:
		return handleMentionCallback(ctx, argments)
	case CANCEL_COMMAND_NAME:
		return handleCancelCallback(ctx, argments)
	default:
		return ctx.EditOrReply("Could not handle command, try again later")
	}
}

func buildCommandString(commandName string, argments map[string]string) string {
	pairs := []string{}
	pairs = append(pairs, fmt.Sprintf("%v=%v", COMMAND_ARGUMENT, commandName))

	for key, value := range argments {
		pairs = append(pairs, fmt.Sprintf("%v=%v", key, value))
	}

	return strings.Join(pairs, "&")
}
