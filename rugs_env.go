package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	xml2json "github.com/basgys/goxml2json"
	"github.com/robertkrimen/otto"
)

func bindRugFunction(f func(call otto.FunctionCall) otto.Value) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	name = strings.Replace(name, "main.", "__duder_", -1)
	Duder.Log(LogVerbose, "[Rugs.BindRugFunction] Binding", name)
	Duder.Rugs.VM.Set(name, f)
	return name
}

func createRugEnvironment() error {
	data, err := os.ReadFile("rugs_env.js")
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
		/* DuderUser */
		bindRugFunction(rugenvRugUserGetIsOwner),
		bindRugFunction(rugenvRugUserGetIsManager),
		bindRugFunction(rugenvRugUserGetIsModerator),
		bindRugFunction(rugenvRugUserSetNickname),
		bindRugFunction(rugenvRugUserGetUsernameByID),
		bindRugFunction(rugenvRugUserGetIDByNickname),
		/* DuderCommand */
		bindRugFunction(rugenvRugCommandReplyToChannel),
		bindRugFunction(rugenvRugCommandReplyToChannelEmbed),
		bindRugFunction(rugenvRugCommandReplyToAuthor),
		bindRugFunction(rugenvRugCommandDeleteMessage),
		bindRugFunction(rugenvRugCommandSendFile),
		/* DuderMessageReaction */
		bindRugFunction(rugenvRugCommandReplyToChannel),
		bindRugFunction(rugenvRugCommandReplyToChannelEmbed),
		/* DuderRug */
		bindRugFunction(rugenvRugCreate),
		bindRugFunction(rugenvRugAddCommand),
		bindRugFunction(rugenvRugBindOnMessage),
		bindRugFunction(rugenvRugBindOnMessageReactionAdd),
		bindRugFunction(rugenvRugBindOnMessageReactionRemove),
		bindRugFunction(rugenvRugBindOnPresenceUpdate),
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
		bindRugFunction(rugenvBase64EncodeToString),
		/* XML */
		bindRugFunction(rugenvXMLtoJSON))

	if _, err := Duder.Rugs.VM.Run(env); err != nil {
		//fmt.Print(env)
		return errors.New(err.Error())
	}

	Duder.Rugs.VM.Set("print", func(msg string) { log.Println("[JS]", msg) })

	return nil
}

// response description
func response(value interface{}, defaultValue otto.Value) otto.Value {
	if v, err := Duder.Rugs.VM.ToValue(value); err == nil {
		return v
	}
	return defaultValue
}

/* Duder */
func rugenvSetStatus(call otto.FunctionCall) otto.Value {
	status := call.Argument(0).String()
	if len(status) == 0 {
		status = ""
	}
	Duder.Discord.SetStatus(status)
	Duder.Config.SetStatus(status)

	return otto.TrueValue()
}

func rugenvSetAvatar(call otto.FunctionCall) otto.Value {
	avatar := call.Argument(0).String()

	if ok := Duder.Discord.SetAvatarByImage(avatar); ok {
		return otto.TrueValue()
	}

	return otto.FalseValue()
}

func rugenvSaveAvatar(call otto.FunctionCall) otto.Value {
	filename := call.Argument(0).String()

	if err := Duder.Discord.SaveAvatar(filename); err != nil {
		return response(err.Error(), otto.FalseValue())
	}

	return otto.TrueValue()
}

func rugenvGetAvatars(call otto.FunctionCall) otto.Value {
	return response(Duder.Discord.Avatars(), otto.FalseValue())
}

func rugenvUseAvatar(call otto.FunctionCall) otto.Value {
	filename := call.Argument(0).String()

	if err := Duder.Discord.SetAvatarByFile(filename); err != nil {
		return response(err.Error(), otto.FalseValue())
	}

	return otto.TrueValue()
}

func rugenvStartTyping(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	if len(channelID) == 0 {
		return otto.FalseValue()
	}

	ok := Duder.Discord.StartTyping(channelID)
	return response(ok, otto.FalseValue())
}

/* DuderUser */
func rugenvRugUserGetIsOwner(call otto.FunctionCall) otto.Value {
	userID := call.Argument(0).String()

	return response((userID == Duder.Config.OwnerID()), otto.FalseValue())
}

func rugenvRugUserGetIsManager(call otto.FunctionCall) otto.Value {
	guildID := call.Argument(0).String()
	userID := call.Argument(1).String()

	return response(Duder.Permissions.IsUserManager(guildID, userID), otto.FalseValue())
}

