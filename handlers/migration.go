package handlers

import tele "gopkg.in/telebot.v3"

func HandleMigration(ctx tele.Context) error {
	oldId := ctx.Message().MigrateFrom
	newId := ctx.Message().MigrateTo
	Storage.ChangeChatId(oldId, newId)
	return nil
}
