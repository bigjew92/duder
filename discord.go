package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/foszor/duder/helpers/rugutils"
)

func init() {
	Duder.Discord = DiscordManager{}
}

// DiscordManager description
type DiscordManager struct {
	session *discordgo.Session
	me      *discordgo.User
	owner   *discordgo.User
}

// Connect description
func (manager *DiscordManager) Connect() error {
	// create the Discord session
	Duder.Logf(LogChannel.General, "Creating Discord session with token '%v'", Duder.Config.BotToken())
	session, err := discordgo.New(Duder.Config.BotToken())
	if err != nil {
		return fmt.Errorf("Error creating discord session; %s", err)
	}
	manager.session = session

	// obtain bot account details
	Duder.Log(LogChannel.General, "Obtaining bot account details")
	me, err := manager.session.User("@me")
	if err != nil {
		return fmt.Errorf("Error obtaining bot account details; %s", err)
	}
	manager.me = me
	Duder.Log(LogChannel.General, "> Bot Client ID:", manager.me.ID)

	// obtain owner account details
	Duder.Log(LogChannel.General, "Obtaining owner account details")
	owner, err := manager.session.User(Duder.Config.OwnerID())
	if err != nil {
		return fmt.Errorf("Error obtaining owner account details; %s", err)
	}
	manager.owner = owner
	Duder.Log(LogChannel.General, "> Owner Client ID:", manager.owner.ID)

	// register callback for messageCreate
	//manager.session.AddHandler(onMessageCreate)
	manager.session.AddHandler(func(session *discordgo.Session, message *discordgo.MessageCreate) {
		if strings.HasPrefix(message.Content, fmt.Sprintf("%s ", Duder.Config.CommandPrefix())) {
			Duder.Log(LogChannel.Verbose, "Proccessing command", message.Content)
			manager.runCommand(message)
		}
	})

	// open the Discord connection
	Duder.Log(LogChannel.General, "Opening Discord connection")
	err = manager.session.Open()
	if err != nil {
		return fmt.Errorf("Error opening discord connection; %s", err)
	}

	manager.SetStatus(Duder.Config.Status())

	return nil
}

// MemberUsername description
func (manager *DiscordManager) MemberUsername(guildID string, memberID string) string {
	if guild, err := manager.session.Guild(guildID); err == nil {
		for _, member := range guild.Members {
			if member.User.ID == memberID {
				return member.User.Username
			}
		}
	}

	return "Unknown"
}

// SetMemberNickname description
func (manager *DiscordManager) SetMemberNickname(guildID string, memberID string, nickname string) {
	manager.session.GuildMemberNickname(guildID, memberID, nickname)
}

// Guild description
func (manager *DiscordManager) Guild(guildID string) (*discordgo.Guild, bool) {
	guild, err := manager.session.Guild(guildID)
	if err != nil {
		return nil, false
	}

	return guild, true
}

// MessageGuild description
func (manager *DiscordManager) MessageGuild(message *discordgo.MessageCreate) (*discordgo.Guild, bool) {
	channel, ok := manager.MessageChannel(message)
	if !ok {
		return nil, false
	}

	guild, err := manager.session.Guild(channel.GuildID)
	if err != nil {
		return nil, false
	}

	return guild, true
}

// MessageChannel description
func (manager *DiscordManager) MessageChannel(message *discordgo.MessageCreate) (*discordgo.Channel, bool) {
	channel, err := manager.session.Channel(message.ChannelID)
	if err != nil {
		return nil, false
	}

	return channel, true
}

// DeleteChannelMessage description
func (manager *DiscordManager) DeleteChannelMessage(channelID string, messageID string) {
	manager.session.ChannelMessageDelete(channelID, messageID)
}

// SendMessageToChannel description
func (manager *DiscordManager) SendMessageToChannel(channelID string, content string) {
	manager.session.ChannelMessageSend(channelID, content)
}

// SendEmbedToChannel description
func (manager *DiscordManager) SendEmbedToChannel(channelID string, content *discordgo.MessageEmbed) {
	manager.session.ChannelMessageSendEmbed(channelID, content)
}

// SetStatus description
func (manager *DiscordManager) SetStatus(status string) bool {
	Duder.Logf(LogChannel.Verbose, "[Duder.SetStatus] Setting status to '%s'", status)
	if err := manager.session.UpdateStatus(0, status); err != nil {
		return false
	}
	return true
}

// StartTyping description
func (manager *DiscordManager) StartTyping(channelID string) bool {
	if err := manager.session.ChannelTyping(channelID); err != nil {
		Duder.Logf(LogChannel.Verbose, "[Duder.StartTyping] Unable to start typing in channel '%s'; %s", channelID, err.Error())
		return false
	}
	return true
}

// SetAvatarByImage description
func (manager *DiscordManager) SetAvatarByImage(base64 string) bool {
	if _, err := manager.session.UserUpdate("", "", "", base64, ""); err != nil {
		Duder.Log(LogChannel.Verbose, "[Duder.SetAvatarByImage] Unable to set avatar;", err.Error())
		return false
	}

	return true
}

