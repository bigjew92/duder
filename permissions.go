package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func init() {
	Duder.Permissions = PermissionsManager{}
}

// GuildPermissions description
type GuildPermissions struct {
	ManagerRoles   []string `json:"managerRoles"`
	ModeratorRoles []string `json:"moderatorRoles"`
}

// IsManagerRole description
func (permissions *GuildPermissions) IsManagerRole(roleID string) bool {
	for _, m := range permissions.ManagerRoles {
		if m == roleID {
			return true
		}
	}
	return false
}

// IsModeratorRole description
func (permissions *GuildPermissions) IsModeratorRole(roleID string) bool {
	for _, m := range permissions.ModeratorRoles {
		if m == roleID {
			return true
		}
	}
	return false
}

// Permissions description
type Permissions struct {
	Guilds map[string]GuildPermissions `json:"guilds"`
}

// PermissionsManager description
type PermissionsManager struct {
	data Permissions
}

// Load description
func (manager *PermissionsManager) Load() error {
	path := strings.TrimSpace(Duder.Config.PermissionsFile())
	if len(path) == 0 {
		return errors.New("permissions file isn't defined")
	}

	Duder.Logf(LogChannel.General, "Loading permissions file '%s'", path)

	// check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		Duder.Log(LogChannel.General, "Permissions file not found; creating new one...")

		// create the configuration file
		if err := ioutil.WriteFile(path, []byte("{}"), 0644); err != nil {
			return fmt.Errorf("unable to create permissions file '%s'; %s", path, err.Error())
		}

		Duder.Logf(LogChannel.General, "Permissions file '%s' created", path)
	} else {
		var bytes []byte
		if bytes, err = ioutil.ReadFile(path); err != nil {
			return fmt.Errorf("unable to read permissions file '%s'; %s", path, err.Error())
		}

		if err := json.Unmarshal(bytes, &Duder.Permissions.data); err != nil {
			return fmt.Errorf("unable to load permissions file '%s'; %s", path, err.Error())
		}
	}

	Duder.Log(LogChannel.Verbose, "Permissions file successfully loaded")

	return nil
}

// Save description
func (manager *PermissionsManager) Save() error {
	return nil
}

// IsUserManager description
func (manager *PermissionsManager) IsUserManager(guildID string, userID string) bool {
	if userID == Duder.Config.OwnerID() {
		return true
	}

	permissions, ok := manager.data.Guilds[guildID]
	if !ok {
		return false
	}

	guild, ok := Duder.Discord.GetGuildByID(guildID)
	if !ok {
		return false
	}
	for _, member := range guild.Members {
		if member.User.ID == userID {
			for _, role := range member.Roles {
				if permissions.IsManagerRole(role) {
					return true
				}
			}
		}
	}
	return false
}

// IsUserModerator description
func (manager *PermissionsManager) IsUserModerator(guildID string, userID string) bool {
	if userID == Duder.Config.OwnerID() {
		return true
	}

	permissions, ok := manager.data.Guilds[guildID]
	if !ok {
		return false
	}

	guild, ok := Duder.Discord.GetGuildByID(guildID)
	if !ok {
		return false
	}
	for _, member := range guild.Members {
		if member.User.ID == userID {
			for _, role := range member.Roles {
				if permissions.IsModeratorRole(role) {
					return true
				}
			}
		}
	}
	return false
}

// AddManagerRole description
func (manager *PermissionsManager) AddManagerRole(guildID string, roleName string) error {
	permissions, ok := manager.data.Guilds[guildID]
	if !ok {
		manager.data.Guilds[guildID] = GuildPermissions{}
		permissions, _ = manager.data.Guilds[guildID]
	}
	guild, ok := Duder.Discord.GetGuildByID(guildID)
	if !ok {
		return errors.New("couldn't find guild")
	}

	roleName = strings.ToLower(roleName)
	roleID := ""

	for _, role := range guild.Roles {
		if strings.ToLower(role.Name) == roleName {
			roleID = role.ID
			break
		}
	}

	if len(roleID) == 0 {
		return fmt.Errorf("couldn't find guild role '%s'", roleName)
	}

	if permissions.IsManagerRole(roleID) {
		return fmt.Errorf("'%s' is already a manager role", roleName)
	}

	permissions.ManagerRoles = append(permissions.ManagerRoles, roleName)

	return nil
}

// teardown description
func (manager *PermissionsManager) teardown() {
}
