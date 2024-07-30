package handlers

import (
	"strings"

	tele "gopkg.in/telebot.v3"
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
