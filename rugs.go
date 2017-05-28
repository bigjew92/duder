package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/foszor/duder/helpers/rugutils"
	"github.com/robertkrimen/otto"
)

// RugCommand defines the rug command
type RugCommand struct {
	Trigger     string
	Exec        string
	Permissions []int
}

// Rug defines the rug
type Rug struct {
	Name        string
	Description string
	Commands    map[string]RugCommand
	Object      *otto.Object
	teardown    func()
}

// rugMap contains all the rugs
var rugMap = map[string]Rug{}

// js is the JavaScript runtime
var js *otto.Otto

func init() {
	// create the JavaScript runtime
	js = otto.New()

	if err := createRugEnvironment(); err != nil {
		log.Fatal("Unable to create Rug environment", err.Error())
	}
}

var loadErrors []error

// LoadRugs loads all the Rugs from the Rug path
func LoadRugs(path string) error {
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
	loadErrors = loadErrors[:0]

	// read the directory to get all the files
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		// ignore directories and non-javascript files
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".js") {
			continue
		}

		// read the file
		Duder.DPrintf("Loading Rug file '%v'", f.Name())
		if buf, err := ioutil.ReadFile(fmt.Sprintf("%v/%v", path, f.Name())); err != nil {
			log.Print("Unable to load Rug file ", f.Name(), " reason ", err.Error())
		} else {
			s := string(buf)
			if _, err := js.Run(fmt.Sprintf("__rbox = function(){ %s }; __rbox();", s)); err != nil {
				//if _, err := js.Run(s); err != nil {
				log.Print("Error loading Rug ", err.Error())
				loadErrors = append(loadErrors, err)
			}
		}
	}

	return nil
}

// RunCommand description
func RunCommand(session *discordgo.Session, message *discordgo.MessageCreate) {
	// strip the command prefix from the message content
	content := message.Content[len(Duder.Config.Prefix)+1 : len(message.Content)]
	content = strings.TrimSpace(content)

	// get the root command
	args := rugutils.ParseArguments(content)
	if len(args) == 0 {
		return
	}
	Duder.DPrintf("Root command '%s'", args[0])

	// core commands
	if message.Author.ID == Duder.Config.OwnerID {
		if strings.EqualFold("reload", args[0]) {
			LoadRugs(Duder.Config.RugPath)
			if len(loadErrors) > 0 {
				session.ChannelMessageSend(message.ChannelID, "Rugs reloaded with errors.")
			} else {
				session.ChannelMessageSend(message.ChannelID, "Rugs successfully reloaded.")
			}
			return
		} else if strings.EqualFold("shutdown", args[0]) {
			session.ChannelMessageSend(message.ChannelID, "Goodbye.")
			Duder.Shutdown()
			return
		}
	}

	// check each rug to find the matching command
	for _, rug := range rugMap {
		for _, rugCmd := range rug.Commands {
			if strings.EqualFold(rugCmd.Trigger, args[0]) {
				execCommand(rug, rugCmd, session, message, args)
				return
			}
		}
	}
}

// execCommand description
func execCommand(rug Rug, command RugCommand, session *discordgo.Session, message *discordgo.MessageCreate, args []string) {
	// set command environment variables
	js.Set("rug", rug.Object)
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

	if _, err := js.Run(command.Exec); err != nil {
		log.Print("Failed to run command ", err.Error())
	}
}
