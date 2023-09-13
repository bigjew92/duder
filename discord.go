package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Duder.Discord = new(DiscordManager)
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
	Duder.Logf(LogGeneral, "Creating Discord session with token '%v'", Duder.Config.BotToken())
	session, err := discordgo.New(Duder.Config.BotToken())
	if err != nil {
		return fmt.Errorf("Error creating discord session; %s", err)
	}
	manager.session = session

	// obtain bot account details
	Duder.Log(LogGeneral, "Obtaining bot account details")
	me, err := manager.session.User("@me")
	if err != nil {
		return fmt.Errorf("Error obtaining bot account details; %s", err)
	}
	manager.me = me
	Duder.Log(LogGeneral, "> Bot Client ID: ", manager.me.ID)

	// obtain owner account details
	Duder.Log(LogGeneral, "Obtaining owner account details")
	owner, err := manager.session.User(Duder.Config.OwnerID())
	if err != nil {
		return fmt.Errorf("Error obtaining owner account details; %s", err)
	}
	manager.owner = owner
	Duder.Log(LogGeneral, "> Owner Client ID: ", manager.owner.ID)

	// register callbacks
	manager.session.AddHandler(manager.onMessageCreate)
	manager.session.AddHandler(manager.onMessageReactionAdd)
	manager.session.AddHandler(manager.onMessageReactionRemove)
	manager.session.AddHandler(manager.onPresenceUpdate)

	// open the Discord connection
	Duder.Log(LogGeneral, "Opening Discord connection")
	err = manager.session.Open()
	if err != nil {
		return fmt.Errorf("Error opening discord connection; %s", err)
	}

	manager.SetStatus(Duder.Config.Status())

	return nil
}

// GetGuildMember description
func (manager *DiscordManager) GetGuildMember(guildID string, userID string) (*discordgo.Member, bool) {
	guild, ok := manager.GetGuildByID(guildID)
	if !ok {
		return nil, false
	}

	for _, member := range guild.Members {
		if member.User.ID == userID {
			return member, true
		}
	}

	return nil, false
}

// SetMemberNickname description
func (manager *DiscordManager) SetMemberNickname(guildID string, memberID string, nickname string) {
	manager.session.GuildMemberNickname(guildID, memberID, nickname)
}

// GetMemberNickname description
func (manager *DiscordManager) GetMemberNickname(guildID string, userID string) (string, bool) {
	//manager.session.GuildMemberNickname(guildID, memberID, nickname)
	member, ok := manager.GetGuildMember(guildID, userID)
	if !ok {
		return "Unknown", false
	}

	if len(member.Nick) == 0 {
		return member.User.Username, true
	}

	return member.Nick, true
}

// GetMemberByNickname description
func (manager *DiscordManager) GetMemberByNickname(guildID string, nickname string) (*discordgo.Member, bool) {
	guild, ok := manager.GetGuildByID(guildID)
	if !ok {
		return nil, false
	}

	nickname = strings.ToLower(nickname)

	for _, member := range guild.Members {
		if len(member.Nick) > 0 {
			if strings.ToLower(member.Nick) == nickname {
				return member, true
			}
		} else if strings.ToLower(member.User.Username) == nickname {
			return member, true
		}
	}

	return nil, false
}

// GetGuildByID description
func (manager *DiscordManager) GetGuildByID(guildID string) (*discordgo.Guild, bool) {
	guild, err := manager.session.Guild(guildID)
	if err != nil {
		return nil, false
	}

	return guild, true
}

// GetMessageGuild description
func (manager *DiscordManager) GetMessageGuild(message *discordgo.MessageCreate) (*discordgo.Guild, bool) {
	channel, ok := manager.GetMessageChannel(message)
	if !ok {
		return nil, false
	}

	guild, err := manager.session.Guild(channel.GuildID)
	if err != nil {
		return nil, false
	}

	return guild, true
}

// GetMessageChannel description
func (manager *DiscordManager) GetMessageChannel(message *discordgo.MessageCreate) (*discordgo.Channel, bool) {
	channel, ok := manager.GetChannelByID(message.ChannelID)
	if !ok {
		return nil, false
	}

	return channel, true
}

// GetGuildByChannelID description
func (manager *DiscordManager) GetGuildByChannelID(channelID string) (*discordgo.Guild, bool) {
	channel, ok := manager.GetChannelByID(channelID)
	if !ok {
		return nil, false
	}
	guild, ok := manager.GetGuildByID(channel.GuildID)
	if !ok {
		return nil, false
	}
	return guild, true
}

