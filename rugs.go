package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bigjew92/duder/helpers/rugutils"

	"github.com/bwmarrin/discordgo"
	"github.com/fsnotify/fsnotify"
	"github.com/robertkrimen/otto"
)

func init() {
	Duder.Rugs = new(RugManager)
	// create the JavaScript runtime
	Duder.Rugs.VM = otto.New()

	if err := createRugEnvironment(); err != nil {
		log.Fatal("Unable to create Rug environment ", err.Error())
	}
}

// RugCommand defines the rug command
type RugCommand struct {
	Trigger string
	Exec    otto.Value
}

// Run description
func (rugCmd *RugCommand) Run(rug Rug, message *discordgo.MessageCreate, args []string) {
	// get the message guild ID
	var guildID string
	if guild, ok := Duder.Discord.GetMessageGuild(message); ok {
		guildID = guild.ID
	}

	// create the author value
	author, err := Duder.Rugs.VM.Call("new DuderUser", nil, guildID, message.Author.ID, message.Author.Username)
	if err != nil {
		Duder.Log(LogWarning, "[RugCommand.Run] Unable to convert author", err.Error())
		return
	}

	// create the mentions array value
	mentions, err := Duder.Rugs.VM.Object(rugutils.ConvertMentions(guildID, message))
	if err != nil {
		Duder.Log(LogWarning, "[RugCommand.Run] Unable to convert mentions", err.Error())
		return
	}

	// create the arguments array value
	cmdArgs, err := Duder.Rugs.VM.Object(rugutils.ConvertArguments(args))
	if err != nil {
		Duder.Log(LogWarning, "[RugCommand.Run] Unable to convert arguments", err.Error())
		return
	}

	// create the command
	cmd, err := Duder.Rugs.VM.Call("new DuderCommand", nil, guildID, message.ChannelID, message.ID, author, mentions, cmdArgs)
	if err != nil {
		Duder.Log(LogWarning, "[RugCommand.Run] Unable to create command", err.Error())
	}

	// execute the command
	if _, err := rugCmd.Exec.Call(rug.Object.Value(), cmd); err != nil {
		Duder.Log(LogWarning, "[RugCommand.Run] Unable to run command", err.Error())
	}

	rug.cleanLeaks()
}

// RugEventHandler struct
type RugEventHandler struct {
	Exec otto.Value
}

// Rug defines the rug
type Rug struct {
	Commands                        map[string]RugCommand
	Description                     string
	File                            string
	Loaded                          time.Time
	Name                            string
	Object                          *otto.Object
	OnMessageHandlers               map[int]RugEventHandler
	OnMessageReactionAddHandlers    map[int]RugEventHandler
	OnMessageReactionRemoveHandlers map[int]RugEventHandler
	OnPresenceUpdateHandlers        map[int]RugEventHandler
}

// AddCommand description
func (rug *Rug) AddCommand(trigger string, exec otto.Value) {
	rugCmd := RugCommand{
		Trigger: trigger,
		Exec:    exec,
	}
	rug.Commands[trigger] = rugCmd
	Duder.Logf(LogVerbose, "[Rug.AddCommand] Added command '%s' to rug '%s'", trigger, rug.Name)
}

// BindOnMessage description
func (rug *Rug) BindOnMessage(onMessage otto.Value) {
	handler := RugEventHandler{
		Exec: onMessage,
	}
	rug.OnMessageHandlers[len(rug.OnMessageHandlers)] = handler
	Duder.Logf(LogVerbose, "[Rug.BindOnMessage] Rug '%s' added an event delegate", rug.Name)
}

// CallOnMessage description
func (rug *Rug) CallOnMessage(guild *discordgo.Guild, message *discordgo.MessageCreate, msg otto.Value) {
	if len(rug.OnMessageHandlers) == 0 {
		return
	}

	// call events
	for i := 0; i < len(rug.OnMessageHandlers); i++ {
		handler := rug.OnMessageHandlers[i]
		if _, err := handler.Exec.Call(rug.Object.Value(), msg); err != nil {
			Duder.Log(LogWarning, "[Rug.CallOnMessage] Unable to call event handler", err.Error())
		}
	}

	rug.cleanLeaks()
}

