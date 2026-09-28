package handlers

import (
	"fmt"
	"strings"

	"fm/pingbot/model"
	tele "gopkg.in/telebot.v3"
)

const CREATE_AND_CONTINUE_COMMAND_NAME = "create"
const CREATE_MENTIONS_COMMAND_NAME = "create_many"
const ORIGINAL_COMMAND_ARGUMENT_NAME = "for"

func suggestCreatingMention(ctx tele.Context, mentionName, command string) error {
	if len(mentionName) > MaxMentionLength || strings.TrimSpace(mentionName) == "" {
		return ctx.Send(ErrorReplyCreateMention)
	}
	if re.MatchString(mentionName) {
		return ctx.Send(ErrorReplyCreateMentionForbiddenSymbols)
	}
	markup := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{{
		Text: fmt.Sprintf("Create %s and continue /%s", mentionName, command),
		Data: buildCommandString(CREATE_AND_CONTINUE_COMMAND_NAME, map[string]string{
			MENTION_ARGUMENT_NAME: mentionName, ORIGINAL_COMMAND_ARGUMENT_NAME: command,
		}),
	}}}}
	markup.InlineKeyboard = addCancelButton(markup.InlineKeyboard, command)
	// The reply preserves the original command and its user entities for /add.
	return ctx.Reply(fmt.Sprintf("Mention %s does not exist. Create it and continue?", mentionName), markup)
}

func suggestCreatingMentions(ctx tele.Context, missing []string) error {
	for _, name := range missing {
		if name == "" || len(name) > MaxMentionLength || re.MatchString(name) {
			return ctx.Send(ErrorReplyCreateMention)
		}
	}
	markup := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{{
		Text: fmt.Sprintf("Create %s and join", strings.Join(missing, ", ")),
		Data: buildCommandString(CREATE_MENTIONS_COMMAND_NAME, nil),
	}}}}
	markup.InlineKeyboard = addCancelButton(markup.InlineKeyboard, JOIN_COMMAND_NAME)
	return ctx.Reply(fmt.Sprintf("Mentions %s do not exist. Create them and join?", strings.Join(missing, ", ")), markup)
}

func handleCreateMentionsCallback(ctx tele.Context, _ map[string]string) error {
	original := ctx.Callback().Message.ReplyTo
	fields := strings.Fields(original.Text)
	if len(fields) == 0 || (fields[0] != "/join" && !strings.HasPrefix(fields[0], "/join@")) {
		return ctx.Edit("Original command does not match /join")
	}
	payload := strings.TrimSpace(strings.TrimPrefix(original.Text, fields[0]))
	existing := Storage.GetChatMentions(ctx.Chat().ID)
	names := parseMentionNames(payload, existing)
	missing := []string{}
	for _, name := range names {
		if name == "" || len(name) > MaxMentionLength || re.MatchString(name) {
			return ctx.Edit(ErrorReplyCreateMention)
		}
		if name != model.MentionEveryoneName && !Storage.IsMentionExists(ctx.Chat().ID, name) {
			missing = append(missing, name)
		}
	}
	count := len(existing)
	if count == 0 {
		count = 1 // creating a chat also creates everyone
	}
	if count+len(missing) > model.MAX_MENTIONS_NUMBER {
		return ctx.Edit("Exceed maximum number of mentions for current chat")
	}
	for _, name := range missing {
		if err := Storage.AddMention(ctx.Chat().ID, name); err != nil && err.Error() != model.ErrorDuplicateMention {
			return ctx.Edit(mapStorageErrorToBotError(err, name))
		}
	}
	return joinMentions(ctx, getCallbackUser(ctx), names)
}

func handleCreateAndContinueCallback(ctx tele.Context, arguments map[string]string) error {
	mentionName := arguments[MENTION_ARGUMENT_NAME]
	command := arguments[ORIGINAL_COMMAND_ARGUMENT_NAME]
	if command != ADD_COMMAND_NAME && command != JOIN_COMMAND_NAME {
		return ctx.Edit("Unknown command")
	}
	if len(mentionName) == 0 || len(mentionName) > MaxMentionLength || re.MatchString(mentionName) {
		return ctx.Edit(ErrorReplyCreateMention)
	}
	original := ctx.Callback().Message.ReplyTo
	fields := strings.Fields(original.Text)
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "/"+command) {
		return ctx.Edit("Original command does not match this mention")
	}
	if command == ADD_COMMAND_NAME {
		name, _ := parseAddMessage(original)
		if name != mentionName {
			return ctx.Edit("Original command does not match this mention")
		}
	} else if strings.TrimSpace(strings.TrimPrefix(original.Text, fields[0])) != mentionName {
		return ctx.Edit("Original command does not match this mention")
	}
	if err := Storage.AddMention(ctx.Chat().ID, mentionName); err != nil {
		if err.Error() != model.ErrorDuplicateMention {
			return ctx.Edit(mapStorageErrorToBotError(err, mentionName))
		}
	}
	if err := ctx.Edit(fmt.Sprintf("Mention %s created", mentionName)); err != nil {
		return err
	}
	if command == JOIN_COMMAND_NAME {
		return addSenderToMention(ctx, getCallbackUser(ctx), mentionName)
	}
	_, users := parseAddMessage(original)
	return addUsersToMention(ctx, mentionName, users)
}
