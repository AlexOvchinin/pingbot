package model

import (
	"errors"
	"fmt"
)

type ChatMention struct {
	Name  string
	Users []*User
}

type Chat struct {
	ID       int64
	Mentions []*ChatMention
}

type ChatStorage struct {
	chats        []*Chat
	chatIndex    map[int64]*Chat
	mentionIndex map[string]*ChatMention
	dataPath     string
}

const (
	MentionEveryoneName string = "everyone"
)

const (
	ErrorUnknownMention string = "unknown-mention"
)

func NewChatStorage(dataPath string) *ChatStorage {
	result := &ChatStorage{
		chats:        make([]*Chat, 0),
		chatIndex:    make(map[int64]*Chat),
		mentionIndex: make(map[string]*ChatMention),
		dataPath:     dataPath,
	}
	result.load()
	return result
}

func (cs *ChatStorage) AddUserToMention(chatId int64, mentionName string, user *User) error {
	mention := cs.getMention(chatId, mentionName)
	if mention == nil {
		return errors.New(ErrorUnknownMention)
	}
	cs.addUser(mention, user)
	go cs.save()
	//modify under mutex
	return nil
}

func (cs *ChatStorage) AddUsersToMention(chatId int64, mentionName string, users []*User) []*User {
	mention := cs.getMention(chatId, mentionName)
	if mention == nil {
		return []*User{}
	}

	result := []*User{}

	for _, user := range users {
		if cs.addUser(mention, user) {
			result = append(result, user)
		}
	}

	go cs.save()

	return result
}

func (cs *ChatStorage) addUser(mention *ChatMention, user *User) bool {
	//modify under mutex
	newUsers := AddUser(user, mention.Users)
	if len(mention.Users) != len(newUsers) {
		mention.Users = newUsers
		return true
	}

	return false
}

func (cs *ChatStorage) RemoveUser(chatId int64, user *User) {
	chat, ok := cs.chatIndex[chatId]
	if !ok {
		return
	}

	for _, mention := range chat.Mentions {
		mention.Users = RemoveUser(user, mention.Users)
	}

	go cs.save()
}

func (cs *ChatStorage) GetMentionUsers(chatId int64, mentionName string) ([]*User, error) {
	mention := cs.mentionIndex[getMentionKey(chatId, mentionName)]
	if mention == nil {
		return nil, errors.New(ErrorUnknownMention)
	}

	return mention.Users, nil
}

func (cs *ChatStorage) getMention(chatId int64, mentionName string) *ChatMention {
	_, ok := cs.chatIndex[chatId]
	if !ok {
		cs.createChat(chatId)
	}

	return cs.mentionIndex[getMentionKey(chatId, mentionName)]
}

func (cs *ChatStorage) createChat(id int64) *Chat {
	// TODO: use monitor with double checking
	// TODO: update chat and mention indices
	chat := &Chat{
		ID: id,
		Mentions: []*ChatMention{
			createMention(MentionEveryoneName),
		},
	}
	cs.chats = append(cs.chats, chat)
	cs.chatIndex[id] = chat
	for _, mention := range chat.Mentions {
		cs.mentionIndex[getMentionKey(id, mention.Name)] = mention
	}

	go cs.save()

	return chat
}

func createMention(name string) *ChatMention {
	return &ChatMention{
		Name:  name,
		Users: make([]*User, 0),
	}
}

func getMentionKey(chatId int64, mentionName string) string {
	return fmt.Sprintf("%v:%v", chatId, mentionName)
}

func (cs *ChatStorage) ChangeChatId(oldId int64, newId int64) {
	chat := cs.chatIndex[oldId]
	chat.ID = newId

	cs.chatIndex[newId] = chat
	delete(cs.chatIndex, oldId)

	for _, mention := range chat.Mentions {
		newMentionKey := getMentionKey(newId, mention.Name)
		cs.mentionIndex[newMentionKey] = mention

		oldMentionKey := getMentionKey(oldId, mention.Name)
		delete(cs.mentionIndex, oldMentionKey)
	}

	go cs.save()
}

func (cs *ChatStorage) switchChats(chats []*Chat) {
	//add global mutex
	cs.chats = nil
	cs.chatIndex = make(map[int64]*Chat)
	cs.mentionIndex = make(map[string]*ChatMention)

	for _, chat := range chats {
		cs.addChat(chat)
	}
}

func (cs *ChatStorage) addChat(chat *Chat) {
	if cs.chatIndex[chat.ID] != nil {
		return
	}

	cs.chats = append(cs.chats, chat)
	cs.chatIndex[chat.ID] = chat
	for _, mention := range chat.Mentions {
		cs.mentionIndex[getMentionKey(chat.ID, mention.Name)] = mention
	}
}
