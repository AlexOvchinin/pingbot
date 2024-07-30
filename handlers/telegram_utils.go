package handlers

import (
	tele "gopkg.in/telebot.v3"
)

func replyWithMentionKeyboard(ctx tele.Context, text string, commandName string) error {
	mentionNames := Storage.GetChatMentions(ctx.Chat().ID)
	markup := tele.ReplyMarkup{}
	markup.InlineKeyboard = buildReplyInlineKeyboard(mentionNames, func(value string) string {
		return buildCommandString(commandName, map[string]string{
			MENTION_ARGUMENT_NAME: value,
		})
	})
	return ctx.EditOrReply(text, &markup)
}

func buildReplyInlineKeyboard(values []string, callbackBuilder func(string) string) [][]tele.InlineButton {
	result := [][]tele.InlineButton{}

	rowNumber := len(values)/2 + len(values)%2

	for i := 0; i < rowNumber; i++ {
		result = append(result, []tele.InlineButton{})
	}

	for i, value := range values {
		row := i / 2
		result[row] = append(result[row], tele.InlineButton{
			Text: value,
			Data: callbackBuilder(value),
		})
	}

	return result
}