// GetChannelByID description
func (manager *DiscordManager) GetChannelByID(channelID string) (*discordgo.Channel, bool) {
	channel, err := manager.session.Channel(channelID)
	if err != nil {
		return nil, false
	}

	return channel, true
}

// ChannelTypeName description
func (manager *DiscordManager) ChannelTypeName(channel *discordgo.Channel) string {
	switch channel.Type {
	case discordgo.ChannelTypeDM:
		return "Direct Message"
	case discordgo.ChannelTypeGroupDM:
		return "Group Message"
	case discordgo.ChannelTypeGuildCategory:
		return "Guild Category"
	case discordgo.ChannelTypeGuildText:
		return "Guild Text"
	case discordgo.ChannelTypeGuildVoice:
		return "Guild Voice"
	}

	return "Unknown"
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
func (manager *DiscordManager) SendEmbedToChannel(channelID string, jsonData string) error {
	// https://godoc.org/github.com/bwmarrin/discordgo#MessageEmbed

	embed := new(discordgo.MessageEmbed)
	if err := json.Unmarshal([]byte(jsonData), &embed); err != nil {
		return err
	}

	manager.session.ChannelMessageSendEmbed(channelID, embed)

	return nil
}

// SendFileToChannel description
func (manager *DiscordManager) SendFileToChannel(channelID string, name string, reader io.Reader) {
	//data := discordgo.MessageSend{}
	//data.File = discordgo.File{}
	//manager.session.ChannelFileSend(channelID)
}

// SetStatus description
func (manager *DiscordManager) SetStatus(status string) bool {
	Duder.Logf(LogVerbose, "[Duder.SetStatus] Setting status to '%s'", status)
	if err := manager.session.UpdateWatchStatus(0, status); err != nil {
		return false
	}
	return true
}

// StartTyping description
func (manager *DiscordManager) StartTyping(channelID string) bool {
	if err := manager.session.ChannelTyping(channelID); err != nil {
		Duder.Logf(LogVerbose, "[Duder.StartTyping] Unable to start typing in channel '%s'; %s", channelID, err.Error())
		return false
	}
	return true
}

// SetAvatarByImage description
func (manager *DiscordManager) SetAvatarByImage(base64 string) bool {
	if _, err := manager.session.UserUpdate("", base64); err != nil {
		Duder.Log(LogVerbose, "[Duder.SetAvatarByImage] Unable to set avatar;", err.Error())
		return false
	}

	return true
}

// SaveAvatar description
func (manager *DiscordManager) SaveAvatar(filename string) error {
	if len(filename) == 0 {
		Duder.Log(LogVerbose, "[Duder.SaveAvatar] Unable to save avatar; no filename provided")
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
		Duder.Log(LogVerbose, "[Duder.SaveAvatar] Unable to save avatar;", err.Error())
		return err
	}
	defer resp.Body.Close()

	if _, err := os.Stat("avatars"); os.IsNotExist(err) {
		os.Mkdir("avatars", 0777)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		Duder.Log(LogVerbose, "[Duder.SaveAvatar] Unable to read response body;", err.Error())
		return err
	}
	if err = os.WriteFile(fmt.Sprintf("%s/%s", "avatars", filename), data, 0777); err != nil {
		Duder.Log(LogVerbose, "[Duder.SaveAvatar] Unable to write file;", err.Error())
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
	files, _ := os.ReadDir("avatars")
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
		Duder.Log(LogVerbose, "[Duder.SetAvatarByFile] Unable to set avatar; avatar path missing")
		return errors.New("avatars path missing")
	}

	filePath := fmt.Sprintf("%s/%s", "avatars", filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		Duder.Logf(LogVerbose, "[Duder.SetAvatarByFile] Unable to set avatar; file '%s' doesn't exist", filePath)
		return errors.New("file doesn't exist")
	}

	var bytes []byte
	var err error
	if bytes, err = os.ReadFile(filePath); err != nil {
		Duder.Logf(LogVerbose, "[Duder.SetAvatarByFile] Couldn't read avatar file '%s'; %s", filePath, err.Error())
		return errors.New("couldn't read avatar file")
	}
	base64 := base64.StdEncoding.EncodeToString(bytes)
	avatar := fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(bytes), base64)
	if ok := manager.SetAvatarByImage(avatar); !ok {
		Duder.Log(LogVerbose, "[Duder.SetAvatarByFile] Couldn't set avatar")
		return errors.New("couldn't set avatar")
	}

	return nil
}

