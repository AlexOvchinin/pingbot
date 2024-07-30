package handlers

import (
	"strings"

	tele "gopkg.in/telebot.v3"
)

const (
	COMMAND_ARGUMENT = "command"
)

func OnCallback(ctx tele.Context) error {
	callback := ctx.Callback()
	arguments := make(map[string]string)
	keyValuePairs := strings.Split(callback.Data, "&")
	for _, keyValuePair := range keyValuePairs {
		pair := strings.Split(keyValuePair, "=")
		if len(pair) > 2 {
			continue
		}

		if len(pair) == 2 {
			arguments[pair[0]] = pair[1]
		} else {
			arguments[keyValuePair] = ""
		}
	}
	return dispatch(ctx, arguments)
}

func dispatch(ctx tele.Context, argments map[string]string) error {
	command, ok := argments[COMMAND_ARGUMENT]
	if !ok {
		return ctx.EditOrReply("Could not handle command, try again later")
	}

	switch command {
	case JOIN_COMMAND_NAME:
		return handleJoinCallback(ctx, argments)
	default:
		return ctx.EditOrReply("Could not handle command, try again later")
	}
}
