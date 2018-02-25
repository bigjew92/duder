package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/foszor/duder/helpers/rugutils"
	"github.com/robertkrimen/otto"
)

func bindRugFunction(f func(call otto.FunctionCall) otto.Value) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	name = strings.Replace(name, "main.", "__duder_", -1)
	Duder.dprint("Binding", name)
	Duder.jsvm.Set(name, f)
	return name
}

func createRugEnvironment() error {
	data, err := ioutil.ReadFile("rugs_env.js")
	if err != nil {
		return errors.New("unable to read rugs_env.js")
	}
	content := strings.Replace(string(data), "__BIND__", "%s", -1)
	// create the rug environment
	env := fmt.Sprintf(content,
		/* Duder */
		bindRugFunction(rugenvSetStatus),
		bindRugFunction(rugenvSetAvatar),
		bindRugFunction(rugenvSaveAvatar),
		bindRugFunction(rugenvGetAvatars),
		bindRugFunction(rugenvUseAvatar),
		bindRugFunction(rugenvStartTyping),
		/* DuderPermission */
		rugenvGetPermissionsDefinition(),
		/* DuderUser */
		bindRugFunction(rugenvRugUserGetIsOwner),
		bindRugFunction(rugenvRugUserGetIsModerator),
		bindRugFunction(rugenvRugUserGetPermissions),
		bindRugFunction(rugenvRugUserSetPermissions),
		bindRugFunction(rugenvRugUserGetUsernameByID),
		/* DuderCommand */
		bindRugFunction(rugenvRugCommandReplyToChannel),
		bindRugFunction(rugenvRugCommandReplyToChannelEmbed),
		bindRugFunction(rugenvRugCommandReplyToAuthor),
		bindRugFunction(rugenvRugCommandDeleteMessage),
		/* DuderRug */
		bindRugFunction(rugenvRugCreate),
		bindRugFunction(rugenvRugAddCommand),
		bindRugFunction(rugenvRugLoadStorage),
		bindRugFunction(rugenvRugSaveStorage),
		bindRugFunction(rugenvRugDPrint),
		bindRugFunction(rugenvRugWPrint),
		/* String */
		bindRugFunction(rugenvStringDecodeHTML),
		/* HTTP */
		bindRugFunction(rugenvHTTPGet),
		bindRugFunction(rugenvHTTPPost),
		bindRugFunction(rugenvHTTPDetectContentType),
		bindRugFunction(rugenvHTTPParseURL),
		/* Base64 */
		bindRugFunction(rugenvBase64EncodeToString))

	if _, err := Duder.jsvm.Run(env); err != nil {
		//fmt.Print(env)
		return errors.New(err.Error())
	}

	return nil
}

func stringResponse(response string) otto.Value {
	if result, e := Duder.jsvm.ToValue(response); e == nil {
		return result
	}
	return otto.FalseValue()
}

/* Duder */
func rugenvSetStatus(call otto.FunctionCall) otto.Value {
	status := call.Argument(0).String()
	if len(status) == 0 {
		status = ""
	}
	if err := Duder.session.UpdateStatus(0, status); err != nil {
		Duder.dprint("[Duder.setStatus] Unable to set status:", err.Error())
		return stringResponse(err.Error())
	}

	return otto.TrueValue()
}

func rugenvSetAvatar(call otto.FunctionCall) otto.Value {
	avatar := call.Argument(0).String()
	if _, err := Duder.session.UserUpdate("", "", "", avatar, ""); err != nil {
		Duder.dprint("[Duder.setAvatar] Unable to set avatar:", err.Error())
		return otto.FalseValue()
	}

	return otto.TrueValue()
}

func rugenvSaveAvatar(call otto.FunctionCall) otto.Value {
	filename := call.Argument(0).String()
	if len(filename) == 0 {
		return stringResponse("invalid filename")
	}
	baseURL := Duder.me.AvatarURL("256")
	urlNoSize := baseURL[0 : len(baseURL)-9]
	parts := strings.Split(urlNoSize, ".")
	ext := "." + parts[len(parts)-1]

	if !strings.HasSuffix(filename, ext) {
		filename = fmt.Sprintf("%s%s", filename, ext)
	}

	req := http.Client{Timeout: time.Duration(5 * time.Second)}
	resp, err := req.Get(baseURL)
	if err != nil {
		Duder.dprint("[Duder.saveAvatar] Failed to download current avatar:", err.Error())
		return stringResponse("failed to download current avatar")
	}
	defer resp.Body.Close()

	if _, err := os.Stat(Duder.avatarPath); os.IsNotExist(err) {
		os.Mkdir(Duder.avatarPath, 0777)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Duder.dprint("[Duder.saveAvatar] Unable to read response body:", err.Error())
		return stringResponse("unable to read downloaded file")
	}
	if err = ioutil.WriteFile(fmt.Sprintf("%s/%s", Duder.avatarPath, filename), data, 0644); err != nil {
		Duder.dprint("[Duder.saveAvatar] Unable to save downloaded file:", err.Error())
		return stringResponse("unable to save downloaded file")
	}

	return otto.TrueValue()
}

