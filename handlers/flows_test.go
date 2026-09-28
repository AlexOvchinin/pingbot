package handlers

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"fm/pingbot/model"
	tele "gopkg.in/telebot.v3"
)

// Embed Context so unanticipated Telegram calls fail the test instead of
// silently returning success. Only the methods used by these flows are faked.
type testContext struct {
	tele.Context
	message  *tele.Message
	callback *tele.Callback
	chat     *tele.Chat
	calls    []testCall
	result   error
}

type testCall struct {
	method string
	text   string
	opts   []interface{}
}

func (c *testContext) Message() *tele.Message   { return c.message }
func (c *testContext) Callback() *tele.Callback { return c.callback }
func (c *testContext) Chat() *tele.Chat         { return c.chat }
func (c *testContext) record(method string, what interface{}, opts ...interface{}) error {
	c.calls = append(c.calls, testCall{method, what.(string), opts})
	return c.result
}
func (c *testContext) Send(what interface{}, opts ...interface{}) error {
	return c.record("send", what, opts...)
}
func (c *testContext) Reply(what interface{}, opts ...interface{}) error {
	return c.record("reply", what, opts...)
}
func (c *testContext) Edit(what interface{}, opts ...interface{}) error {
	return c.record("edit", what, opts...)
}
func (c *testContext) EditOrReply(what interface{}, opts ...interface{}) error {
	return c.record("editOrReply", what, opts...)
}
func (c *testContext) RespondAlert(text string) error { return c.record("alert", text) }
func (c *testContext) Delete() error {
	c.calls = append(c.calls, testCall{method: "delete"})
	return c.result
}

func setupHandlers(t *testing.T) *testContext {
	t.Helper()
	previous := Storage
	Storage = model.NewChatStorage(filepath.Join(t.TempDir(), "chats.json"))
	t.Cleanup(func() { Storage = previous })
	sender := &tele.User{ID: 42, Username: "alice", FirstName: "Alice"}
	chat := &tele.Chat{ID: 7}
	return &testContext{chat: chat, message: &tele.Message{Sender: sender, Chat: chat}, callback: &tele.Callback{Sender: sender, Message: &tele.Message{Chat: chat}}}
}

func expectCall(t *testing.T, c *testContext, method, text string) {
	t.Helper()
	if len(c.calls) == 0 || c.calls[len(c.calls)-1].method != method || c.calls[len(c.calls)-1].text != text {
		t.Fatalf("calls = %+v; want last %s(%q)", c.calls, method, text)
	}
}

func testMessage(text string) *tele.Message {
	m := &tele.Message{Text: text, Sender: &tele.User{ID: 42, Username: "alice"}}
	m.Entities = append(m.Entities, tele.MessageEntity{Type: tele.EntityCommand, Offset: 0, Length: 4})
	return m
}

func TestCreateMentionCommand(t *testing.T) {
	for _, tc := range []struct {
		payload string
		want    string
	}{
		{"  team  ", "Mention team was successfully created"},
		{"team", "Mention team already exists"},
		{"", ErrorReplyCreateMention},
		{strings.Repeat("a", MaxMentionLength+1), ErrorReplyCreateMention},
		{"a&b", ErrorReplyCreateMentionForbiddenSymbols},
		{"a=b", ErrorReplyCreateMentionForbiddenSymbols},
	} {
		t.Run(tc.payload, func(t *testing.T) {
			c := setupHandlers(t)
			c.message.Payload = tc.payload
			if tc.want == "Mention team already exists" {
				_ = Storage.AddMention(7, "team")
			}
			if err := HandleCreateMention(c); err != nil {
				t.Fatal(err)
			}
			expectCall(t, c, "send", tc.want)
		})
	}
}

func TestJoinLeaveCommands(t *testing.T) {
	c := setupHandlers(t)
	_ = Storage.AddMention(7, "team")
	c.message.Payload = "team"
	if err := HandleJoinCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "Sucessfully added user @alice to mention team")
	if err := HandleLeaveCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "Successfully left mention team")
	if err := HandleLeaveCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "You are not in mention team")
	c.message.Payload = "missing"
	if err := HandleLeaveCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "Mention missing not found. Add it with /create_mention command")
	users, _ := Storage.GetMentionUsers(7, "team")
	if len(users) != 0 {
		t.Fatalf("team users after leave = %v", users)
	}
}