// BindOnMessageReactionAdd description
func (rug *Rug) BindOnMessageReactionAdd(onMessageReactionAdd otto.Value) {
	handler := RugEventHandler{
		Exec: onMessageReactionAdd,
	}
	rug.OnMessageReactionAddHandlers[len(rug.OnMessageReactionAddHandlers)] = handler
	Duder.Logf(LogVerbose, "[Rug.BindOnMessageReactionAdd] Rug '%s' added an event delegate", rug.Name)
}

// CallOnMessageReactionAdd description
func (rug *Rug) CallOnMessageReactionAdd(reactionAdd otto.Value) {
	if len(rug.OnMessageReactionAddHandlers) == 0 {
		return
	}

	// call events
	for i := 0; i < len(rug.OnMessageReactionAddHandlers); i++ {
		handler := rug.OnMessageReactionAddHandlers[i]
		if _, err := handler.Exec.Call(rug.Object.Value(), reactionAdd); err != nil {
			Duder.Log(LogWarning, "[Rug.CallOnMessageReactionAdd] Unable to call event handler", err.Error())
		}
	}

	rug.cleanLeaks()
}

// BindOnMessageReactionRemove description
func (rug *Rug) BindOnMessageReactionRemove(onMessageReactionRemove otto.Value) {
	handler := RugEventHandler{
		Exec: onMessageReactionRemove,
	}
	rug.OnMessageReactionRemoveHandlers[len(rug.OnMessageReactionRemoveHandlers)] = handler
	Duder.Logf(LogVerbose, "[Rug.BindOnMessageReactionRemove] Rug '%s' added an event delegate", rug.Name)
}

// CallOnMessageReactionRemove description
func (rug *Rug) CallOnMessageReactionRemove(reactionRemove otto.Value) {
	if len(rug.OnMessageReactionRemoveHandlers) == 0 {
		return
	}

	// call events
	for i := 0; i < len(rug.OnMessageReactionRemoveHandlers); i++ {
		handler := rug.OnMessageReactionRemoveHandlers[i]
		if _, err := handler.Exec.Call(rug.Object.Value(), reactionRemove); err != nil {
			Duder.Log(LogWarning, "[Rug.CallOnMessageReactionRemove] Unable to call event handler", err.Error())
		}
	}

	rug.cleanLeaks()
}

// BindOnPresenceUpdate description
func (rug *Rug) BindOnPresenceUpdate(onPresence otto.Value) {
	handler := RugEventHandler{
		Exec: onPresence,
	}
	rug.OnPresenceUpdateHandlers[len(rug.OnPresenceUpdateHandlers)] = handler
	Duder.Logf(LogVerbose, "[Rug.BindOnPresenceUpdate] Rug '%s' added an event delegate", rug.Name)
}

// CallOnPresenceUpdate description
func (rug *Rug) CallOnPresenceUpdate(guild *discordgo.Guild, user otto.Value, presence *discordgo.PresenceUpdate) {
	if len(rug.OnPresenceUpdateHandlers) == 0 {
		return
	}

	// TODO: provide more information
	status := string(presence.Status)

	// call events
	for i := 0; i < len(rug.OnPresenceUpdateHandlers); i++ {
		handler := rug.OnPresenceUpdateHandlers[i]
		if _, err := handler.Exec.Call(rug.Object.Value(), guild.ID, user, status); err != nil {
			Duder.Log(LogWarning, "[Rug.CallOnPresenceUpdate] Unable to call event handler", err.Error())
		}
	}

	rug.cleanLeaks()
}

