package model

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type PersistentChatStorage struct {
	Chats []*Chat
}

// save is called with cs.mu held. Holding the lock through the write ensures
// each completed mutation is persisted in order before another can begin.
func (cs *ChatStorage) save() {
	chats := []*Chat{}
	for _, chat := range cs.chats {
		chats = append(chats, chat)
	}

	storage := &PersistentChatStorage{
		Chats: chats,
	}

	f, err := os.Create(cs.dataPath)
	if err != nil {
		log.Println(err)
		return
	}

	defer f.Close()

	marshaledChats, err := json.Marshal(storage)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = f.WriteString(string(marshaledChats))
	if err != nil {
		log.Println(err)
	}
}

func (cs *ChatStorage) load() {
	marshaledChats, err := os.ReadFile(cs.dataPath)
	if err != nil {
		log.Println(err)
		fmt.Printf("No chats to load")
		return
	}

	var storage PersistentChatStorage

	err = json.Unmarshal(marshaledChats, &storage)
	if err != nil {
		log.Fatal(err)
	}

	cs.switchChats(storage.Chats)
}
