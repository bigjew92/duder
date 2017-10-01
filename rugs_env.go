package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"
	"encoding/json"

	"github.com/foszor/duder/helpers/rugutils"
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
		// Define DuderPermission class
		function DuderPermission() {};
		DuderPermission.permissions = %s;
		DuderPermission.getName = function(val) {
			val = val.toString();
			for(var name in DuderPermission.permissions) {
				if (DuderPermission.permissions[name] == val) {
					return name;
				}
			}

			return "invalid";
		}
		DuderPermission.getNames = function(vals) {
			names = "";
			for(v in vals) {
				if (names.length > 0) {
					names += ", ";
				}
				names += DuderPermission.getName(vals[v]);
			}

			return names.length > 0 ? names : "none";
		}

		// Define DuderUser class
		function DuderUser(id, username) {
			this.id = id;
			this.username = username;
			this.isOwner = %s(id);
		}
		DuderUser.prototype.getPermissions = function(channelID) {
			return %s(channelID, this.id);
		}
		DuderUser.prototype.modifyPermission = function(channelID, permission, add) {
			// ensure add is boolean
			add = (add == true);
			return %s(channelID, this.id, permission, add);
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
		DuderCommand.prototype.isMention = function(str) {
			return ((str.substring(0,2) == "<@") && (str.substring(str.length-1) == ">"));
		}

		// Define DuderRug class
		function DuderRug(name, description) {
			%s(this, name, description);
		}

		DuderRug.prototype.addCommand = function(trigger, exec) {
			%s(this, trigger, exec);
		}

		// Math
		Math.getRandomInRange = function(min,max) {
			min = Math.ceil(min);
			max = Math.floor(max);
			return Math.floor(Math.random() * (max - min + 1)) + min;
		}
		Math.clamp = function(val, min, max) {
			return Math.max(min, Math.min(val, max));
		}

		// HTTP
		function HTTP() {};
		HTTP.get = function(timeout,url) {
			return %s(timeout,url);
		}
		HTTP.post = function(timeout,url,data) {
			return %s(timeout,url,data);
		}

		// JSON
		function JSON() {};
		JSON.decode = function(json) {
			return %s(json);
		}
	`,
		/* DuderPermission */
		getPermissionsDefinition(),
		/* DuderUser */
		bindRugFunction(rugUserGetIsOwner),
		bindRugFunction(rugUserGetPermissions),
		bindRugFunction(rugUserModifyPermission),
		/* DuderCommand */
		bindRugFunction(rugCommandReplyToChannel),
		bindRugFunction(rugCommandReplyToAuthor),
		/* DuderRug */
		bindRugFunction(rugCreate),
		bindRugFunction(rugAddCommand),
		/* HTTP */
		bindRugFunction(httpGet),
		bindRugFunction(httpPost),
		/* JSON */
		bindRugFunction(jsonDecode))

	if _, err := js.Run(env); err != nil {
		fmt.Print(env)
		return errors.New(fmt.Sprint("error creating rug environment: ", err.Error()))
	}

	js.Set("print", func(msg string) { log.Print(msg, "\n") })

	return nil
}

func getHTTPClient(timeout int64) http.Client {
	//print(fmt.Sprintf("timeout is %d", timeout))
	Duder.DPrintf("timeout now is '%v'", timeout)
	to := time.Duration(5 * time.Second)
	return http.Client{
		Timeout: to}
}

func httpGet(call otto.FunctionCall) otto.Value {
	var timeout int64
	timeout, _ = call.Argument(0).ToInteger()
	url := call.Argument(1).String()

	h := getHTTPClient(timeout)
	resp, err := h.Get(url)
	if err != nil {
		return otto.FalseValue()
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return otto.FalseValue()
	}

	if result, err := js.ToValue(string(body)); err == nil {
		return result
	}

	return otto.FalseValue()
}

func httpPost(call otto.FunctionCall) otto.Value {
	var timeout int64
	timeout, _ = call.Argument(0).ToInteger()
	url := call.Argument(1).String()
	data := call.Argument(2).Object()
	print(timeout);
	print(url);

	for _, k := range data.Keys() {
		log.Print(k)
		if v, err := data.Get(k); err == nil {
			log.Print(v)
		}
	}

	return otto.TrueValue()
}

func jsonDecode(call otto.FunctionCall) otto.Value {
	text := call.Argument(0).String()

	in := []byte(text)
	var raw map[string]interface{}
	if err := json.Unmarshal(in, &raw); err != nil {
		return otto.FalseValue()
	}

	if result, err := js.ToValue(raw); err == nil {
		return result
	}

	return otto.FalseValue();
}

func getPermissionsDefinition() string {
	var buffer bytes.Buffer

	buffer.WriteString("{")
	first := true
	for _, p := range permissionDefinitions {
		if !first {
			buffer.WriteString(", ")
		}
		buffer.WriteString(fmt.Sprintf("'%s': '%v'", p.Names[0], p.Value))
		first = false
	}
	buffer.WriteString("}")

	return buffer.String()
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

	Duder.DPrintf("Created Rug '%v'", name)

	return otto.Value{}
}

func rugAddCommand(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()

	// validate the trigger
	trigger := strings.TrimSpace(call.Argument(1).String())
	if len(trigger) == 0 {
		Duder.DPrintf("Unable to add command to Rug '%v', trigger is empty", rugObj)
		return otto.Value{}
	}

	// validate the execution code
	exec := strings.TrimSpace(call.Argument(2).String())
	if len(exec) == 0 {
		Duder.DPrintf("Unable to add command '%v' to Rug '%v', trigger is empty", trigger, rugObj)
		return otto.Value{}
	}

	// add to parent rug
	if rug, ok := rugMap[fmt.Sprintf("%v", rugObj)]; ok {
		rugCmd := RugCommand{}
		rugCmd.Trigger = trigger
		rugCmd.Exec = fmt.Sprintf("__execCmd = %s; __execCmd()", exec)
		rug.Commands[trigger] = rugCmd
		Duder.DPrintf("Added command '%v' to Rug '%v'", trigger, rug.Name)
	} else {
		Duder.DPrintf("Unable to add command to Rug '%v'", rugObj)
	}

	return otto.Value{}
}

func rugUserModifyPermission(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	userID := call.Argument(1).String()
	permName := call.Argument(2).String()
	add, _ := call.Argument(3).ToBoolean()

	perm := Duder.Permissions.GetPermissionByName(permName)
	if perm.Value == -1 {
		if result, e := js.ToValue(fmt.Sprintf("invalid permission '%s'", permName)); e == nil {
			return result
		}
		return otto.NullValue()
	}

	if add {
		if err := Duder.Permissions.AddPermission(channelID, userID, perm.Value); err != nil {
			if result, e := js.ToValue(err.Error()); e == nil {
				return result
			}
			return otto.TrueValue()
		}
	} else {
		if err := Duder.Permissions.RemovePermission(channelID, userID, perm.Value); err != nil {
			if result, e := js.ToValue(err.Error()); e == nil {
				return result
			}
			return otto.TrueValue()
		}
	}

	return otto.NullValue()
}

func rugUserGetIsOwner(call otto.FunctionCall) otto.Value {
	userID := call.Argument(0).String()
	if userID == Duder.Config.OwnerID {
		return otto.TrueValue()
	}
	return otto.FalseValue()
}

func rugUserGetPermissions(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	userID := call.Argument(1).String()
	perms := Duder.Permissions.GetPermissions(channelID, userID)

	if result, err := js.Run(rugutils.ConvertUserPermission(perms)); err == nil {
		return result
	}

	return otto.Value{}
}

func rugCommandReplyToChannel(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	content := call.Argument(1).String()

	Duder.Session.ChannelMessageSend(channelID, content)

	return otto.Value{}
}

func rugCommandReplyToAuthor(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	authorID := call.Argument(1).String()
	authorUsername := call.Argument(2).String()
	content := call.Argument(3).String()
	mention, _ := call.Argument(4).ToBoolean()

	if mention {
		Duder.Session.ChannelMessageSend(channelID, fmt.Sprintf("<@%s> %s", authorID, content))
	} else {
		Duder.Session.ChannelMessageSend(channelID, fmt.Sprintf("%s, %s", authorUsername, content))
	}

	return otto.Value{}
}
