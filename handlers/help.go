package handlers

import (
	tele "gopkg.in/telebot.v3"
)

func HandleHelpCommand(ctx tele.Context) error {
	return ctx.Send(`You can use following commands:
/add @username1 @username2 - Adds listed users to default mention "everyone"
/everyone - Mentions all users that were added to the "everyone" mention
/create_mention mention-name - Creates new mention named "mention-name". Maximum mention length is 20 latin and digit symbols. You can add up to 10 mentions for a chat.
/join - allows user to join some mention by presenting options. Could also be used with mention name (/join everyone) to skip selecting options step.
/mention - calls mention by presenting by presenting options. Could also be used with mention name (e.g. /mention everyone) to skip selecting options step.
/help - Presents help text

Additionally, you can tag your groups by simply typing their name as a command in message (/everyone). This should work by default in most cases.
`)
}
