package handlers

import (
	"fm/pingbot/model"
	"sort"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func replyWithMentionKeyboard(ctx tele.Context, text string, commandName string) error {
	mentionNames := Storage.GetChatMentions(ctx.Chat().ID)
	markup := tele.ReplyMarkup{}
	markup.InlineKeyboard = buildReplyInlineKeyboard(sortMentions(mentionNames), func(value string) string {
		return buildCommandString(commandName, map[string]string{
			MENTION_ARGUMENT_NAME: value,
		})
	})
	return ctx.EditOrReply(text, &markup)
}

func sortMentions(values []string) []string {
	sortedValues := make([]string, len(values))
	copy(sortedValues, values)
	sort.Slice(sortedValues, func(i, j int) bool {
		if sortedValues[i] == model.MentionEveryoneName {
			return true
		}
		if sortedValues[j] == model.MentionEveryoneName {
			return false
		}
		return strings.ToLower(sortedValues[i]) < strings.ToLower(sortedValues[j])
	})
	return sortedValues
}

func buildReplyInlineKeyboard(values []string, callbackBuilder func(string) string) [][]tele.InlineButton {
	result := [][]tele.InlineButton{}

	rowNumber := len(values)

	for i := 0; i < rowNumber; i++ {
		result = append(result, []tele.InlineButton{})
	}

	for i, value := range values {
		result[i] = append(result[i], tele.InlineButton{
			Text: value,
			Data: callbackBuilder(value),
		})
	}

	return result
}
