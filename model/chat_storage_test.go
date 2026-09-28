package model

import (
	"path/filepath"
	"testing"
)

func TestRemoveUserFromMention(t *testing.T) {
	storage := NewChatStorage(filepath.Join(t.TempDir(), "chats.json"))
	user := &User{ID: 42, Username: "alice"}
	if err := storage.AddMention(1, "team"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{MentionEveryoneName, "team"} {
		if err := storage.AddUserToMention(1, name, user); err != nil {
			t.Fatal(err)
		}
	}
	removed, err := storage.RemoveUserFromMention(1, "team", user)
	if err != nil || !removed {
		t.Fatalf("remove: removed=%v err=%v", removed, err)
	}
	team, _ := storage.GetMentionUsers(1, "team")
	everyone, _ := storage.GetMentionUsers(1, MentionEveryoneName)
	if len(team) != 0 || len(everyone) != 1 {
		t.Fatalf("team=%v everyone=%v", team, everyone)
	}
	removed, err = storage.RemoveUserFromMention(1, "team", user)
	if err != nil || removed {
		t.Fatalf("second remove: removed=%v err=%v", removed, err)
	}
	_, err = storage.RemoveUserFromMention(1, "missing", user)
	if err == nil || err.Error() != ErrorUnknownMention {
		t.Fatalf("expected unknown mention, got %v", err)
	}
}

func TestRemoveUsersFromMention(t *testing.T) {
	storage := NewChatStorage(filepath.Join(t.TempDir(), "chats.json"))
	first := &User{ID: 1, Username: "alice"}
	second := &User{ID: 2, Username: "bob"}
	storage.AddMention(1, "team")
	storage.AddUsersToMention(1, "team", []*User{first, second})
	storage.AddUsersToMention(1, MentionEveryoneName, []*User{first, second})
	removed, err := storage.RemoveUsersFromMention(1, "team", []*User{first, &User{ID: 3}, first})
	if err != nil || len(removed) != 1 || removed[0] != first {
		t.Fatalf("removed=%v err=%v", removed, err)
	}
	team, _ := storage.GetMentionUsers(1, "team")
	everyone, _ := storage.GetMentionUsers(1, MentionEveryoneName)
	if len(team) != 1 || !IsSameUser(team[0], second) || len(everyone) != 2 {
		t.Fatalf("team=%v everyone=%v", team, everyone)
	}
	_, err = storage.RemoveUsersFromMention(1, "unknown", []*User{first})
	if err == nil || err.Error() != ErrorUnknownMention {
		t.Fatalf("expected unknown mention, got %v", err)
	}
}