// cleanLeaks description
func (rug *Rug) cleanLeaks() {
	// check for any leaked variables (strict mode; weren't declared) and delete them
	if result, err := Duder.Rugs.VM.Run(`
			(function() {
				var __leaks__ = [];
				for (var __n__ in this) {
					if (typeof this[__n__] == "function") {
						continue;
					}
					if (__n__ != "console" && __n__ != "__n__" && __n__ != "__leaks__") {
						__leaks__.push(__n__);
						delete this[__n__];
					}
				}
				return __leaks__;
			})()`); err != nil {
		Duder.Log(LogWarning, "[Rug.CleanLeaks] Failed to check for leaks", err.Error())
	} else {
		var leaks []string
		export, _ := result.Export()
		{
			leaks, _ = export.([]string)
		}
		if len(leaks) > 0 {
			for _, leak := range leaks {
				Duder.Logf(LogWarning, "[Rug.CleanLeaks] '%s' leaked variable '%s'", rug.File, leak)
			}
		}
	}
	// clean up
	Duder.Rugs.VM.Run(`delete __leaks__;`)
}

// Key description
func (rug *Rug) Key() string {
	return fmt.Sprintf("%v", rug.Object)
}

// LogPrefix description
func (rug *Rug) LogPrefix() string {
	return fmt.Sprintf("[Rug:%s(%s)]", rug.Name, rug.File)
}

// DPrint description
func (rug *Rug) DPrint(msg string) {
	Duder.Logf(LogVerbose, "%s %s", rug.LogPrefix(), msg)
}

// WPrint description
func (rug *Rug) WPrint(msg string) {
	Duder.Logf(LogWarning, "%s %s", rug.LogPrefix(), msg)
}

// StorageFile description
func (rug *Rug) StorageFile() string {
	return strings.TrimSuffix(rug.File, filepath.Ext(rug.File)) + ".json"
}

// LoadStorage description
func (rug *Rug) LoadStorage() (string, bool) {
	Duder.Logf(LogVerbose, "[Rug.LoadStorage] Loading storage for rug '%s'", rug.Name)

	path := rug.StorageFile()

	// check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		Duder.Logf(LogVerbose, "[Rug.LoadStorage] Storage file for rug '%s' not found; creating new one...", rug.Name)

		// create the storage file
		if e := os.WriteFile(path, []byte("{}"), 0777); e != nil {
			Duder.Logf(LogVerbose, "[Rug.LoadStorage] Unable to create storage file for rug '%s'", rug.Name)
			return "{}", false
		}

		// return empty storage
		Duder.Logf(LogVerbose, "[Rug.LoadStorage] Successfully created storage file for rug '%s'", rug.Name)
		return "{}", true
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		Duder.Logf(LogVerbose, "[Rug.LoadStorage] Unable to read storage file for rug '%s'", rug.Name)
		return "{}", false
	}

	return string(bytes), true
}

// SaveStorage description
func (rug *Rug) SaveStorage(data string) bool {
	Duder.Logf(LogVerbose, "[Rug.SaveStorage] Saving storage for rug '%s'", rug.Name)

	path := rug.StorageFile()

	if err := os.WriteFile(path, []byte(data), 0777); err != nil {
		Duder.Logf(LogVerbose, "[Rug.LoadStorage] Unable to write storage file for rug '%s'; %s", rug.Name, err.Error())
		return false
	}

	return true
}

// RugManager defines the rug manager
type RugManager struct {
	Rugs           map[string]Rug
	VM             *otto.Otto
	Watcher        *fsnotify.Watcher
	WatcherEnabled bool
	loadErrors     []error
	loadFile       string
}

