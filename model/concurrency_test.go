package model

import (
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

func TestChatStorageConcurrentMutationsPersistLatestState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chats.json")
	storage := NewChatStorage(path)
	if err := storage.AddMention(1, "team"); err != nil {
		t.Fatal(err)
	}
	const count = 20
	var wg sync.WaitGroup
	for i := 1; i <= count; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			user := &User{ID: id}
			if err := storage.AddUserToMention(1, "team", user); err != nil {
				t.Errorf("add user %d: %v", id, err)
			}
			if _, err := storage.GetMentionUsers(1, "team"); err != nil {
				t.Errorf("read users: %v", err)
			}
			storage.GetChatMentions(1)
			storage.IsMentionExists(1, "team")
		}(int64(i))
	}
	wg.Wait()
	for _, current := range []*ChatStorage{storage, NewChatStorage(path)} {
		users, err := current.GetMentionUsers(1, "team")
		if err != nil || len(users) != count {
			t.Fatalf("members = %v, err = %v; want %d", users, err, count)
		}
		for i := 1; i <= count; i++ {
			if !ContainsUser(&User{ID: int64(i)}, users) {
				t.Fatalf("missing user %d", i)
			}
		}
	}
}

func TestChatStorageConcurrentCreationIsUnique(t *testing.T) {
	storage := newTestStorage(t)
	var successes atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := storage.AddMention(1, "team"); err == nil {
				successes.Add(1)
			} else if err.Error() != ErrorDuplicateMention {
				t.Errorf("unexpected create error: %v", err)
			}
		}()
	}
	wg.Wait()
	if got := successes.Load(); got != 1 {
		t.Fatalf("successful creations = %d, want 1", got)
	}
	if got := storage.GetChatMentions(1); len(got) != 2 {
		t.Fatalf("mentions after concurrent creation = %v", got)
	}
}

func TestChatStorageReturnsIndependentUsers(t *testing.T) {
	storage := newTestStorage(t)
	user := &User{ID: 1, Username: "alice"}
	if err := storage.AddUserToMention(1, MentionEveryoneName, user); err != nil {
		t.Fatal(err)
	}
	user.Username = "changed by caller"
	users, err := storage.GetMentionUsers(1, MentionEveryoneName)
	if err != nil || len(users) != 1 || users[0].Username != "alice" {
		t.Fatalf("stored users after input mutation = %v, %v", users, err)
	}
	users[0].Username = "changed snapshot"
	users[0] = nil
	again, err := storage.GetMentionUsers(1, MentionEveryoneName)
	if err != nil || len(again) != 1 || again[0].Username != "alice" {
		t.Fatalf("stored users after output mutation = %v, %v", again, err)
	}
}
