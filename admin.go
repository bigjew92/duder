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
	"time"

	"github.com/bwmarrin/discordgo"
)

// test4

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

// adminAvatar description
func adminAvatar(session *discordgo.Session, message *discordgo.MessageCreate, content string, args []string) {
	var avatarPath = "avatars"

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

		h := http.Client{Timeout: time.Duration(5 * time.Second)}
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
			h := http.Client{Timeout: time.Duration(5 * time.Second)}
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