func rugenvGetAvatars(call otto.FunctionCall) otto.Value {
	if _, err := os.Stat(Duder.avatarPath); os.IsNotExist(err) {
		os.Mkdir(Duder.avatarPath, 0777)
	}

	avatars := []string{}
	files, _ := ioutil.ReadDir(Duder.avatarPath)
	for _, f := range files {
		// ignore directories and non-image files
		if f.IsDir() || (!strings.HasSuffix(f.Name(), ".png") && !strings.HasSuffix(f.Name(), ".jpg") && !strings.HasSuffix(f.Name(), ".jpeg")) {
			continue
		}
		avatars = append(avatars, f.Name())
	}

	if result, err := Duder.jsvm.ToValue(avatars); err == nil {
		return result
	}

	return otto.FalseValue()
}

func rugenvUseAvatar(call otto.FunctionCall) otto.Value {
	filename := call.Argument(0).String()
	if len(filename) == 0 {
		return stringResponse("empty filename")
	} else if _, err := os.Stat(Duder.avatarPath); os.IsNotExist(err) {
		return stringResponse("invalid avatar path")
	}

	filePath := fmt.Sprintf("%s/%s", Duder.avatarPath, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return stringResponse("invalid filename")
	}

	if bytes, err := ioutil.ReadFile(filePath); err == nil {
		base64 := base64.StdEncoding.EncodeToString(bytes)
		avatar := fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(bytes), base64)
		_, err = Duder.session.UserUpdate("", "", "", avatar, "")
		if err != nil {
			Duder.dprint("[Duder.useAvatar] Unable to update avatar:", err.Error())
			return stringResponse("unable to update avatar")
		}
		return otto.TrueValue()
	}

	return otto.FalseValue()
}

func rugenvStartTyping(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	if len(channelID) == 0 {
		return otto.FalseValue()
	}
	if err := Duder.session.ChannelTyping(channelID); err != nil {
		Duder.dprint("[Duder.startTyping] Unable to start typing in channel", channelID, ":", err.Error())
		return otto.FalseValue()
	}
	return otto.TrueValue()
}