func TestMentionKeyboardAndCallback(t *testing.T) {
	c := setupHandlers(t)
	_ = Storage.AddMention(7, "zebra")
	_ = Storage.AddMention(7, "Alpha")
	c.message.Payload = ""
	if err := HandleJoinCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "Please choose which mention to join")
	markup := c.calls[0].opts[0].(*tele.ReplyMarkup)
	var labels []string
	for _, row := range markup.InlineKeyboard {
		labels = append(labels, row[0].Text)
	}
	if !reflect.DeepEqual(labels, []string{"everyone", "Alpha", "zebra", "cancel /join command"}) {
		t.Fatalf("keyboard labels = %v", labels)
	}
	for i, name := range labels[:3] {
		if !strings.Contains(markup.InlineKeyboard[i][0].Data, "mention="+name) {
			t.Fatalf("button data = %q", markup.InlineKeyboard[i][0].Data)
		}
	}
	c.callback.Data = "command=join&mention=Alpha"
	if err := OnCallback(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "Sucessfully added user @alice to mention Alpha")
	c.callback.Data = "command=leave&mention=Alpha"
	if err := OnCallback(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "Successfully left mention Alpha")
	c.callback.Data = "command=cancel"
	if err := OnCallback(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "edit", "Request cancelled")
	c.callback.Data = "command=invalid"
	if err := OnCallback(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "editOrReply", "Could not handle command, try again later")
}

func TestAddCommandAndSuggestion(t *testing.T) {
	c := setupHandlers(t)
	_ = Storage.AddMention(7, "team")
	c.message = testMessage("/add team @bob")
	c.message.Entities = append(c.message.Entities, tele.MessageEntity{Type: tele.EntityMention, Offset: 10, Length: 4})
	if err := HandleAddCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "send", "Added @bob to group team")
	if c.calls[0].opts[0] != tele.ModeHTML {
		t.Fatalf("parse mode = %v", c.calls[0].opts)
	}
	if err := HandleAddCommand(c); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "send", "All transferred users already belong to the group")
	c.message = testMessage("/add newteam")
	if err := HandleAddCommand(c); err != nil {
		t.Fatal(err)
	}
	if c.calls[len(c.calls)-1].method != "reply" || !strings.Contains(c.calls[len(c.calls)-1].text, "Create it and continue?") {
		t.Fatalf("suggestion = %+v", c.calls[len(c.calls)-1])
	}
}

func TestCreateAndContinueCallbackAuthorizationAndOriginalCommand(t *testing.T) {
	for _, tc := range []struct {
		name, command, original, mention string
		callbackID                       int64
		want                             string
		created                          bool
	}{
		{"add", "add", "/add team @bob", "team", 42, "Added @bob to group team", true},
		{"join", "join", "/join team", "team", 42, "Sucessfully added user @alice to mention team", true},
		{"wrong user", "join", "/join team", "team", 99, "Only the sender of the original command can create this mention", false},
		{"mismatched name", "join", "/join other", "team", 42, "Original command does not match this mention", false},
		{"bad command", "leave", "/leave team", "team", 42, "Unknown command", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := setupHandlers(t)
			original := testMessage(tc.original)
			if tc.command == "add" {
				original.Entities = append(original.Entities, tele.MessageEntity{Type: tele.EntityMention, Offset: 10, Length: 4})
			}
			c.callback.Message.ReplyTo = original
			c.callback.Sender.ID = tc.callbackID
			c.callback.Data = buildCommandString("create", map[string]string{"mention": tc.mention, "for": tc.command})
			if err := OnCallback(c); err != nil {
				t.Fatal(err)
			}
			method := "edit"
			if tc.name == "wrong user" {
				method = "alert"
			}
			if tc.name == "add" {
				method = "send"
			}
			if tc.name == "join" {
				method = "editOrReply"
			}
			expectCall(t, c, method, tc.want)
			if tc.created != Storage.IsMentionExists(7, tc.mention) {
				t.Fatalf("mention created = %t", Storage.IsMentionExists(7, tc.mention))
			}
		})
	}
}