// SaveAvatar description
func (manager *DiscordManager) SaveAvatar(filename string) error {
	if len(filename) == 0 {
		Duder.Log(LogChannel.Verbose, "[Duder.SaveAvatar] Unable to save avatar; no filename provided")
		return errors.New("no filename provided")
	}

	// get the avatar URL
	baseURL := manager.me.AvatarURL("256")

	// strip the size portion of the URL
	urlNoSize := baseURL[0 : len(baseURL)-9]

	// extract the file extension
	parts := strings.Split(urlNoSize, ".")
	ext := "." + parts[len(parts)-1]

	// make sure the new filename has the same extension
	if !strings.HasSuffix(filename, ext) {
		filename = fmt.Sprintf("%s%s", filename, ext)
	}

	req := http.Client{Timeout: time.Duration(5 * time.Second)}
	resp, err := req.Get(baseURL)
	if err != nil {
		Duder.Log(LogChannel.Verbose, "[Duder.SaveAvatar] Unable to save avatar;", err.Error())
		return err
	}
	defer resp.Body.Close()

	if _, err := os.Stat("avatars"); os.IsNotExist(err) {
		os.Mkdir("avatars", 0777)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Duder.Log(LogChannel.Verbose, "[Duder.SaveAvatar] Unable to read response body;", err.Error())
		return err
	}
	if err = ioutil.WriteFile(fmt.Sprintf("%s/%s", "avatars", filename), data, 0777); err != nil {
		Duder.Log(LogChannel.Verbose, "[Duder.SaveAvatar] Unable to write file;", err.Error())
		return err
	}

	return nil
}

// Avatars description
func (manager *DiscordManager) Avatars() []string {
	if _, err := os.Stat("avatars"); os.IsNotExist(err) {
		os.Mkdir("avatars", 0777)
	}

	avatars := []string{}
	files, _ := ioutil.ReadDir("avatars")
	for _, f := range files {
		// ignore directories and non-image files
		if f.IsDir() || (!strings.HasSuffix(f.Name(), ".png") && !strings.HasSuffix(f.Name(), ".jpg") && !strings.HasSuffix(f.Name(), ".jpeg")) {
			continue
		}
		avatars = append(avatars, f.Name())
	}

	return avatars
}

// SetAvatarByFile description
func (manager *DiscordManager) SetAvatarByFile(filename string) error {
	if _, err := os.Stat("avatars"); os.IsNotExist(err) {
		Duder.Log(LogChannel.Verbose, "[Duder.SetAvatarByFile] Unable to set avatar; avatar path missing")
		return errors.New("avatars path missing")
	}

	filePath := fmt.Sprintf("%s/%s", "avatars", filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		Duder.Logf(LogChannel.Verbose, "[Duder.SetAvatarByFile] Unable to set avatar; file '%s' doesn't exist", filePath)
		return errors.New("file doesn't exist")
	}

	var bytes []byte
	var err error
	if bytes, err = ioutil.ReadFile(filePath); err != nil {
		Duder.Logf(LogChannel.Verbose, "[Duder.SetAvatarByFile] Couldn't read avatar file '%s'; %s", filePath, err.Error())
		return errors.New("couldn't read avatar file")
	}
	base64 := base64.StdEncoding.EncodeToString(bytes)
	avatar := fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(bytes), base64)
	if ok := manager.SetAvatarByImage(avatar); !ok {
		Duder.Log(LogChannel.Verbose, "[Duder.SetAvatarByFile] Couldn't set avatar")
		return errors.New("couldn't set avatar")
	}

	return nil
}

// teardown description
func (manager *DiscordManager) teardown() {
	manager.session.Close()
}

// onMessageCreate description
/*
func onMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if strings.HasPrefix(message.Content, fmt.Sprintf("%s ", Duder.Config.CommandPrefix())) {
		Duder.Log(LogChannel.Verbose, "Proccessing command", message.Content)
		runCommand(message)
	}
}
*/

// runCommand description
func (manager *DiscordManager) runCommand(message *discordgo.MessageCreate) {
	// strip the command prefix from the message content
	content := message.Content[len(Duder.Config.CommandPrefix())+1 : len(message.Content)]
	content = strings.TrimSpace(content)

	// get the root command
	args := rugutils.ParseArguments(content)
	if len(args) == 0 {
		return
	}
	cmd := strings.ToLower(args[0])
	Duder.Logf(LogChannel.Verbose, "Root command '%s'", cmd)

	// hardcoded commands
	switch cmd {
	case "update":
		Duder.Update(message)
		return
	case "shutdown":
		Duder.Shutdown(message)
		return
	}

	// check each rug to find the matching command
	for _, rug := range Duder.Rugs.Rugs {
		for _, rugCmd := range rug.Commands {
			if rugCmd.Trigger == cmd {
				rugCmd.Run(rug, message, args)
				return
			}
		}
	}
}