/* DuderPermission */
func rugenvGetPermissionsDefinition() string {
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

/* DuderUser */
func rugenvRugUserGetIsOwner(call otto.FunctionCall) otto.Value {
	guildID := call.Argument(0).String()
	userID := call.Argument(1).String()
	if userID == Duder.config.OwnerID {
		return otto.TrueValue()
	} else if Duder.permissions.isOwner(guildID, userID) {
		return otto.TrueValue()
	}
	return otto.FalseValue()
}

func rugenvRugUserGetIsModerator(call otto.FunctionCall) otto.Value {
	guildID := call.Argument(0).String()
	userID := call.Argument(1).String()
	if userID == Duder.config.OwnerID {
		return otto.TrueValue()
	} else if Duder.permissions.isModerator(guildID, userID) {
		return otto.TrueValue()
	}
	return otto.FalseValue()
}

func rugenvRugUserGetPermissions(call otto.FunctionCall) otto.Value {
	guildID := call.Argument(0).String()
	userID := call.Argument(1).String()
	perms := Duder.permissions.getAll(guildID, userID)

	if result, err := Duder.jsvm.Run(rugutils.ConvertUserPermission(perms)); err == nil {
		return result
	}

	return otto.Value{}
}

func rugenvRugUserSetPermissions(call otto.FunctionCall) otto.Value {
	guildID := call.Argument(0).String()
	userID := call.Argument(1).String()
	permName := call.Argument(2).String()
	add, _ := call.Argument(3).ToBoolean()

	perm := Duder.permissions.getByName(permName)
	if perm.Value == -1 {
		if result, e := Duder.jsvm.ToValue(fmt.Sprintf("invalid permission '%s'", permName)); e == nil {
			return result
		}
		return otto.NullValue()
	}

	if add {
		if err := Duder.permissions.addToUser(guildID, userID, perm.Value); err != nil {
			if result, e := Duder.jsvm.ToValue(err.Error()); e == nil {
				return result
			}
			return otto.TrueValue()
		}
	} else {
		if err := Duder.permissions.removeFromUser(guildID, userID, perm.Value); err != nil {
			if result, e := Duder.jsvm.ToValue(err.Error()); e == nil {
				return result
			}
			return otto.TrueValue()
		}
	}

	return otto.NullValue()
}

func rugenvRugUserGetUsernameByID(call otto.FunctionCall) otto.Value {
	guildID := call.Argument(0).String()
	userID := call.Argument(1).String()

	username := "Unknown"

	if guild, err := Duder.session.Guild(guildID); err == nil {
		for _, member := range guild.Members {
			if member.User.ID == userID {
				username = member.User.Username
				break
			}
		}
	}

	if result, err := Duder.jsvm.ToValue(username); err == nil {
		return result
	}
	return otto.NullValue()
}

/* DuderCommand */
func rugenvRugCommandReplyToChannel(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	content := call.Argument(1).String()

	Duder.session.ChannelMessageSend(channelID, content)

	return otto.TrueValue()
}

func rugenvRugCommandReplyToChannelEmbed(call otto.FunctionCall) otto.Value {
	// https://godoc.org/github.com/bwmarrin/discordgo#MessageEmbed
	channelID := call.Argument(0).String()
	content := call.Argument(1).String()

	embed := new(discordgo.MessageEmbed)
	if err := json.Unmarshal([]byte(content), &embed); err != nil {
		Duder.dprint("[cmd.replyToChannelEmbed] Failed to create embed", err.Error())
	} else {
		Duder.session.ChannelMessageSendEmbed(channelID, embed)
	}

	return otto.TrueValue()
}

func rugenvRugCommandReplyToAuthor(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	authorID := call.Argument(1).String()
	authorUsername := call.Argument(2).String()
	content := call.Argument(3).String()
	mention, _ := call.Argument(4).ToBoolean()

	if mention {
		Duder.session.ChannelMessageSend(channelID, fmt.Sprintf("<@%s> %s", authorID, content))
	} else {
		Duder.session.ChannelMessageSend(channelID, fmt.Sprintf("%s, %s", authorUsername, content))
	}

	return otto.TrueValue()
}

func rugenvRugCommandDeleteMessage(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	messageID := call.Argument(1).String()

	Duder.session.ChannelMessageDelete(channelID, messageID)

	return otto.Value{}
}

/* DuderRug */
func rugenvRugCreate(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	name := call.Argument(1).String()
	description := call.Argument(2).String()

	rug := Rug{}
	rug.name = name
	rug.description = description
	rug.commands = map[string]rugCommand{}
	rug.object = rugObj

	addRug(rug)

	Duder.dprintf("Created Rug '%v'", name)

	return otto.Value{}
}

func rugenvRugAddCommand(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()

	// validate the trigger
	trigger := strings.TrimSpace(call.Argument(1).String())
	if len(trigger) == 0 {
		Duder.wprintf("[rug.addCommand] Unable to add command to Rug '%v', trigger is empty", rugObj)
		return otto.FalseValue()
	}

	// add to parent rug
	if rug, ok := rugMap[fmt.Sprintf("%v", rugObj)]; ok {
		rugCmd := rugCommand{}
		rugCmd.trigger = trigger
		rugCmd.exec = call.Argument(2)
		rug.commands[trigger] = rugCmd
		Duder.dprintf("Added command '%v' to Rug '%v'", trigger, rug.name)
	} else {
		Duder.wprintf("Unable to add command to Rug '%v'", rugObj)
	}

	return otto.TrueValue()
}

func rugenvRugLoadStorage(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()

	if rug, ok := rugMap[fmt.Sprintf("%v", rugObj)]; ok {
		path := getRugStorageFile(rug)

		// check if the file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			Duder.dprintf("[rug.loadStorage] Storage file for '%v' not found; creating new one...", rug.name)

			// create the storage file
			if e := ioutil.WriteFile(path, []byte("{}"), 0644); e != nil {
				Duder.dprint("[rug.loadStorage] Unable to create storage file", path)
				return otto.FalseValue()
			}
			Duder.dprintf("[rug.loadStorage] Storage file for '%v' created", rug.name)
			if result, e := Duder.jsvm.ToValue("{}"); e == nil {
				return result
			}
		} else {
			var bytes []byte
			if bytes, err = ioutil.ReadFile(path); err != nil {
				Duder.dprint("[rug.loadStorage] Unable to read storage file", path)
				return otto.FalseValue()
			}

			if result, e := Duder.jsvm.ToValue(string(bytes)); e == nil {
				return result
			}
		}
		return otto.FalseValue()
	}
	return otto.FalseValue()
}

