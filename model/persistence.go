package model

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type PersistentChatStorage struct {
	Chats []*Chat
}

var mu sync.Mutex

func (cs *ChatStorage) save() {
	chats := []*Chat{}
	for _, chat := range cs.chats {
		chats = append(chats, chat)
	}

	storage := &PersistentChatStorage{
		Chats: chats,
	}

	mu.Lock()
	defer mu.Unlock()

	f, err := os.Create(cs.dataPath)
	if err != nil {
		log.Println(err)
	}

	defer f.Close()

	marshaledChats, err := json.Marshal(storage)
	if err != nil {
		log.Println(err)
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
