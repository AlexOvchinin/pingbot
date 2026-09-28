package handlers

import (
	"strings"

	tele "gopkg.in/telebot.v3"
)

func OnCallback(ctx tele.Context) error {
	callback := ctx.Callback()
	// Every inline keyboard is sent as a reply to the command that created it.
	// Do not let other chat members act on (or cancel) someone else's request.
	if callback == nil || callback.Sender == nil || callback.Message == nil ||
		callback.Message.ReplyTo == nil || callback.Message.ReplyTo.Sender == nil ||
		callback.Message.ReplyTo.Sender.ID != callback.Sender.ID {
		return ctx.RespondAlert("Only the sender of the original command can use this keyboard")
	}
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