func TestMentionFormattingAndExcludesCaller(t *testing.T) {
	c := setupHandlers(t)
	_ = Storage.AddMention(7, "team")
	_ = Storage.AddUserToMention(7, "team", &model.User{ID: 42, Username: "alice"})
	_ = Storage.AddUserToMention(7, "team", &model.User{ID: 99, FirstName: "A <B>"})
	if err := mention(c, getSenderUser(c), "team", "hi <all>"); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "send", "alice calling team: hi &lt;all&gt;\n<a href=\"tg://user?id=99\">A &lt;B&gt;</a>")
	if c.calls[0].opts[0] != tele.ModeHTML {
		t.Fatalf("parse mode = %v", c.calls[0].opts)
	}
	Storage.RemoveUser(7, &model.User{ID: 99})
	if err := mention(c, getSenderUser(c), "team", ""); err != nil {
		t.Fatal(err)
	}
	expectCall(t, c, "send", "Noone to mention in team. Please use /add to add users to mention manually or /join to join it yourself")
}

func TestHandlerPropagatesTelegramError(t *testing.T) {
	c := setupHandlers(t)
	want := errors.New("telegram unavailable")
	c.result = want
	c.message.Payload = "missing"
	if err := HandleLeaveCommand(c); !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
}

func TestUserJoinAndDepartureEvents(t *testing.T) {
	c := setupHandlers(t)
	_ = Storage.AddMention(7, "team")
	c.message.UserJoined = &tele.User{ID: 3, Username: "bob"}
	if err := HandleUserJoined(c); err != nil {
		t.Fatal(err)
	}
	users, _ := Storage.GetMentionUsers(7, "everyone")
	if len(users) != 1 || users[0].ID != 3 {
		t.Fatalf("joined users = %v", users)
	}
	_ = Storage.AddUserToMention(7, "team", &model.User{ID: 3})
	c.message.UserLeft = &tele.User{ID: 3, Username: "bob"}
	if err := HandleUserLeft(c); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"everyone", "team"} {
		users, _ := Storage.GetMentionUsers(7, name)
		if len(users) != 0 {
			t.Fatalf("%s still contains departing user: %v", name, users)
		}
	}
	c.message.UserJoined = &tele.User{ID: 4, IsBot: true}
	c.message.UserLeft = &tele.User{ID: 4, IsBot: true}
	if err := HandleUserJoined(c); err != nil {
		t.Fatal(err)
	}
	if err := HandleUserLeft(c); err != nil {
		t.Fatal(err)
	}
	users, _ = Storage.GetMentionUsers(7, "everyone")
	if len(users) != 0 {
		t.Fatalf("bot was added: %v", users)
	}
}

func TestParseAddMessageTextMentionAndRepeatedUsername(t *testing.T) {
	m := testMessage("/add @bob team @bob Alex")
	m.Entities = append(m.Entities,
		tele.MessageEntity{Type: tele.EntityMention, Offset: 5, Length: 4},
		tele.MessageEntity{Type: tele.EntityMention, Offset: 15, Length: 4},
		tele.MessageEntity{Type: tele.EntityTMention, Offset: 20, Length: 4, User: &tele.User{ID: 8, FirstName: "Alex"}},
	)
	name, users := parseAddMessage(m)
	if name != "team" || len(users) != 3 || users[0].Username != "bob" || users[1].Username != "bob" || users[2].ID != 8 {
		t.Fatalf("name = %q, users = %v", name, users)
	}
}

func TestMissingJoinSuggestsCreationWithoutAddingUser(t *testing.T) {
	c := setupHandlers(t)
	c.message.Payload = "team"
	if err := HandleJoinCommand(c); err != nil {
		t.Fatal(err)
	}
	if len(c.calls) != 1 || c.calls[0].method != "reply" || Storage.IsMentionExists(7, "team") {
		t.Fatalf("unexpected join suggestion: %+v", c.calls)
	}
	markup := c.calls[0].opts[0].(*tele.ReplyMarkup)
	if !strings.Contains(markup.InlineKeyboard[0][0].Data, "for=join") {
		t.Fatalf("creation callback = %s", markup.InlineKeyboard[0][0].Data)
	}
}
