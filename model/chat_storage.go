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
	chats    map[int64]*Chat
	mentions map[string]*ChatMention
	dataPath string
}

const (
	MentionEveryoneName string = "everyone"
)

const (
	ErrorUnknownMention                  string = "unknown-mention"
	ErrorDuplicateMention                string = "duplicate-mention"
	ErrorExceededMaximumNumberOfMentions string = "exceeded-maximum-number-of-mentions"
)

const (
	MAX_MENTIONS_NUMBER = 10
)

func NewChatStorage(dataPath string) *ChatStorage {
	result := &ChatStorage{
		chats:    make(map[int64]*Chat),
		mentions: make(map[string]*ChatMention),
		dataPath: dataPath,
	}
	result.load()
	return result
}

func (cs *ChatStorage) AddMention(chatId int64, mentionName string) error {
	chat := cs.getOrCreateChat(chatId)

	mention := cs.getMention(chatId, mentionName)
	if mention != nil {
		return errors.New(ErrorDuplicateMention)
	}

	if len(chat.Mentions) >= MAX_MENTIONS_NUMBER {
		return errors.New(ErrorExceededMaximumNumberOfMentions)
	}

	mention = createMention(mentionName)
	chat.Mentions = append(chat.Mentions, mention)
	cs.mentions[getMentionKey(chatId, mentionName)] = mention

	go cs.save()
	return nil
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
	chat, ok := cs.chats[chatId]
	if !ok {
		return
	}

	for _, mention := range chat.Mentions {
		mention.Users = RemoveUser(user, mention.Users)
	}

	go cs.save()
}

func (cs *ChatStorage) IsMentionExists(chatId int64, mentionName string) bool {
	_, ok := cs.mentions[getMentionKey(chatId, mentionName)]
	return ok
}

func (cs *ChatStorage) GetMentionUsers(chatId int64, mentionName string) ([]*User, error) {
	mention := cs.mentions[getMentionKey(chatId, mentionName)]
	if mention == nil {
		return nil, errors.New(ErrorUnknownMention)
	}

	return mention.Users, nil
}

func (cs *ChatStorage) getMention(chatId int64, mentionName string) *ChatMention {
	cs.getOrCreateChat(chatId)
	return cs.mentions[getMentionKey(chatId, mentionName)]
}

func (cs *ChatStorage) getOrCreateChat(chatId int64) *Chat {
	chat, ok := cs.chats[chatId]
	if !ok {
		chat = cs.createChat(chatId)
	}
	return chat
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
	cs.chats[id] = chat
	for _, mention := range chat.Mentions {
		cs.mentions[getMentionKey(id, mention.Name)] = mention
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
	chat := cs.chats[oldId]
	chat.ID = newId

	cs.chats[newId] = chat
	delete(cs.chats, oldId)

	for _, mention := range chat.Mentions {
		newMentionKey := getMentionKey(newId, mention.Name)
		cs.mentions[newMentionKey] = mention

		oldMentionKey := getMentionKey(oldId, mention.Name)
		delete(cs.mentions, oldMentionKey)
	}

	go cs.save()
}

func (cs *ChatStorage) switchChats(chats []*Chat) {
	//add global mutex
	cs.chats = make(map[int64]*Chat)
	cs.mentions = make(map[string]*ChatMention)

	for _, chat := range chats {
		cs.addChat(chat)
	}
}

func (cs *ChatStorage) addChat(chat *Chat) {
	if cs.chats[chat.ID] != nil {
		return
	}

	cs.chats[chat.ID] = chat
	for _, mention := range chat.Mentions {
		cs.mentions[getMentionKey(chat.ID, mention.Name)] = mention
	}
}

func (cs *ChatStorage) GetChatMentions(chatId int64) []string {
	chat, ok := cs.chats[chatId]
	if !ok {
		return []string{}
	}

	chatMentions := []string{}

	for _, mention := range chat.Mentions {
		chatMentions = append(chatMentions, mention.Name)
	}

	return chatMentions
}
