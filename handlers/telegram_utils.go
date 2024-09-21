package handlers

import (
	"fm/pingbot/model"
	"fmt"
	"sort"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func replyWithMentionKeyboard(ctx tele.Context, text string, commandName string) error {
	mentionNames := Storage.GetChatMentions(ctx.Chat().ID)

	mentionsKeyboard := buildReplyInlineKeyboard(sortMentions(mentionNames), func(value string) string {
		return buildCommandString(commandName, map[string]string{
			MENTION_ARGUMENT_NAME: value,
		})
	})

	mentionsKeyboard = addCancelButton(mentionsKeyboard, commandName)

	markup := tele.ReplyMarkup{}
	markup.InlineKeyboard = mentionsKeyboard
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

func addCancelButton(keyboard [][]tele.InlineButton, commandName string) [][]tele.InlineButton {
	return append(keyboard, []tele.InlineButton{
		{
			Text: fmt.Sprintf(CANCEL_COMMAND_TEXT_TEMPLATE, commandName),
			Data: buildCommandString(CANCEL_COMMAND_NAME, map[string]string{}),
		},
	})
}

func buildReplyInlineKeyboard(values []string, callbackBuilder func(string) string) [][]tele.InlineButton {
	result := [][]tele.InlineButton{}

	for i := 0; i < len(values); i++ {
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
