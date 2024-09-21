package handlers

import tele "gopkg.in/telebot.v3"

const (
	CANCEL_COMMAND_NAME          = "cancel"
	CANCEL_COMMAND_TEXT_TEMPLATE = "cancel /%s command"
)

func handleCancelCallback(ctx tele.Context, _ map[string]string) error {
	return ctx.Edit("Request cancelled")
}