// Load loads all the Rugs from the Rug path
func (manager *RugManager) Load() error {
	path := strings.TrimSpace(Duder.Config.RugsPath())
	if len(path) == 0 {
		return errors.New("rugs path isn't defined")
	}

	Duder.Logf(LogGeneral, "Loading Rugs from folder '%v'", path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0777)
	}

	watcher, err := fsnotify.NewWatcher()
	if err == nil {
		manager.Watcher = watcher
		manager.WatcherEnabled = true
	}

	if manager.WatcherEnabled {
		go func() {
			for {
				select {
				case event := <-manager.Watcher.Events:
					if event.Op&fsnotify.Write == fsnotify.Write {
						// fix filename slashes
						file := filepath.ToSlash(event.Name)
						// make sure the file is .js
						if strings.HasSuffix(file, ".js") {
							Duder.Logf(LogVerbose, "[RugManager.Watcher] Rug file '%s' was modified", file)
							if rug, ok := manager.FindRugByFile(file); ok {
								duration := time.Since(rug.Loaded)
								if duration.Seconds() > 0.5 {
									delete(manager.Rugs, rug.Key())
									manager.LoadRug(file)
								}
							} else {
								Duder.Logf(LogWarning, "[RugManager.Watcher] Error finding rug for file '%s'", file)
							}
						}
					}
				}

			}
		}()

		err = manager.Watcher.Add(path)
		if err != nil {
			Duder.Logf(LogWarning, "[RugManager.Watcher] Unable to watch rugs path '%s': %s", path, err.Error())
		} else {
			Duder.Logf(LogVerbose, "[RugManager.Watcher] Watching rugs in path '%s'", path)
		}
	}

	// clear the rugMap
	manager.Rugs = map[string]Rug{}
	manager.loadErrors = manager.loadErrors[:0]

	// read the directory to get all the files
	files, _ := os.ReadDir(path)
	for _, f := range files {
		// ignore directories and non-javascript files
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".js") {
			continue
		}

		manager.LoadRug(fmt.Sprintf("%s/%s", path, f.Name()))
	}

	return nil
}

// LoadRug description
func (manager *RugManager) LoadRug(file string) {
	// read the file
	Duder.Logf(LogVerbose, "[RugManager.LoadRug] Loading rug file '%s'", file)
	manager.loadFile = file
	if buf, err := os.ReadFile(file); err != nil {
		Duder.Logf(LogWarning, "[RugManager.LoadRug] Unable to read rug file '%s': '%s'", file, err.Error())
	} else {
		s := string(buf)
		script := fmt.Sprintf("(function(){%s})()", s)
		if _, err := manager.VM.Run(script); err != nil {
			Duder.Logf(LogWarning, "[RugManager.LoadRug] Error loading rug file '%s': '%s'", file, err.Error())
			manager.loadErrors = append(manager.loadErrors, err)
		}
	}
}

// CreateRug description
func (manager *RugManager) CreateRug(rugObj *otto.Object, name string, description string) {
	rug := Rug{}
	rug.Name = name
	rug.Description = description
	rug.Commands = map[string]RugCommand{}
	rug.Object = rugObj
	rug.File = manager.loadFile
	rug.Loaded = time.Now()
	rug.OnMessageHandlers = map[int]RugEventHandler{}
	rug.OnMessageReactionAddHandlers = map[int]RugEventHandler{}
	rug.OnMessageReactionRemoveHandlers = map[int]RugEventHandler{}
	rug.OnPresenceUpdateHandlers = map[int]RugEventHandler{}
	manager.Rugs[rug.Key()] = rug

	Duder.Logf(LogVerbose, "[RugManager.CreateRug] Created rug '%s' from file '%s'", rug.Name, rug.File)
}

// FindRugByFile description
func (manager *RugManager) FindRugByFile(file string) (Rug, bool) {
	// check each rug to find the matching file
	for _, rug := range manager.Rugs {
		if rug.File == file {
			return rug, true
		}
	}
	return Rug{}, false
}

// FindRugByObject description
func (manager *RugManager) FindRugByObject(rugObj *otto.Object) (Rug, bool) {
	if rug, ok := manager.Rugs[fmt.Sprintf("%v", rugObj)]; ok {
		return rug, true
	}
	return Rug{}, false
}

// RunCommand description
func (manager *RugManager) RunCommand(message *discordgo.MessageCreate, cmd string, args []string) {
	// check each rug to find the matching command
	for _, rug := range manager.Rugs {
		for _, rugCmd := range rug.Commands {
			if rugCmd.Trigger == cmd {
				rugCmd.Run(rug, message, args)
				return
			}
		}
	}
}

// OnMessage description
func (manager *RugManager) OnMessage(guild *discordgo.Guild, message *discordgo.MessageCreate) {
	// create the author
	msgAuthor, err := Duder.Rugs.VM.Call("new DuderUser", nil, guild.ID, message.Author.ID, message.Author.Username)
	if err != nil {
		Duder.Log(LogWarning, "[RugManager.OnMessage] Unable to create author", err.Error())
		return
	}

	// create the message
	msg, err := Duder.Rugs.VM.Call("new DuderMessage", nil, guild.ID, message.ChannelID, msgAuthor, message.ID, message.Content)
	if err != nil {
		Duder.Log(LogWarning, "[RugManager.OnMessage] Unable to create message", err.Error())
		return
	}

	for _, rug := range manager.Rugs {
		rug.CallOnMessage(guild, message, msg)
	}
}

