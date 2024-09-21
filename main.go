package main

import (
	"fm/pingbot/handlers"
	"fm/pingbot/model"
	"log"
	"os"
	"time"

	tele "gopkg.in/telebot.v3"
)

var storage *model.ChatStorage

func main() {
	storage = model.NewChatStorage(os.Getenv("PING_BOT_DATA_PATH"))

	token := os.Getenv("PING_BOT_TOKEN")
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	handlers.Storage = storage

	// bot commands
	b.Handle("/add", handlers.HandleAddCommand)
	b.Handle("/everyone", handlers.HandleEveryoneCommand)
	b.Handle("/join", handlers.HandleJoinCommand)
	b.Handle("/create_mention", handlers.HandleCreateMention)
	b.Handle("/mention", handlers.HandleMention)
	b.Handle("/help", handlers.HandleHelpCommand)

	// chat events
	b.Handle(tele.OnText, handlers.HandleText)
	b.Handle(tele.OnUserJoined, handlers.HandleUserJoined)
	b.Handle(tele.OnUserLeft, handlers.HandleUserLeft)
	b.Handle(tele.OnMigration, handlers.HandleMigration)
	b.Handle(tele.OnCallback, handlers.OnCallback)

	b.Start()

	log.Println("Ping Telegram Bot Started")
}
