package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/robertkrimen/otto"
)

var vm *otto.Otto

// RugCommand defines the rug command
type RugCommand struct {
	Trigger string
	Exec    string
}

// Rug defines the rug
type Rug struct {
	Name        string
	Description string
	Commands    map[string]RugCommand
	Object      *otto.Object
}

var rugs = map[string]Rug{}

func init() {
	vm = otto.New()

	// add rug creation
	vm.Set("createRug", createRug)
	vm.Set("addRugCommand", addRugCommand)
	if _, err := vm.Run(`
		function Rug(name, description) {
			this.name = name;
			createRug(this, name, description);
		}

		Rug.prototype.addCommand = function(trigger, cmd) {
			addRugCommand(this.name, trigger, cmd);
		}
	`); err != nil {
		log.Print("error initializing rugs ", err.Error())
	}

	// expose functions
	vm.Set("print", func(msg string) { log.Print(msg) })
}

// LoadRugs loads the rugs
func LoadRugs(path string) {
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
		Duder.DPrint("loading rug file: ", f.Name())
		if buf, err := ioutil.ReadFile(fmt.Sprintf("%v/%v", path, f.Name())); err != nil {
			log.Print("unable to load rug file", f.Name(), " reason ", err.Error())
		} else {
			s := string(buf)
			if _, err := vm.Run(s); err != nil {
				log.Print("error loading rug ", err.Error())
			}
		}
	}
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

	rugs[name] = rug

	Duder.DPrintf("created rug '%v'", name)

	return otto.Value{}
}

func addRugCommand(call otto.FunctionCall) otto.Value {
	rugName := call.Argument(0).String()
	trigger := call.Argument(1).String()
	exec := call.Argument(2).String()
	if rug, ok := rugs[rugName]; ok {
		rugCmd := RugCommand{}
		rugCmd.Trigger = trigger
		rugCmd.Exec = fmt.Sprintf("cmd = %s; cmd()", exec)
		rug.Commands[trigger] = rugCmd
		Duder.DPrintf("added command '%v' to rug '%v'", trigger, rugName)
	} else {

	}
	return otto.Value{}
}

// RunCommand does the thing
func RunCommand(trigger string) {
	for _, rug := range rugs {
		for _, rugCmd := range rug.Commands {
			if rugCmd.Trigger == trigger {
				execCommand(rug, rugCmd)
				return
			}
		}
	}
}

func execCommand(rug Rug, command RugCommand) {
	vm.Set("rug", rug.Object)
	if _, err := vm.Run(command.Exec); err != nil {
		log.Print("failed to run command ", err.Error())
	}
}
