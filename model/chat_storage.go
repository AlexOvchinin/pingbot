package model

import (
	"errors"
	"fmt"
	"sync"
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
	mu       sync.RWMutex
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
	ErrorProtectedMention                string = "protected-mention"
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
	cs.mu.Lock()
	defer cs.mu.Unlock()
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

	cs.save()
	return nil
}

// DeleteMention removes one named mention and all of its memberships.
// The default everyone mention is a permanent part of each chat.
func (cs *ChatStorage) DeleteMention(chatId int64, mentionName string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if mentionName == MentionEveryoneName {
		return errors.New(ErrorProtectedMention)
	}
	chat := cs.chats[chatId]
	if chat == nil || cs.mentions[getMentionKey(chatId, mentionName)] == nil {
		return errors.New(ErrorUnknownMention)
	}
	for i, mention := range chat.Mentions {
		if mention.Name == mentionName {
			chat.Mentions = append(chat.Mentions[:i], chat.Mentions[i+1:]...)
			break
		}
	}
	delete(cs.mentions, getMentionKey(chatId, mentionName))
	cs.save()
	return nil
}

func (cs *ChatStorage) AddUserToMention(chatId int64, mentionName string, user *User) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	mention := cs.getMention(chatId, mentionName)
	if mention == nil {
		return errors.New(ErrorUnknownMention)
	}
	cs.addUser(mention, user)
	cs.save()
	return nil
}

func (cs *ChatStorage) AddUsersToMention(chatId int64, mentionName string, users []*User) []*User {
	cs.mu.Lock()
	defer cs.mu.Unlock()
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

	cs.save()

	return result
}

func (cs *ChatStorage) addUser(mention *ChatMention, user *User) bool {
	// Retain our own value: callers may reuse or edit their input after return.
	if user == nil {
		return false
	}
	newUsers := mention.Users
	if !ContainsUser(user, mention.Users) {
		copyUser := *user
		newUsers = append(newUsers, &copyUser)
	}
	if len(mention.Users) != len(newUsers) {
		mention.Users = newUsers
		return true
	}

	return false
}

func (cs *ChatStorage) RemoveUser(chatId int64, user *User) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	chat, ok := cs.chats[chatId]
	if !ok {
		return
	}

	for _, mention := range chat.Mentions {
		mention.Users = RemoveUser(user, mention.Users)
	}

	cs.save()
}

// RemoveUserFromMention removes a user only from the named mention.
func (cs *ChatStorage) RemoveUserFromMention(chatId int64, mentionName string, user *User) (bool, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	mention := cs.getMention(chatId, mentionName)
	if mention == nil {
		return false, errors.New(ErrorUnknownMention)
	}
	users := RemoveUser(user, mention.Users)
	if len(users) == len(mention.Users) {
		return false, nil
	}
	mention.Users = users
	cs.save()
	return true, nil
}

// RemoveUsersFromMention removes only users present in the named mention.
func (cs *ChatStorage) RemoveUsersFromMention(chatId int64, mentionName string, users []*User) ([]*User, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	mention := cs.getMention(chatId, mentionName)
	if mention == nil {
		return nil, errors.New(ErrorUnknownMention)
	}
	removed := []*User{}
	for _, user := range users {
		if ContainsUser(user, mention.Users) && !ContainsUser(user, removed) {
			mention.Users = RemoveUser(user, mention.Users)
			removed = append(removed, user)
		}
	}
	if len(removed) > 0 {
		cs.save()
	}
	return removed, nil
}

func (cs *ChatStorage) IsMentionExists(chatId int64, mentionName string) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	_, ok := cs.mentions[getMentionKey(chatId, mentionName)]
	return ok
}

func (cs *ChatStorage) GetMentionUsers(chatId int64, mentionName string) ([]*User, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	mention := cs.mentions[getMentionKey(chatId, mentionName)]
	if mention == nil {
		return nil, errors.New(ErrorUnknownMention)
	}

	users := make([]*User, len(mention.Users))
	for i, user := range mention.Users {
		if user != nil {
			copyUser := *user
			users[i] = &copyUser
		}
	}
	return users, nil
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

	cs.save()

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
	cs.mu.Lock()
	defer cs.mu.Unlock()
	chat := cs.chats[oldId]
	if chat == nil || oldId == newId {
		return
	}
	chat.ID = newId

	cs.chats[newId] = chat
	delete(cs.chats, oldId)

	for _, mention := range chat.Mentions {
		newMentionKey := getMentionKey(newId, mention.Name)
		cs.mentions[newMentionKey] = mention

		oldMentionKey := getMentionKey(oldId, mention.Name)
		delete(cs.mentions, oldMentionKey)
	}

	cs.save()
}

func (cs *ChatStorage) switchChats(chats []*Chat) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
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
	cs.mu.RLock()
	defer cs.mu.RUnlock()
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
