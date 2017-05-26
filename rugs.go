package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/robertkrimen/otto"
)

var vm *otto.Otto

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

var rugs = map[string]Rug{}

func init() {
}

func createRugEnvironment() error {
	vm = otto.New()

	// create the rug environment
	vm.Set("createRug", createRug)
	vm.Set("addRugCommand", addRugCommand)
	if _, err := vm.Run(`
		function Rug(name, description) {
			this.name = name;
			createRug(this, name, description);
		}

		Rug.prototype.addCommand = function(trigger, exec) {
			addRugCommand(this, trigger, exec);
		}
	`); err != nil {
		return errors.New(fmt.Sprint("error creating rug environment: ", err.Error()))
	}

	// expose functions
	vm.Set("print", func(msg string) { fmt.Println(msg) })
	vm.Set("log", func(msg string) { log.Println(msg) })
	vm.Set("shutdown", Duder.Shutdown)

	return nil
}

// LoadRugs loads the rugs
func LoadRugs(path string) error {
	if err := createRugEnvironment(); err != nil {
		return err
	}

	// validate the rug path
	path = strings.TrimSpace(path)
	if len(path) == 0 {
		return errors.New("rug path is undefined")
	}

	log.Print("Loading rugs from folder: ", path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0644)
	}

	rugs = map[string]Rug{}

	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !strings.HasSuffix(f.Name(), ".js") {
			continue
		}
		Duder.DPrint("Loading rug file: ", f.Name())
		if buf, err := ioutil.ReadFile(fmt.Sprintf("%v/%v", path, f.Name())); err != nil {
			log.Print("Unable to load rug file", f.Name(), " reason ", err.Error())
		} else {
			s := string(buf)
			if _, err := vm.Run(s); err != nil {
				log.Print("Error loading rug ", err.Error())
			}
		}
	}

	return nil
}

func createRug(call otto.FunctionCall) otto.Value {
	obj := call.Argument(0).Object()
	name := call.Argument(1).String()
	description := call.Argument(2).String()

	rug := Rug{}
	rug.Name = name
	rug.Description = description
	rug.Commands = map[string]RugCommand{}
	rug.Object = obj

	rugs[fmt.Sprintf("%v", obj)] = rug

	Duder.DPrintf("Created rug '%v' %v", name, rug.Object)

	return otto.Value{}
}

func addRugCommand(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	trigger := call.Argument(1).String()
	exec := call.Argument(2).String()

	if rug, ok := rugs[fmt.Sprintf("%v", rugObj)]; ok {
		rugCmd := RugCommand{}
		rugCmd.Trigger = trigger
		rugCmd.Exec = fmt.Sprintf("cmd = %s; cmd()", exec)
		rug.Commands[trigger] = rugCmd
		Duder.DPrintf("Added command '%v' to rug '%v'", trigger, rug.Name)
	} else {
		Duder.DPrintf("Unable to add command to rug '%v'", rugObj)
	}
	return otto.Value{}
}

// OnMessageCreate description
func OnMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if strings.HasPrefix(message.Content, fmt.Sprintf("%s ", Duder.Config.Prefix)) {
		Duder.DPrint("command triggered")
		RunCommand(session, message)
	}
}

// RunCommand does the thing
func RunCommand(session *discordgo.Session, message *discordgo.MessageCreate) {
	content := message.Content[len(Duder.Config.Prefix)+1 : len(message.Content)]
	args := strings.Split(content, " ")
	Duder.DPrintf("root command '%s'", args[0])
	for _, rug := range rugs {
		for _, rugCmd := range rug.Commands {
			if rugCmd.Trigger == args[0] {
				execCommand(rug, rugCmd)
				return
			}
		}
	}
}

func execCommand(rug Rug, command RugCommand) {
	// set command environment variables
	vm.Set("rug", rug.Object)
	if _, err := vm.Run(command.Exec); err != nil {
		log.Print("Failed to run command ", err.Error())
	}
}
