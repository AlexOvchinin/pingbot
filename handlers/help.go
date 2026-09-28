package handlers

import (
	tele "gopkg.in/telebot.v3"
)

func HandleHelpCommand(ctx tele.Context) error {
	return ctx.Send(`You can use following commands:
/add @username1 @username2 - Adds listed users to default mention "everyone". Include a mention name anywhere to add to that mention (e.g. /add team @username1 @username2).
/remove team @username1 @username2 - Removes listed users from the named mention (name can appear anywhere among users).
/delete team - Asks for confirmation before deleting a mention. Use /delete to choose from options. The default everyone mention cannot be deleted.
/everyone - Mentions all users that were added to the "everyone" mention
/create_mention mention-name - Creates new mention named "mention-name". Maximum mention length is 20 latin and digit symbols. You can add up to 10 mentions for a chat.
/join - choose a mention, or use /join team ops to join several. Separate multi-word names with commas (e.g. /join new team, ops).
/leave - choose a mention, or use /leave team ops to leave several. Separate multi-word names with commas.
/mention - calls mention by presenting by presenting options. Could also be used with mention name (e.g. /mention everyone) to skip selecting options step.
/help - Presents help text

Additionally, you can tag your groups by simply typing their name as a command in message (/everyone). This should work by default in most cases.
`)
}
