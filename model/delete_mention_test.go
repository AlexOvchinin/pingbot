package model

import (
	"path/filepath"
	"testing"
)

func TestDeleteMention(t *testing.T) {
	storage := NewChatStorage(filepath.Join(t.TempDir(), "chats.json"))
	if err := storage.AddMention(1, "team"); err != nil {
		t.Fatal(err)
	}
	storage.AddUserToMention(1, "team", &User{ID: 1})
	if err := storage.DeleteMention(1, "team"); err != nil {
		t.Fatal(err)
	}
	if storage.IsMentionExists(1, "team") || len(storage.GetChatMentions(1)) != 1 {
		t.Fatalf("mentions after delete: %v", storage.GetChatMentions(1))
	}
	if err := storage.AddMention(1, "team"); err != nil {
		t.Fatal(err)
	}
	users, _ := storage.GetMentionUsers(1, "team")
	if len(users) != 0 {
		t.Fatalf("deleted membership restored: %v", users)
	}
	if err := storage.DeleteMention(1, MentionEveryoneName); err == nil || err.Error() != ErrorProtectedMention {
		t.Fatalf("expected protected mention error, got %v", err)
	}
	if err := storage.DeleteMention(1, "absent"); err == nil || err.Error() != ErrorUnknownMention {
		t.Fatalf("expected unknown mention error, got %v", err)
	}
}