// ParseArguments description
func (manager *DiscordManager) ParseArguments(content string) []string {
	content = strings.TrimSpace(content)
	inQuote := false
	var args []string
	var arg string

	for _, c := range content {
		if inQuote {
			if c == '"' {
				inQuote = false
				args = append(args, arg)
				arg = ""
			} else {
				arg = arg + string(c)
			}
		} else {
			if c == ' ' {
				if len(arg) > 0 {
					args = append(args, arg)
					arg = ""
				}
			} else if c == '"' {
				inQuote = true
			} else {
				arg = arg + string(c)
			}
		}
	}

	if len(arg) > 0 {
		args = append(args, arg)
	}

	return args
}

// onMessageCreate description
func (manager *DiscordManager) onMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	// stop playing with yourself!
	if message.Author.ID == manager.me.ID {
		return
	}

	// get message channel
	channel, ok := manager.GetMessageChannel(message)
	if !ok {
		return
	}

	// get message guild
	guild, ok := manager.GetMessageGuild(message)
	if ok {
		// message event
		Duder.Rugs.OnMessage(guild, message)
	}

	// check if the message has the command prefix
	if strings.HasPrefix(message.Content, fmt.Sprintf("%s ", Duder.Config.CommandPrefix())) {
		// strip the command prefix from the message content
		content := message.Content[len(Duder.Config.CommandPrefix())+1 : len(message.Content)]
		content = strings.TrimSpace(content)

		// get the root command
		args := manager.ParseArguments(content)
		if len(args) == 0 {
			return
		}
		cmd := strings.ToLower(args[0])
		Duder.Logf(LogVerbose, "Root command '%s'", cmd)

		Duder.Logf(LogVerbose, "Command in %s(%s:%s) from %s(%s): %s", channel.Name, channel.ID, manager.ChannelTypeName(channel), message.Author.Username, message.Author.ID, message.Content)
		// check for internal commands first
		if !manager.runCommand(message, cmd, args) {
			// run rug commands
			Duder.Rugs.RunCommand(message, cmd, args)
		}
	}
}

// onMessageReactionAdd description
func (manager *DiscordManager) onMessageReactionAdd(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	guild, ok := manager.GetGuildByChannelID(reaction.ChannelID)
	if !ok {
		return
	}
	message, err := manager.session.ChannelMessage(reaction.ChannelID, reaction.MessageID)
	if err != nil {
		return
	}
	instigator, ok := manager.GetGuildMember(guild.ID, reaction.UserID)
	if !ok {
		return
	}

	Duder.Rugs.OnMessageReactionAdd(guild, message, instigator.User, reaction.MessageReaction)
}

// onMessageReactionRemove description
func (manager *DiscordManager) onMessageReactionRemove(session *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	guild, ok := manager.GetGuildByChannelID(reaction.ChannelID)
	if !ok {
		return
	}
	message, err := manager.session.ChannelMessage(reaction.ChannelID, reaction.MessageID)
	if err != nil {
		return
	}
	instigator, ok := manager.GetGuildMember(guild.ID, reaction.UserID)
	if !ok {
		return
	}

	Duder.Rugs.OnMessageReactionRemove(guild, message, instigator.User, reaction.MessageReaction)
}

// onPresenceUpdate description
func (manager *DiscordManager) onPresenceUpdate(session *discordgo.Session, presence *discordgo.PresenceUpdate) {
	guild, ok := manager.GetGuildByID(presence.GuildID)
	if !ok {
		return
	}
	member, ok := manager.GetGuildMember(presence.GuildID, presence.User.ID)
	if !ok {
		return
	}

	Duder.Rugs.OnPresenceUpdate(guild, member.User, presence)
}

// runCommand description
func (manager *DiscordManager) runCommand(message *discordgo.MessageCreate, cmd string, args []string) bool {
	switch cmd {
	case "update":
		Duder.Update(message)
		return true
	case "shutdown":
		Duder.Shutdown(message)
		return true
	}

	return false
}

// teardown description
func (manager *DiscordManager) teardown() {
	manager.session.Close()
}