func rugenvRugSaveStorage(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	data := call.Argument(1).String()
	if rug, ok := rugMap[fmt.Sprintf("%v", rugObj)]; ok {
		path := getRugStorageFile(rug)
		if err := ioutil.WriteFile(path, []byte(data), 0644); err != nil {
			Duder.dprint("[rug.saveStorage] Unable to save rug storage", err.Error())
			return otto.FalseValue()
		}
		Duder.dprint("[rug.saveStorage] Unable to save rug storage")
		return otto.TrueValue()
	}
	return otto.FalseValue()
}

func rugenvRugDPrint(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	msg := call.Argument(1).String()
	if rug, ok := rugMap[fmt.Sprintf("%v", rugObj)]; ok {
		Duder.dprintf("[Rug:%v(%v)] %v", rug.name, rug.file, msg)
		return otto.TrueValue()
	}
	return otto.FalseValue()
}

func rugenvRugWPrint(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	msg := call.Argument(1).String()
	if rug, ok := rugMap[fmt.Sprintf("%v", rugObj)]; ok {
		Duder.wprintf("[Rug:%v(%v)] %v", rug.name, rug.file, msg)
		return otto.TrueValue()
	}
	return otto.FalseValue()
}

/* String */
func rugenvStringDecodeHTML(call otto.FunctionCall) otto.Value {
	text := call.Argument(0).String()
	text = html.UnescapeString(text)

	if result, err := Duder.jsvm.ToValue(text); err == nil {
		return result
	}

	return otto.NullValue()
}

/* HTTP */
func rugenvHTTPGet(call otto.FunctionCall) otto.Value {
	var timeout int64
	timeout, _ = call.Argument(0).ToInteger()
	url := call.Argument(1).String()
	headers, _ := call.Argument(2).Export()
	stringResult, _ := call.Argument(3).ToBoolean()

	client := http.Client{Timeout: time.Duration(timeout * int64(time.Second))}
	req, err := http.NewRequest("GET", url, nil)
	for k, v := range headers.(map[string]interface{}) {
		if value, ok := v.(string); ok {
			req.Header.Add(k, value)
			Duder.dprintf("[http.Get] Adding header %s : %s", k, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return otto.FalseValue()
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return otto.FalseValue()
	}

	if stringResult {
		if result, err := Duder.jsvm.ToValue(string(body)); err == nil {
			Duder.dprint("[http.Get] Returning string")
			return result
		}
	} else {
		if result, err := Duder.jsvm.ToValue(body); err == nil {
			Duder.dprint("[http.Get] Returning byte array")
			return result
		}
	}

	return otto.FalseValue()
}

func rugenvHTTPPost(call otto.FunctionCall) otto.Value {
	var timeout int64
	timeout, _ = call.Argument(0).ToInteger()
	url := call.Argument(1).String()
	data := call.Argument(2).Object()

	// this doesn't actually work yet

	print(timeout)
	print(url)

	for _, k := range data.Keys() {
		log.Print(k)
		if v, err := data.Get(k); err == nil {
			log.Print(v)
		}
	}

	return otto.TrueValue()
}

func rugenvHTTPDetectContentType(call otto.FunctionCall) otto.Value {
	var data []byte
	export, _ := call.Argument(0).Export()
	{
		data, _ = export.([]byte)
	}

	contentType := http.DetectContentType(data)
	if result, err := Duder.jsvm.ToValue(contentType); err == nil {
		return result
	}
	return otto.FalseValue()
}

func rugenvHTTPParseURL(call otto.FunctionCall) otto.Value {
	urlString := call.Argument(0).String()
	if u, err := url.Parse(urlString); err == nil {
		if result, err := Duder.jsvm.ToValue(u.String()); err == nil {
			return result
		}
	}
	return otto.FalseValue()
}

/* Base64 */
func rugenvBase64EncodeToString(call otto.FunctionCall) otto.Value {
	var data []byte
	export, _ := call.Argument(0).Export()
	{
		data, _ = export.([]byte)
	}

	str := base64.StdEncoding.EncodeToString(data)
	if result, err := Duder.jsvm.ToValue(str); err == nil {
		return result
	}
	return otto.FalseValue()
}
