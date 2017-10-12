package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// runAdminCommand description
func runAdminCommand(session *discordgo.Session, message *discordgo.MessageCreate, content string, args []string) bool {
	if strings.EqualFold("reload", args[0]) {
		adminReload(session, message, content, args)
		return true
	} else if strings.EqualFold("shutdown", args[0]) {
		adminShutdown(session, message, content, args)
		return true
	} else if strings.EqualFold("setuser", args[0]) {
		adminSetUser(session, message, content, args)
		return true
	} else if strings.EqualFold("setstatus", args[0]) {
		adminSetStatus(session, message, content, args)
		return true
	} else if strings.EqualFold("avatar", args[0]) {
		adminAvatar(session, message, content, args)
		return true
	}

	return false
}

// replyToAuthor description
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
		replyToAuthor(session, message, "you don't have permissions.", false)
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
		replyToAuthor(session, message, "you don't have permissions.", false)
		return
	}

	session.ChannelMessageSend(message.ChannelID, "Goodbye.")
	Duder.shutdown()
}

// adminSetUser description
func adminSetUser(session *discordgo.Session, message *discordgo.MessageCreate, content string, args []string) {
	if !Duder.permissions.isOwner(message.ChannelID, message.Author.ID) {
		replyToAuthor(session, message, "you don't have permissions.", false)
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

// adminSetStatus description
func adminSetStatus(session *discordgo.Session, message *discordgo.MessageCreate, content string, args []string) {
	if message.Author.ID != Duder.config.OwnerID {
		replyToAuthor(session, message, "you don't have permissions.", false)
		return
	}

	status := ""
	if len(args) > 1 {
		status = args[1]
	}

	if err := Duder.setStatus(status); err != nil {
		replyToAuthor(session, message, "unable to update status", false)
	}
}

var avatarPath = "avatars"

// adminAvatar description
func adminAvatar(session *discordgo.Session, message *discordgo.MessageCreate, content string, args []string) {
	if message.Author.ID != Duder.config.OwnerID {
		replyToAuthor(session, message, "you don't have permissions.", false)
		return
	}

	if _, err := os.Stat(avatarPath); os.IsNotExist(err) {
		os.Mkdir(avatarPath, 0777)
	}

	// args 1: action
	// args 2: url/file
	action := args[1]

	if len(args) < 2 {
		return
	}
	if strings.EqualFold("save", action) {
		if len(args) < 3 {
			replyToAuthor(session, message, "usage", false)
			return
		}
		file := args[2]
		session.ChannelTyping(message.ChannelID)

		baseURL := Duder.me.AvatarURL("256")
		//print("base " + baseURL + "\n")
		urlNoSize := baseURL[0 : len(baseURL)-9]
		//print("nosize " + urlNoSize + "\n")
		parts := strings.Split(urlNoSize, ".")
		ext := "." + parts[len(parts)-1]
		//print("ext " + ext + "\n")

		if !strings.HasSuffix(file, ext) {
			file = fmt.Sprintf("%s%s", file, ext)
		}
		//print("file " + file + "\n")

		h := createHTTPClient(5)
		resp, err := h.Get(Duder.me.AvatarURL("256"))
		if err != nil {
			replyToAuthor(session, message, "failed to download.", false)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			replyToAuthor(session, message, "failed to read.", false)
			return
		}
		if e := ioutil.WriteFile(fmt.Sprintf("%s/%s", avatarPath, file), body, 0644); e != nil {
			replyToAuthor(session, message, "unable to save.", false)
			return
		}
		replyToAuthor(session, message, "avatar saved.", false)
	} else if strings.EqualFold("download", action) {
		if len(args) < 3 {
			replyToAuthor(session, message, "usage", false)
			return
		}
		if avatarURL, err := url.Parse(args[2]); err != nil {
			replyToAuthor(session, message, "invalid URL.", false)
		} else {
			h := createHTTPClient(5)
			resp, err := h.Get(avatarURL.String())
			if err != nil {
				replyToAuthor(session, message, "failed to download.", false)
				return
			}

			img, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				replyToAuthor(session, message, "failed to read.", false)
				return
			}

			base64 := base64.StdEncoding.EncodeToString(img)
			avatar := fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(img), base64)

			_, err = Duder.session.UserUpdate("", "", Duder.me.Username, avatar, "")
			if err != nil {
				replyToAuthor(session, message, "failed to set.", false)
				return
			}

			replyToAuthor(session, message, ":ok_hand:", false)
		}
	} else if strings.EqualFold("list", action) {
		imgs := []string{}
		files, _ := ioutil.ReadDir(avatarPath)
		for _, f := range files {
			// ignore directories and non-image files
			if f.IsDir() || (!strings.HasSuffix(f.Name(), ".png") && !strings.HasSuffix(f.Name(), ".jpg") && !strings.HasSuffix(f.Name(), ".jpeg")) {
				continue
			}
			imgs = append(imgs, f.Name())
		}
		if len(imgs) == 0 {
			replyToAuthor(session, message, "no saved avatars.", false)
			return
		}
		var buffer bytes.Buffer
		buffer.WriteString("```")
		for i, img := range imgs {
			if i > 0 {
				buffer.WriteString("\n")
			}
			buffer.WriteString(img)
		}
		buffer.WriteString("```")
		replyToChannel(session, message, buffer.String())
	}
}
