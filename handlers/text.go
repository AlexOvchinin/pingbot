package handlers

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func HandleText(ctx tele.Context) error {
	possibleMentions := []string{}

	for _, entity := range ctx.Message().Entities {
		switch entity.Type {
		case tele.EntityCommand:
			possibleMentions = append(possibleMentions, ctx.Message().EntityText(entity))
		}
	}

	textWithoutMentions := ctx.Message().Text

	for _, mention := range possibleMentions {
		fragments := strings.Split(textWithoutMentions, mention)
		fragments = mapStrings(fragments, func(str string) string { return strings.TrimSpace(str) })
		fragments = filterStrings(fragments, func(str string) bool { return len(str) > 0 })
		if len(fragments) > 0 {
			textWithoutMentions = strings.Join(fragments, " ")
		} else {
			textWithoutMentions = ""
		}
	}

	for _, mention := range possibleMentions {
		if len(mention) > 1 {
			error := tryMention(ctx, getSenderUser(ctx), mention[1:], textWithoutMentions)
			if error != nil {
				fmt.Println(error)
			}
		}
	}

	return nil
}

func mapStrings(strings []string, fn func(string) string) []string {
	result := []string{}
	for _, str := range strings {
		result = append(result, fn(str))
	}
	return result
}

func filterStrings(strings []string, fn func(string) bool) []string {
	result := []string{}
	for _, str := range strings {
		if fn(str) {
			result = append(result, str)
		}
	}
	return result
}