func rugenvRugUserGetIsModerator(call otto.FunctionCall) otto.Value {
	guildID := call.Argument(0).String()
	userID := call.Argument(1).String()

	return response(Duder.Permissions.IsUserModerator(guildID, userID), otto.FalseValue())
}

func rugenvRugUserSetNickname(call otto.FunctionCall) otto.Value {
	guildID := call.Argument(0).String()
	userID := call.Argument(1).String()
	nickname := call.Argument(1).String()

	Duder.Discord.SetMemberNickname(guildID, userID, nickname)

	return otto.TrueValue()
}

func rugenvRugUserGetUsernameByID(call otto.FunctionCall) otto.Value {
	guildID := call.Argument(0).String()
	userID := call.Argument(1).String()

	member, ok := Duder.Discord.GetGuildMember(guildID, userID)
	if !ok {
		return otto.FalseValue()
	}

	return response(member.User.Username, otto.FalseValue())
}

func rugenvRugUserGetIDByNickname(call otto.FunctionCall) otto.Value {
	guildID := call.Argument(0).String()
	nickname := call.Argument(1).String()

	member, ok := Duder.Discord.GetMemberByNickname(guildID, nickname)
	if !ok {
		return otto.FalseValue()
	}

	return response(member.User.ID, otto.FalseValue())
}

/* DuderCommand */
func rugenvRugCommandReplyToChannel(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	content := call.Argument(1).String()

	//Duder.session.ChannelMessageSend(channelID, content)
	Duder.Discord.SendMessageToChannel(channelID, content)

	return otto.TrueValue()
}

func rugenvRugCommandReplyToChannelEmbed(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	jsonData := call.Argument(1).String()

	if err := Duder.Discord.SendEmbedToChannel(channelID, jsonData); err != nil {
		Duder.Log(LogVerbose, "[cmd.replyToChannelEmbed] Failed to create embed", err.Error())
		return otto.FalseValue()
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
		Duder.Discord.SendMessageToChannel(channelID, fmt.Sprintf("<@%s> %s", authorID, content))
	} else {
		Duder.Discord.SendMessageToChannel(channelID, fmt.Sprintf("%s, %s", authorUsername, content))
	}

	return otto.TrueValue()
}

func rugenvRugCommandDeleteMessage(call otto.FunctionCall) otto.Value {
	channelID := call.Argument(0).String()
	messageID := call.Argument(1).String()

	Duder.Discord.DeleteChannelMessage(channelID, messageID)

	return otto.TrueValue()
}

func rugenvRugCommandSendFile(call otto.FunctionCall) otto.Value {
	//channelID := call.Argument(0).String()
	//name := call.Argument(1).String()
	//data := call.Argument(2).String()

	return otto.TrueValue()
}

/* DuderRug */
func rugenvRugCreate(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	name := call.Argument(1).String()
	description := call.Argument(2).String()

	Duder.Rugs.CreateRug(rugObj, name, description)

	return otto.TrueValue()
}

func rugenvRugAddCommand(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	trigger := strings.TrimSpace(call.Argument(1).String())
	exec := call.Argument(2)

	// validate the trigger
	if len(trigger) == 0 {
		return otto.FalseValue()
	}

	if rug, ok := Duder.Rugs.FindRugByObject(rugObj); ok {
		rug.AddCommand(trigger, exec)
	}

	return otto.TrueValue()
}

func rugenvRugBindOnMessage(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	onMessage := call.Argument(1)

	if rug, ok := Duder.Rugs.FindRugByObject(rugObj); ok {
		rug.BindOnMessage(onMessage)
	}

	return otto.TrueValue()
}

func rugenvRugBindOnPresenceUpdate(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	onPresenceUpdate := call.Argument(1)

	if rug, ok := Duder.Rugs.FindRugByObject(rugObj); ok {
		rug.BindOnPresenceUpdate(onPresenceUpdate)
	}

	return otto.TrueValue()
}

func rugenvRugBindOnMessageReactionAdd(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	onMessageReactionAdd := call.Argument(1)

	if rug, ok := Duder.Rugs.FindRugByObject(rugObj); ok {
		rug.BindOnMessageReactionAdd(onMessageReactionAdd)
	}

	return otto.TrueValue()
}

func rugenvRugBindOnMessageReactionRemove(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	onMessageReactionRemove := call.Argument(1)

	if rug, ok := Duder.Rugs.FindRugByObject(rugObj); ok {
		rug.BindOnMessageReactionRemove(onMessageReactionRemove)
	}

	return otto.TrueValue()
}

