package model

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func newTestStorage(t *testing.T) *ChatStorage {
	t.Helper()
	return NewChatStorage(filepath.Join(t.TempDir(), "chats.json"))
}

func requireStorageError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil || err.Error() != want {
		t.Fatalf("got error %v, want %q", err, want)
	}
}

func TestChatStorageMentionsAndLimits(t *testing.T) {
	storage := newTestStorage(t)
	if got := storage.GetChatMentions(1); len(got) != 0 {
		t.Fatalf("unknown chat mentions: %v", got)
	}
	if err := storage.AddMention(1, "team"); err != nil {
		t.Fatal(err)
	}
	if got := storage.GetChatMentions(1); !reflect.DeepEqual(got, []string{"everyone", "team"}) {
		t.Fatalf("chat mentions = %v", got)
	}
	if !storage.IsMentionExists(1, "team") || storage.IsMentionExists(2, "team") {
		t.Fatal("mention index is not scoped to chat")
	}
	requireStorageError(t, storage.AddMention(1, "team"), ErrorDuplicateMention)
	for i := 2; i < MAX_MENTIONS_NUMBER; i++ {
		if err := storage.AddMention(1, string(rune('a'+i))); err != nil {
			t.Fatal(err)
		}
	}
	if got := len(storage.GetChatMentions(1)); got != MAX_MENTIONS_NUMBER {
		t.Fatalf("mention count = %d", got)
	}
	requireStorageError(t, storage.AddMention(1, "overflow"), ErrorExceededMaximumNumberOfMentions)
	if err := storage.AddMention(2, "team"); err != nil {
		t.Fatalf("limit should be per chat: %v", err)
	}
}

func TestChatStorageMembershipAndChatIsolation(t *testing.T) {
	storage := newTestStorage(t)
	alice := &User{ID: 1, Username: "alice"}
	bob := &User{ID: 2, Username: "bob"}
	requireStorageError(t, storage.AddUserToMention(1, "missing", alice), ErrorUnknownMention)
	if added := storage.AddUsersToMention(1, "missing", []*User{alice}); len(added) != 0 {
		t.Fatalf("added users to missing mention: %v", added)
	}
	_, err := storage.GetMentionUsers(1, "missing")
	requireStorageError(t, err, ErrorUnknownMention)
	if err := storage.AddMention(1, "team"); err != nil {
		t.Fatal(err)
	}
	added := storage.AddUsersToMention(1, "team", []*User{alice, alice, bob})
	if !reflect.DeepEqual(added, []*User{alice, bob}) {
		t.Fatalf("new members = %v", added)
	}
	if err := storage.AddUserToMention(1, "team", &User{ID: 1}); err != nil {
		t.Fatal(err)
	}
	users, err := storage.GetMentionUsers(1, "team")
	if err != nil || !reflect.DeepEqual(users, []*User{alice, bob}) {
		t.Fatalf("team members = %v, %v", users, err)
	}
	if err := storage.AddUserToMention(2, "everyone", alice); err != nil {
		t.Fatal(err)
	}
	storage.RemoveUser(1, alice)
	users, _ = storage.GetMentionUsers(1, "team")
	if !reflect.DeepEqual(users, []*User{bob}) {
		t.Fatalf("removal in chat 1 = %v", users)
	}
	users, _ = storage.GetMentionUsers(2, "everyone")
	if !reflect.DeepEqual(users, []*User{alice}) {
		t.Fatalf("removal affected another chat: %v", users)
	}
	storage.RemoveUser(999, bob)
}

func TestChatStorageMigrationPreservesMembership(t *testing.T) {
	storage := newTestStorage(t)
	if err := storage.AddMention(1, "team"); err != nil {
		t.Fatal(err)
	}
	user := &User{ID: 42}
	if err := storage.AddUserToMention(1, "team", user); err != nil {
		t.Fatal(err)
	}
	storage.ChangeChatId(1, 10)
	if len(storage.GetChatMentions(1)) != 0 || storage.IsMentionExists(1, "team") {
		t.Fatal("old chat still indexed after migration")
	}
	users, err := storage.GetMentionUsers(10, "team")
	if err != nil || !reflect.DeepEqual(users, []*User{user}) {
		t.Fatalf("migrated users = %v, %v", users, err)
	}
	if got := storage.GetChatMentions(10); !reflect.DeepEqual(got, []string{"everyone", "team"}) {
		t.Fatalf("migrated mentions = %v", got)
	}
}

func TestChatStorageLoadsPersistedChats(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chats.json")
	data := `{"Chats":[{"ID":7,"Mentions":[{"Name":"everyone","Users":[{"ID":12,"Username":"alice","FirstName":"Alice"}]},{"Name":"team","Users":[]}]}]}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	storage := NewChatStorage(path)
	if got := storage.GetChatMentions(7); !reflect.DeepEqual(got, []string{"everyone", "team"}) {
		t.Fatalf("loaded mentions = %v", got)
	}
	users, err := storage.GetMentionUsers(7, "everyone")
	if err != nil || len(users) != 1 || users[0].ID != 12 || users[0].Username != "alice" {
		t.Fatalf("loaded users = %v, %v", users, err)
	}
	if storage.IsMentionExists(8, "team") {
		t.Fatal("loaded mention leaked to another chat")
	}
}

func TestChatStoragePersistsMutationsAcrossReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chats.json")
	storage := NewChatStorage(path)
	if err := storage.AddMention(7, "team"); err != nil {
		t.Fatal(err)
	}
	if err := storage.AddUserToMention(7, "team", &User{ID: 12, Username: "alice"}); err != nil {
		t.Fatal(err)
	}
	reloaded := NewChatStorage(path)
	users, err := reloaded.GetMentionUsers(7, "team")
	if err != nil || len(users) != 1 || users[0].ID != 12 {
		t.Fatalf("members after reload = %v, %v", users, err)
	}
	removed, err := reloaded.RemoveUserFromMention(7, "team", &User{ID: 12})
	if err != nil || !removed {
		t.Fatalf("remove after reload = %t, %v", removed, err)
	}
	users, err = NewChatStorage(path).GetMentionUsers(7, "team")
	if err != nil || len(users) != 0 {
		t.Fatalf("removed members after reload = %v, %v", users, err)
	}
}
