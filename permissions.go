package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type permissionsChannel struct {
	Users map[string][]int
}

type permissions struct {
	Channels map[string]permissionsChannel
	path     string
}

const (
	// PermissionNone is given to everyone
	PermissionNone = 0
	// PermissionOwner only the bot owner can use
	PermissionOwner = 1
	// PermissionModerator stuff
	PermissionModerator = 2
)

type permissionDefinition struct {
	Value int
	Names []string
}

var permissionDefinitions map[int]permissionDefinition

func init() {
	permissionDefinitions = make(map[int]permissionDefinition)
	permissionDefinitions[-1] = permissionDefinition{Value: -1, Names: []string{"Invalid"}}
	permissionDefinitions[1] = permissionDefinition{Value: 1, Names: []string{"Moderator", "Mod"}}
	permissionDefinitions[2] = permissionDefinition{Value: 2, Names: []string{"Owner"}}
}

// LoadPermissions description
func LoadPermissions(path string) error {
	// validate the config file
	path = strings.TrimSpace(path)
	if len(path) == 0 {
		return errors.New("permissions file is undefined")
	}
	log.Print("Loading permissions file: ", path)

	// check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Print("Permissions file not found; creating new one...\r\n")

		// create the configuration file
		if e := ioutil.WriteFile(path, []byte("{}"), 0644); e != nil {
			return errors.New(fmt.Sprint("unable to create permissions file ", path, e.Error()))
		}
		log.Printf("Permissions file %v created", path)
	} else {
		var bytes []byte
		if bytes, err = ioutil.ReadFile(path); err != nil {
			return errors.New(fmt.Sprint("unable to read permissions file ", path, err.Error()))
		}

		if err := json.Unmarshal(bytes, &Duder.Permissions); err != nil {
			return errors.New(fmt.Sprint("unable to load permissions file ", path, err.Error()))
		}
	}
	Duder.Permissions.path = path

	return nil
}

func (p *permissions) GetPermissionByName(name string) permissionDefinition {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return permissionDefinitions[-1]
	}

	for _, p := range permissionDefinitions {
		for _, n := range p.Names {
			if strings.EqualFold(n, name) {
				return p
			}
		}
	}

	return permissionDefinitions[-1]
}

func (p *permissions) GetPermissionByValue(value int) permissionDefinition {
	if perm, ok := permissionDefinitions[value]; ok {
		return perm
	}

	return permissionDefinitions[-1]
}

func (p *permissions) GetPermissions(channelID string, userID string) []int {
	var perms []int

	if channel, ok := p.Channels[channelID]; ok {
		if user, ok := channel.Users[userID]; ok {
			perms = user
		}
	}

	return perms
}

func (p *permissions) AddPermission(channelID string, userID string, perm int) error {
	if len(p.Channels) == 0 {
		p.Channels = make(map[string]permissionsChannel)
	}
	var channel permissionsChannel
	if c, ok := p.Channels[channelID]; !ok {
		channel = permissionsChannel{}
		channel.Users = make(map[string][]int)
		p.Channels[channelID] = channel
	} else {
		channel = c
	}

	var perms []int

	if p, ok := channel.Users[userID]; ok {
		perms = p
	}

	for _, v := range perms {
		if v == perm {
			return errors.New("user already has that permission")
		}
	}

	perms = append(perms, perm)

	channel.Users[userID] = perms
	p.Save()
	return nil
}

func (p *permissions) RemovePermission(channelID string, userID string, perm int) error {
	var channel permissionsChannel
	if c, ok := p.Channels[channelID]; ok {
		channel = c
	} else {
		return errors.New("no permissions set on channel")
	}

	var perms []int
	var newPerms []int

	if p, ok := channel.Users[userID]; ok {
		perms = p
	}

	found := false
	for _, v := range perms {
		if v == perm {
			found = true
		} else {
			newPerms = append(newPerms, v)
		}
	}

	if !found {
		return errors.New("user doesn't have that permission")
	}

	channel.Users[userID] = newPerms
	p.Save()
	return nil
}

func (p *permissions) Save() {
	if bytes, err := json.MarshalIndent(p, "", "\t"); err != nil {
		log.Print("unable to marshal permissions ", err.Error())
	} else {
		if err := ioutil.WriteFile(p.path, bytes, 0644); err != nil {
			log.Print("unable to save permissions ", err.Error())
		}
	}
}
