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

// rugCommand defines the rug command
type rugCommand struct {
	trigger     string
	permissions []int
	exec        otto.Value
}

// Rug defines the rug
type Rug struct {
	name        string
	description string
	commands    map[string]rugCommand
	object      *otto.Object
	file        string
	teardown    func()
	loaded      time.Time
}

// rugMap contains all the rugs
var rugMap = map[string]Rug{}

// rugFile is the rug currently being loaded
var rugFile string

// rugLoadErrors in an array of errors
var rugLoadErrors []error

func init() {
	// create the JavaScript runtime
	Duder.jsvm = otto.New()

	if err := createRugEnvironment(); err != nil {
		log.Fatal("Unable to create Rug environment ", err.Error())
	}
}

// loadRugs loads all the Rugs from the Rug path
func loadRugs(path string) error {
	// validate the rug path
	path = strings.TrimSpace(path)
	if len(path) == 0 {
		return errors.New("Rug path is undefined")
	}

	log.Printf("Loading Rugs from folder '%v'", path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0777)
	}

	// clear the rugMap
	rugMap = map[string]Rug{}
	rugLoadErrors = rugLoadErrors[:0]

	// read the directory to get all the files
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		// ignore directories and non-javascript files
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".js") {
			continue
		}

		loadRug(fmt.Sprintf("%v/%v", path, f.Name()))
	}

	return nil
}

// loadRug description
func loadRug(file string) {
	// read the file
	Duder.dprintf("Loading Rug file '%v'", file)
	rugFile = file
	if buf, err := ioutil.ReadFile(rugFile); err != nil {
		Duder.wprint("Unable to load Rug file ", file, " reason ", err.Error())
	} else {
		s := string(buf)
		script := fmt.Sprintf("(function(){%s})()", s)
		if _, err := Duder.jsvm.Run(script); err != nil {
			Duder.wprint("Error loading Rug ", err.Error())
			rugLoadErrors = append(rugLoadErrors, err)
		}
	}
}

// observeRugs description
func observeRugs(path string) (*fsnotify.Watcher, error) {
	// validate the rug path
	path = strings.TrimSpace(path)
	if len(path) == 0 {
		return nil, errors.New("Rug path is undefined")
	}

	// creates a new file watcher
	rugWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("Unable to create Rug watcher: %v", err)
	}

	//go func(chan os.Signal) {
	go func() {
		for {
			select {
			case event := <-rugWatcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					// fix filename slashes
					file := filepath.ToSlash(event.Name)
					// make sure the file is .js
					if strings.HasSuffix(file, ".js") {
						Duder.dprintf("Rug file modified '%s'", file)
						if rug, ok := getRugByFile(file); ok {
							duration := time.Since(rug.loaded)
							if duration.Seconds() > 0.5 {
								key := getRugKey(rug)
								delete(rugMap, key)
								loadRug(file)
							}
						} else {
							Duder.dprint("Error finding rug", err)
						}
					}
				}
			}

		}
		//}(Duder.shutdownSignal)
	}()

	err = rugWatcher.Add(path)
	if err != nil {
		Duder.wprintf("Unable to watch Rugs path '%v': %v", path, err)
	}

	Duder.dprintf("Watching Rugs in path '%v'", path)

	return rugWatcher, nil
}

// getRugKey description
func getRugKey(rug Rug) string {
	return fmt.Sprintf("%v", rug.object)
}

// addRug description
func addRug(rug Rug) {
	Duder.dprint("Adding rug ", rug.file)
	rug.file = rugFile
	rug.loaded = time.Now()
	rugMap[getRugKey(rug)] = rug
}

// getRugByFile description
func getRugByFile(file string) (Rug, bool) {
	// check each rug to find the matching file
	for _, rug := range rugMap {
		if rug.file == file {
			return rug, true
		}
	}
	return Rug{}, false
}

// getRugByObject description
func getRugByObject(rugObj otto.Object) (Rug, bool) {
	if rug, ok := rugMap[fmt.Sprintf("%v", rugObj)]; ok {
		return rug, true
	}
	return Rug{}, false
}

// getRugStorageFile description
func getRugStorageFile(rug Rug) string {
	return strings.TrimSuffix(rug.file, filepath.Ext(rug.file)) + ".json"
}

// execRugCommand description
func execRugCommand(rug Rug, command rugCommand, session *discordgo.Session, message *discordgo.MessageCreate, args []string) {
	// get the message guild ID
	var guildID string
	if guild, err := getMessageGuild(session, message); err == nil {
		guildID = guild.ID
	}

	// create the author value
	cmdAuthor, err := Duder.jsvm.Call("new DuderUser", nil, message.ChannelID, message.Author.ID, message.Author.Username)
	if err != nil {
		Duder.wprint("Unable to convert author ", err.Error())
		return
	}

	// create the mentions array value
	cmdMentions, err := Duder.jsvm.Object(rugutils.ConvertMentions(message))
	if err != nil {
		Duder.wprint("Unable to convert mentions", err.Error())
		return
	}

	// create the arguments array value
	cmdArgs, err := Duder.jsvm.Object(rugutils.ConvertArgs(args))
	if err != nil {
		Duder.wprint("Unable to convert arguments", err.Error())
		return
	}

	// create the command
	cmd, err := Duder.jsvm.Call("new DuderCommand", nil, guildID, message.ChannelID, message.ID, cmdAuthor, cmdMentions, cmdArgs)
	if err != nil {
		Duder.wprint("Unable to make command", err.Error())
	}

	// execute the command
	if _, err := command.exec.Call(rug.object.Value(), cmd); err != nil {
		Duder.wprint("Unable to run command", err.Error())
	}

	// check for any leaked variables (strict mode; weren't declared) and delete them
	if result, err := Duder.jsvm.Run(`
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
		Duder.wprint("Failed to check for leaks", err.Error())
	} else {
		var leaks []string
		export, _ := result.Export()
		{
			leaks, _ = export.([]string)
		}
		if len(leaks) > 0 {
			for _, leak := range leaks {
				Duder.wprintf("%s leaked variable '%s'", rug.file, leak)
			}
		}
	}
	// clean up
	Duder.jsvm.Run(`delete __leaks__;`)
}

func (rug *Rug) logPrefix() string {
	return fmt.Sprintf("[Rug:%s(%s)]", rug.name, rug.file)
}
