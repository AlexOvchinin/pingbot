package handlers

import (
	"fm/pingbot/model"
	"fmt"
)

func mapStorageErrorToBotError(e error, mentionName string) string {
	switch e.Error() {
	case model.ErrorDuplicateMention:
		return fmt.Sprintf("Mention %v already exists", mentionName)
	case model.ErrorUnknownMention:
		return fmt.Sprintf("Mention %v not found. Add it with /create_mention command", mentionName)
	}
	return "Unknown error has happened"
}
