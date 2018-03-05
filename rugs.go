package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/foszor/duder/helpers/rugutils"

	"github.com/bwmarrin/discordgo"
	"github.com/go-fsnotify/fsnotify"
	"github.com/robertkrimen/otto"
)

func init() {
	Duder.Rugs = RugManager{}
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
	if guild, ok := Duder.Discord.MessageGuild(message); ok {
		guildID = guild.ID
	}

	// create the author value
	cmdAuthor, err := Duder.Rugs.VM.Call("new DuderUser", nil, guildID, message.Author.ID, message.Author.Username)
	if err != nil {
		Duder.Log(LogChannel.Warning, "[RugCommand.Run] Unable to convert author", err.Error())
		return
	}

	// create the mentions array value
	cmdMentions, err := Duder.Rugs.VM.Object(rugutils.ConvertMentions(message))
	if err != nil {
		Duder.Log(LogChannel.Warning, "[RugCommand.Run] Unable to convert mentions", err.Error())
		return
	}

	// create the arguments array value
	cmdArgs, err := Duder.Rugs.VM.Object(rugutils.ConvertArguments(args))
	if err != nil {
		Duder.Log(LogChannel.Warning, "[RugCommand.Run] Unable to convert arguments", err.Error())
		return
	}

	// create the command
	cmd, err := Duder.Rugs.VM.Call("new DuderCommand", nil, guildID, message.ChannelID, message.ID, cmdAuthor, cmdMentions, cmdArgs)
	if err != nil {
		Duder.Log(LogChannel.Warning, "[RugCommand.Run] Unable to create command", err.Error())
	}

	// execute the command
	if _, err := rugCmd.Exec.Call(rug.Object.Value(), cmd); err != nil {
		Duder.Log(LogChannel.Warning, "[RugCommand.Run] Unable to run command", err.Error())
	}

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
		Duder.Log(LogChannel.Warning, "[RugCommand.Run] Failed to check for leaks", err.Error())
	} else {
		var leaks []string
		export, _ := result.Export()
		{
			leaks, _ = export.([]string)
		}
		if len(leaks) > 0 {
			for _, leak := range leaks {
				Duder.Logf(LogChannel.Warning, "[RugCommand.Run] %s leaked variable '%s'", rug.File, leak)
			}
		}
	}
	// clean up
	Duder.Rugs.VM.Run(`delete __leaks__;`)
}

// Rug defines the rug
type Rug struct {
	Commands    map[string]RugCommand
	Description string
	File        string
	Loaded      time.Time
	Name        string
	Object      *otto.Object
}

// AddCommand description
func (rug *Rug) AddCommand(trigger string, exec otto.Value) {
	rugCmd := RugCommand{
		Trigger: trigger,
		Exec:    exec,
	}
	rug.Commands[trigger] = rugCmd
	Duder.Logf(LogChannel.Verbose, "[Rug.AddCommand] Added command '%s' to rug '%s'", trigger, rug.Name)
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
	Duder.Logf(LogChannel.Verbose, "%s %s", rug.LogPrefix(), msg)
}

// WPrint description
func (rug *Rug) WPrint(msg string) {
	Duder.Logf(LogChannel.Warning, "%s %s", rug.LogPrefix(), msg)
}

// StorageFile description
func (rug *Rug) StorageFile() string {
	return strings.TrimSuffix(rug.File, filepath.Ext(rug.File)) + ".json"
}

// LoadStorage description
func (rug *Rug) LoadStorage() (string, bool) {
	Duder.Logf(LogChannel.Verbose, "[Rug.LoadStorage] Loading storage for rug '%s'", rug.Name)

	path := rug.StorageFile()

	// check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		Duder.Logf(LogChannel.Verbose, "[Rug.LoadStorage] Storage file for rug '%s' not found; creating new one...", rug.Name)

		// create the storage file
		if e := ioutil.WriteFile(path, []byte("{}"), 0777); e != nil {
			Duder.Logf(LogChannel.Verbose, "[Rug.LoadStorage] Unable to create storage file for rug '%s'", rug.Name)
			return "{}", false
		}

		// return empty storage
		Duder.Logf(LogChannel.Verbose, "[Rug.LoadStorage] Successfully created storage file for rug '%s'", rug.Name)
		return "{}", true
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		Duder.Logf(LogChannel.Verbose, "[Rug.LoadStorage] Unable to read storage file for rug '%s'", rug.Name)
		return "{}", false
	}

	return string(bytes), true
}

// SaveStorage description
func (rug *Rug) SaveStorage(data string) bool {
	Duder.Logf(LogChannel.Verbose, "[Rug.SaveStorage] Saving storage for rug '%s'", rug.Name)

	path := rug.StorageFile()

	if err := ioutil.WriteFile(path, []byte(data), 0777); err != nil {
		Duder.Logf(LogChannel.Verbose, "[Rug.LoadStorage] Unable to write storage file for rug '%s'; %s", rug.Name, err.Error())
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

	Duder.Logf(LogChannel.General, "Loading Rugs from folder '%v'", path)

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
							Duder.Logf(LogChannel.Verbose, "[RugManager.Watcher] Rug file '%s' was modified", file)
							if rug, ok := manager.FindRugByFile(file); ok {
								duration := time.Since(rug.Loaded)
								if duration.Seconds() > 0.5 {
									delete(manager.Rugs, rug.Key())
									manager.LoadRug(file)
								}
							} else {
								Duder.Logf(LogChannel.Warning, "[RugManager.Watcher] Error finding rug for file '%s'", file)
							}
						}
					}
				}

			}
		}()

		err = manager.Watcher.Add(path)
		if err != nil {
			Duder.Logf(LogChannel.Warning, "[RugManager.Watcher] Unable to watch rugs path '%s': %s", path, err.Error())
		} else {
			Duder.Logf(LogChannel.Verbose, "[RugManager.Watcher] Watching rugs in path '%s'", path)
		}
	}

	// clear the rugMap
	manager.Rugs = map[string]Rug{}
	manager.loadErrors = manager.loadErrors[:0]

	// read the directory to get all the files
	files, _ := ioutil.ReadDir(path)
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
	Duder.Logf(LogChannel.Verbose, "[RugManager.LoadRug] Loading rug file '%s'", file)
	manager.loadFile = file
	if buf, err := ioutil.ReadFile(file); err != nil {
		Duder.Logf(LogChannel.Warning, "[RugManager.LoadRug] Unable to read rug file '%s': '%s'", file, err.Error())
	} else {
		s := string(buf)
		script := fmt.Sprintf("(function(){%s})()", s)
		if _, err := manager.VM.Run(script); err != nil {
			Duder.Logf(LogChannel.Warning, "[RugManager.LoadRug] Error loading rug file '%s': '%s'", file, err.Error())
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
	manager.Rugs[rug.Key()] = rug

	Duder.Logf(LogChannel.Verbose, "[RugManager.CreateRug] Created rug '%s' from file '%s'", rug.Name, rug.File)
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

// teardown description
func (manager *RugManager) teardown() {
	if manager.Watcher != nil && manager.WatcherEnabled {
		defer manager.Watcher.Close()
	}
}
