package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

/// replyToAuthor description
func replyToAuthor(session *discordgo.Session, message *discordgo.MessageCreate, content string, mention bool) {
	if mention {
		session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("<@%s> %s", message.Author.ID, content))
	} else {
		session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%s, %s", message.Author.Username, content))
	}
}

// replyToChannel description
func replyToChannel(session *discordgo.Session, message *discordgo.MessageCreate, content string) {
	session.ChannelMessageSend(message.ChannelID, content)
}

// adminReload description
func adminReload(session *discordgo.Session, message *discordgo.MessageCreate, content string, args []string) {
	if message.Author.ID != Duder.config.OwnerID {
		replyToAuthor(session, message, "you don't have permissions", false)
		return
	}

	loadRugs(Duder.config.RugPath)
	if len(rugLoadErrors) > 0 {
		session.ChannelMessageSend(message.ChannelID, ":octagonal_sign: Rugs reloaded with errors.")
	} else {
		session.ChannelMessageSend(message.ChannelID, ":ok_hand: Rugs successfully reloaded.")
	}
}

// adminShutdown description
func adminShutdown(session *discordgo.Session, message *discordgo.MessageCreate, content string, args []string) {
	if message.Author.ID != Duder.config.OwnerID {
		replyToAuthor(session, message, "you don't have permissions", false)
		return
	}

	session.ChannelMessageSend(message.ChannelID, "Goodbye.")
	Duder.shutdown()
}

// adminSetUser description
func adminSetUser(session *discordgo.Session, message *discordgo.MessageCreate, content string, args []string) {
	if !Duder.permissions.isOwner(message.ChannelID, message.Author.ID) {
		replyToAuthor(session, message, "you don't have permissions", false)
		return
	}

	if (len(message.Mentions) != 1) || (len(args) < 3) {
		replyToAuthor(session, message, "usage: `setuser @mention (+/-)permission`", false)
		return
	} else if message.Author.ID == message.Mentions[0].ID {
		replyToAuthor(session, message, "you cannot change your own permissions", false)
		return
	}

	modifier := args[2][:1]

	if modifier != "+" && modifier != "-" {
		replyToAuthor(session, message, "the permission must start with `+` or `-` to add or remove.", false)
		return
	}

	permName := args[2][1:]
	perm := Duder.permissions.getByName(permName)
	if perm.Value == -1 {
		replyToAuthor(session, message, fmt.Sprintf("invalid permission '%s'.", permName), false)
		return
	}

	user := message.Mentions[0]

	if modifier == "+" {
		if err := Duder.permissions.addToUser(message.ChannelID, user.ID, perm.Value); err != nil {
			replyToAuthor(session, message, fmt.Sprintf("unable to add permission '%s'.", err), false)
			log.Print("adding ", perm.Value)
		}
	} else {
		if err := Duder.permissions.removeFromUser(message.ChannelID, user.ID, perm.Value); err != nil {
			replyToAuthor(session, message, fmt.Sprintf("unable to remove permission '%s'.", err), false)
		}
	}

	perms := Duder.permissions.getAll(message.ChannelID, user.ID)
	if len(perms) == 0 {
		replyToAuthor(session, message, fmt.Sprintf("%s doesn't have any permissions left.", user.Username), false)
	} else {
		var buffer bytes.Buffer
		for _, p := range perms {
			if buffer.Len() > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteString(Duder.permissions.getByValue(p).Names[0])
		}
		replyToAuthor(session, message, fmt.Sprintf("%s now has permission(s) %s.", user.Username, buffer.String()), false)
	}
}
