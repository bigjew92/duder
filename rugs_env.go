package main

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/robertkrimen/otto"
)

func bindRugFunction(f func(call otto.FunctionCall) otto.Value) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	name = strings.Replace(name, "main.", "__duder_", -1)
	Duder.DPrint("Binding ", name)
	js.Set(name, f)
	return name
}

func createRugEnvironment() error {
	// create the rug environment
	env := fmt.Sprintf(`
		// Define DuderUser class
		function DuderUser(id, username) {
			this.id = id;
			this.username = username;
			this.isOwner = %s(id);
			this.addPermission = function(channelID, permission) {
				return %s(channelID, this.id, permission);
			}
		}

		// Define DuderCommand class
		function DuderCommand() {
			this.mentions = new Array();
		}

		DuderCommand.prototype.replyToChannel = function(content) {
			%s(this.channelID, content);
		}

		DuderCommand.prototype.replyToAuthor = function(content, mention) {
			// ensure mention is boolean
			mention = (mention == true);
			%s(this.channelID, this.author.id, this.author.username, content, mention);
		}

		// Define DuderRug class
		function DuderRug(name, description) {
			%s(this, name, description);
		}

		DuderRug.prototype.addCommand = function(trigger, exec) {
			%s(this, trigger, exec);
		}
	`,
		/* RugUser */
		bindRugFunction(rugGetIsOwner),
		bindRugFunction(rugAddPermission),
		/* RugCommand */
		bindRugFunction(rugReplyToChannel),
		bindRugFunction(rugReplyToAuthor),
		/* Rug */
		bindRugFunction(rugCreate),
		bindRugFunction(rugAddCommand))

	if _, err := js.Run(env); err != nil {
		return errors.New(fmt.Sprint("error creating rug environment: ", err.Error()))
	}

	return nil
}

func rugCreate(call otto.FunctionCall) otto.Value {
	obj := call.Argument(0).Object()
	name := call.Argument(1).String()
	description := call.Argument(2).String()

	rug := Rug{}
	rug.Name = name
	rug.Description = description
	rug.Commands = map[string]RugCommand{}
	rug.Object = obj

	rugMap[fmt.Sprintf("%v", obj)] = rug

	Duder.DPrintf("Created rug '%v' %v", name, rug.Object)

	return otto.Value{}
}

func rugAddCommand(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()

	// validate the trigger
	trigger := strings.TrimSpace(call.Argument(1).String())
	if len(trigger) == 0 {
		Duder.DPrintf("Unable to add command to rug '%v', trigger is empty", rugObj)
		return otto.Value{}
	}

	// validate the execution code
	exec := strings.TrimSpace(call.Argument(2).String())
	if len(exec) == 0 {
		Duder.DPrintf("Unable to add command '%v' to rug '%v', trigger is empty", trigger, rugObj)
		return otto.Value{}
	}

	// add to parent rug
	if rug, ok := rugMap[fmt.Sprintf("%v", rugObj)]; ok {
		rugCmd := RugCommand{}
		rugCmd.Trigger = trigger
		rugCmd.Exec = fmt.Sprintf("__execCmd = %s; __execCmd()", exec)
		rug.Commands[trigger] = rugCmd
		Duder.DPrintf("Added command '%v' to rug '%v'", trigger, rug.Name)
	} else {
		Duder.DPrintf("Unable to add command to rug '%v'", rugObj)
	}

	return otto.Value{}
}

func rugAddPermission(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	userID := call.Argument(1).String()
	perm, _ := call.Argument(2).ToInteger()

	if err := Duder.Permissions.AddPermission(channelID, userID, int(perm)); err != nil {
		if result, e := js.ToValue(err.Error()); e == nil {
			return result
		}
		return otto.TrueValue()
	}

	return otto.NullValue()
}

func rugGetIsOwner(call otto.FunctionCall) otto.Value {
	clientID := call.Argument(0).String()
	if clientID == Duder.Config.OwnerID {
		return otto.TrueValue()
	}
	return otto.FalseValue()
}

func rugReplyToChannel(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	content := call.Argument(1).String()

	Duder.Session.ChannelMessageSend(channelID, content)

	return otto.Value{}
}

func rugReplyToAuthor(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	authorID := call.Argument(1).String()
	authorUsername := call.Argument(2).String()
	content := call.Argument(3).String()
	mention, _ := call.Argument(4).ToBoolean()

	if mention {
		Duder.Session.ChannelMessageSend(channelID, fmt.Sprintf("<@%s> %s", authorID, content))
	} else {
		Duder.Session.ChannelMessageSend(channelID, fmt.Sprintf("%s %s", authorUsername, content))
	}

	return otto.Value{}
}