// OnPresenceUpdate description
func (manager *RugManager) OnPresenceUpdate(guild *discordgo.Guild, user *discordgo.User, presence *discordgo.PresenceUpdate) {
	// create the user
	presenceUser, err := Duder.Rugs.VM.Call("new DuderUser", nil, guild.ID, user.ID, user.Username)
	if err != nil {
		Duder.Log(LogWarning, "[RugManager.OnPresenceUpdate] Unable to create user", err.Error())
		return
	}

	for _, rug := range manager.Rugs {
		rug.CallOnPresenceUpdate(guild, presenceUser, presence)
	}
}

// createReaction description
func (manager *RugManager) createReaction(guild *discordgo.Guild, message *discordgo.Message, reactionInstigator *discordgo.User, reaction *discordgo.MessageReaction, add bool) (otto.Value, bool) {
	// create the emoji value
	emoji, err := Duder.Rugs.VM.Call("new DuderEmoji", nil, reaction.Emoji.ID, reaction.Emoji.Name, []string{}, reaction.Emoji.Managed, reaction.Emoji.RequireColons, reaction.Emoji.Animated)
	if err != nil {
		Duder.Log(LogWarning, "[RugManager.createReaction] Unable to create emoji", err.Error())
		return otto.Value{}, false
	}

	// create the author value
	msgAuthor, err := Duder.Rugs.VM.Call("new DuderUser", nil, guild.ID, message.Author.ID, message.Author.Username)
	if err != nil {
		Duder.Log(LogWarning, "[RugManager.createReaction] Unable to create author", err.Error())
		return otto.Value{}, false
	}

	// create the message value
	msg, err := Duder.Rugs.VM.Call("new DuderMessage", nil, guild.ID, message.ChannelID, msgAuthor, message.ID, message.Content)
	if err != nil {
		Duder.Log(LogWarning, "[RugManager.createReaction] Unable to create message", err.Error())
		return otto.Value{}, false
	}

	// create the instigator value
	instigator, err := Duder.Rugs.VM.Call("new DuderUser", nil, guild.ID, reactionInstigator.ID, reactionInstigator.Username)
	if err != nil {
		Duder.Log(LogWarning, "[RugManager.createReaction] Unable to create instigator", err.Error())
		return otto.Value{}, false
	}

	// create the reaction value
	r, err := Duder.Rugs.VM.Call("new DuderMessageReaction", nil, guild.ID, reaction.ChannelID, msg, instigator, emoji, add)
	if err != nil {
		Duder.Log(LogWarning, "[RugManager.createReaction] Unable to create emoji", err.Error())
		return otto.Value{}, false
	}

	return r, true
}

// OnMessageReactionAdd description
func (manager *RugManager) OnMessageReactionAdd(guild *discordgo.Guild, message *discordgo.Message, instigator *discordgo.User, reactionAdd *discordgo.MessageReaction) {
	reaction, ok := manager.createReaction(guild, message, instigator, reactionAdd, true)
	if !ok {
		return
	}

	for _, rug := range manager.Rugs {
		rug.CallOnMessageReactionAdd(reaction)
	}
}

// OnMessageReactionRemove description
func (manager *RugManager) OnMessageReactionRemove(guild *discordgo.Guild, message *discordgo.Message, instigator *discordgo.User, reactionRemove *discordgo.MessageReaction) {
	reaction, ok := manager.createReaction(guild, message, instigator, reactionRemove, false)
	if !ok {
		return
	}

	for _, rug := range manager.Rugs {
		rug.CallOnMessageReactionRemove(reaction)
	}
}

// teardown description
func (manager *RugManager) teardown() {
	if manager.Watcher != nil && manager.WatcherEnabled {
		defer manager.Watcher.Close()
	}
}
