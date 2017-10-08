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

	"github.com/bwmarrin/discordgo"
	"github.com/foszor/duder/helpers/rugutils"
	"github.com/go-fsnotify/fsnotify"
	"github.com/robertkrimen/otto"
)

// rugCommand defines the rug command
type rugCommand struct {
	trigger     string
	exec        string
	permissions []int
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

// js is the JavaScript runtime
var js *otto.Otto

//
var rugFile string

var rugLoadErrors []error

func init() {
	// create the JavaScript runtime
	js = otto.New()

	if err := createRugEnvironment(); err != nil {
		log.Fatal("Unable to create Rug environment", err.Error())
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
		os.Mkdir(path, 0644)
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
		log.Print("Unable to load Rug file ", file, " reason ", err.Error())
	} else {
		s := string(buf)
		if _, err := js.Run(fmt.Sprintf("__rbox = function(){ %s }; __rbox();", s)); err != nil {
			log.Print("Error loading Rug ", err.Error())
			rugLoadErrors = append(rugLoadErrors, err)
		}
	}
}

// watchRugs description
func watchRugs(path string) (*fsnotify.Watcher, error) {
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
						//log.Println("file modified:", file)
						if rug, e := getRugByFile(file); e == nil {
							//log.Print("rug loaded at ", rug.loaded, " current time is ", time.Now())
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
		Duder.dprintf("Unable to watch Rugs path '%v': %v", path, err)
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
	rug.file = rugFile
	rug.loaded = time.Now()
	log.Print("Adding rug ", rug.file)
	rugMap[getRugKey(rug)] = rug
}

// getRugByFile description
func getRugByFile(file string) (Rug, error) {
	// check each rug to find the matching file
	for _, rug := range rugMap {
		if rug.file == file {
			return rug, nil
		}
	}
	return Rug{}, errors.New("Unable to find rug")
}

// execRugCommand description
func execRugCommand(rug Rug, command rugCommand, session *discordgo.Session, message *discordgo.MessageCreate, args []string) {
	// set command environment variables
	js.Set("rug", rug.object)
	if _, err := js.Run(fmt.Sprintf(
		`
			var cmd = new DuderCommand();
			cmd.channelID = "%s";
			cmd.author = new DuderUser("%s", "%s");
			cmd.mentions = %s;
			cmd.args = %s;
			`,
		message.ChannelID,
		message.Author.ID,
		message.Author.Username,
		rugutils.ConvertMentions(message),
		rugutils.ConvertArgs(args))); err != nil {
		log.Print("Failed to set command enviroment ", err.Error())
	}

	if _, err := js.Run(command.exec); err != nil {
		log.Print("Failed to run command ", err.Error())
	}
}
