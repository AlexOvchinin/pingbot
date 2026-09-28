package model

import "testing"

func TestIsSameUser(t *testing.T) {
	tests := []struct {
		name        string
		left, right *User
		want        bool
	}{
		{"both nil", nil, nil, true},
		{"one nil", nil, &User{ID: 1}, false},
		{"same ID", &User{ID: 1, Username: "old"}, &User{ID: 1, Username: "new"}, true},
		{"same username", &User{ID: 1, Username: "alice"}, &User{ID: 2, Username: "alice"}, true},
		{"different users", &User{ID: 1, Username: "alice"}, &User{ID: 2, Username: "bob"}, false},
		{"empty identity", &User{}, &User{}, false},
		{"zero ID", &User{ID: 0}, &User{ID: 0}, false},
		{"case sensitive username", &User{Username: "Alice"}, &User{Username: "alice"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSameUser(tt.left, tt.right); got != tt.want {
				t.Fatalf("IsSameUser(%v, %v) = %t, want %t", tt.left, tt.right, got, tt.want)
			}
			if got := IsSameUser(tt.right, tt.left); got != tt.want {
				t.Fatalf("reverse comparison = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestUserCollectionOperations(t *testing.T) {
	alice := &User{ID: 1, Username: "alice"}
	bob := &User{ID: 2, Username: "bob"}
	users := AddUser(alice, nil)
	users = AddUser(&User{ID: 1, Username: "renamed"}, users)
	users = AddUser(&User{Username: "bob"}, users)
	users = AddUser(bob, users)
	if len(users) != 2 || users[0] != alice || users[1].Username != "bob" {
		t.Fatalf("unexpected deduplicated users: %v", users)
	}
	if !ContainsUser(&User{Username: "alice"}, users) || ContainsUser(&User{ID: 3}, users) {
		t.Fatalf("ContainsUser did not match user identity: %v", users)
	}
	remaining := RemoveUser(&User{ID: 1}, users)
	if len(remaining) != 1 || remaining[0].Username != "bob" || len(users) != 2 {
		t.Fatalf("removal should preserve other users and original slice: remaining=%v original=%v", remaining, users)
	}
	if got := RemoveUser(&User{ID: 100}, remaining); len(got) != 1 {
		t.Fatalf("removing absent user changed collection: %v", got)
	}
}