func rugenvRugLoadStorage(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()

	if rug, ok := Duder.Rugs.FindRugByObject(rugObj); ok {
		storage, ok := rug.LoadStorage()
		if ok {
			return response(storage, otto.FalseValue())
		}

	}

	return otto.FalseValue()
}

func rugenvRugSaveStorage(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	data := call.Argument(1).String()

	if rug, ok := Duder.Rugs.FindRugByObject(rugObj); ok {
		return response(rug.SaveStorage(data), otto.FalseValue())
	}

	return otto.FalseValue()
}

func rugenvRugDPrint(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	msg := call.Argument(1).String()
	if rug, ok := Duder.Rugs.FindRugByObject(rugObj); ok {
		rug.DPrint(msg)
		return otto.TrueValue()
	}
	return otto.FalseValue()
}

func rugenvRugWPrint(call otto.FunctionCall) otto.Value {
	rugObj := call.Argument(0).Object()
	msg := call.Argument(1).String()
	if rug, ok := Duder.Rugs.FindRugByObject(rugObj); ok {
		rug.WPrint(msg)
		return otto.TrueValue()
	}
	return otto.FalseValue()
}

/* String */
func rugenvStringDecodeHTML(call otto.FunctionCall) otto.Value {
	text := call.Argument(0).String()
	text = html.UnescapeString(text)

	return response(text, otto.FalseValue())
}

/* HTTP */
func rugenvHTTPGet(call otto.FunctionCall) otto.Value {
	var timeout int64
	timeout, _ = call.Argument(0).ToInteger()
	uri := call.Argument(1).String()
	headers, _ := call.Argument(2).Export()
	stringResult, _ := call.Argument(3).ToBoolean()

	client := http.Client{Timeout: time.Duration(timeout * int64(time.Second))}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return otto.FalseValue()
	}
	for k, v := range headers.(map[string]interface{}) {
		if value, ok := v.(string); ok {
			req.Header.Add(k, value)
			Duder.Logf(LogVerbose, "[HTTP.Get] Adding header '%s' : '%s'", k, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		Duder.Logf(LogVerbose, "[HTTP.Get] Error retrieving response; %s", err.Error())
		return otto.FalseValue()
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Duder.Logf(LogVerbose, "[HTTP.Get] Error reading response body; %s", err.Error())
		return otto.FalseValue()
	}

	if stringResult {
		Duder.Log(LogVerbose, "[HTTP.Get] Returning string")
		return response(string(body), otto.FalseValue())
	}

	Duder.Log(LogVerbose, "[HTTP.Get] Returning byte array")
	return response(body, otto.FalseValue())
}

func rugenvHTTPPost(call otto.FunctionCall) otto.Value {
	var timeout int64
	timeout, _ = call.Argument(0).ToInteger()
	uri := call.Argument(1).String()
	values, _ := call.Argument(2).Export()

	client := http.Client{Timeout: time.Duration(timeout * int64(time.Second))}
	formValues := url.Values{}
	for k, v := range values.(map[string]interface{}) {
		if value, ok := v.(string); ok {
			formValues[k] = []string{value}
			Duder.Logf(LogVerbose, "[HTTP.Post] Adding form value '%s' : '%s'", k, value)
		}
	}

	resp, err := client.PostForm(uri, formValues)
	if err != nil {
		Duder.Logf(LogVerbose, "[HTTP.Post] Error posting form; %s", err.Error())
		return otto.FalseValue()
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Duder.Logf(LogVerbose, "[HTTP.Post] Error reading response body; %s", err.Error())
		return otto.FalseValue()
	}

	return response(string(body), otto.FalseValue())
}

func rugenvHTTPDetectContentType(call otto.FunctionCall) otto.Value {
	var data []byte
	export, _ := call.Argument(0).Export()
	{
		data, _ = export.([]byte)
	}

	contentType := http.DetectContentType(data)
	if result, err := Duder.Rugs.VM.ToValue(contentType); err == nil {
		return result
	}
	return otto.FalseValue()
}

func rugenvHTTPParseURL(call otto.FunctionCall) otto.Value {
	urlString := call.Argument(0).String()
	if u, err := url.Parse(urlString); err == nil {
		if result, err := Duder.Rugs.VM.ToValue(u.String()); err == nil {
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
	if result, err := Duder.Rugs.VM.ToValue(str); err == nil {
		return result
	}
	return otto.FalseValue()
}

/* XML */
func rugenvXMLtoJSON(call otto.FunctionCall) otto.Value {
	xml := call.Argument(0).String()
	str := strings.NewReader(xml)

	json, err := xml2json.Convert(str)
	if err != nil {
		return otto.FalseValue()
	}

	return response(json.String(), otto.FalseValue())
}
