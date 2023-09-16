package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func init() {
	Duder.Config = new(ConfigManager)
}

// Config description
type Config struct {
	AvatarsPath     string `json:"avatarsPath"`
	BotToken        string `json:"botToken"`
	CommandPrefix   string `json:"commandPrefix"`
	OwnerID         string `json:"ownerID"`
	PermissionsFile string `json:"permissionsFile"`
	RugsPath        string `json:"rugsPath"`
	Status          string `json:"status"`
	UpdateExec      string `json:"updateExec"`
}

// ConfigManager description
type ConfigManager struct {
	path string
	data Config
}

// Load loads the configuration file
func (manager *ConfigManager) Load() error {
	// validate the config file
	manager.path = strings.TrimSpace(manager.path)
	if len(manager.path) == 0 {
		return errors.New("configuration file isn't defined")
	}
	Duder.Logf(LogGeneral, "Loading configuration file '%s'", manager.path)

	var config Config

	// check if the file exists
	if _, err := os.Stat(manager.path); os.IsNotExist(err) {
		Duder.Log(LogGeneral, "Configuration file not found; creating new one...")

		// defaults
		config = Config{
			AvatarsPath:     "avatars",
			CommandPrefix:   "!d",
			PermissionsFile: "permissions.json",
			RugsPath:        "rugs",
			Status:          "with Maude",
			UpdateExec:      "",
		}

		// required
		if os.Getenv("BOT_TOKEN") != "" {
			value, ok := os.LookupEnv("BOT_TOKEN")
			if !ok {
				Duder.Logf(LogVerbose, "BOT_TOKEN not set")
			} else {
				Duder.Logf(LogVerbose, "BOT_TOKEN set: %s", value)
				config.BotToken = value
			}
		} else {
			config.BotToken = Duder.GetUserInput("Bot token", true)
		}

		if os.Getenv("OWNER_ID") != "" {
			value, ok := os.LookupEnv("OWNER_ID")
			if !ok {
				Duder.Log(LogVerbose, "OWNER_ID not set")
			} else {
				Duder.Logf(LogVerbose, "OWNER_ID set: %s", value)
				config.OwnerID = value
			}
		} else {
			config.OwnerID = Duder.GetUserInput("Owner ID", true)
		}

		manager.data = config

		Duder.Logf(LogGeneral, "Created configuration file '%s'", manager.path)

		// save the default configuration
		if err := manager.Save(); err != nil {
			return err
		}
	} else {
		bytes, err := os.ReadFile(manager.path)
		if err != nil {
			return fmt.Errorf("Unable to read configuration file '%s'", manager.path)
		}

		if err := json.Unmarshal(bytes, &config); err != nil {
			return fmt.Errorf("Unable to load configuration file '%s'", manager.path)
		}

		manager.data = config
	}

	Duder.Log(LogVerbose, "Configuration file successfully loaded")

	return nil
}

// Save description
func (manager *ConfigManager) Save() error {
	if len(manager.path) == 0 {
		return errors.New("configuration file isn't defined")
	}
	Duder.Logf(LogGeneral, "Saving configuration file '%s'", manager.path)

	bytes, err := json.MarshalIndent(manager.data, "", "\t")
	if err != nil {
		return fmt.Errorf("unable to create configuration data; %s", err.Error())
	}
	if err := os.WriteFile(manager.path, bytes, 0777); err != nil {
		return fmt.Errorf("unable to save configuration file '%s'; %s", manager.path, err.Error())
	}

	Duder.Logf(LogVerbose, "Configuration file '%s' successfully saved", manager.path)

	return nil
}

// AvatarPath description
func (manager *ConfigManager) AvatarPath() string {
	return manager.data.AvatarsPath
}

// BotToken description
func (manager *ConfigManager) BotToken() string {
	if !strings.HasPrefix(manager.data.BotToken, "Bot ") {
		return fmt.Sprintf("Bot %s", manager.data.BotToken)
	}
	return manager.data.BotToken
}

// CommandPrefix description
func (manager *ConfigManager) CommandPrefix() string {
	return manager.data.CommandPrefix
}

// OwnerID description
func (manager *ConfigManager) OwnerID() string {
	return manager.data.OwnerID
}

// PermissionsFile description
func (manager *ConfigManager) PermissionsFile() string {
	return manager.data.PermissionsFile
}

// RugsPath description
func (manager *ConfigManager) RugsPath() string {
	return manager.data.RugsPath
}

// Status description
func (manager *ConfigManager) Status() string {
	return manager.data.Status
}

// SetStatus description
func (manager *ConfigManager) SetStatus(status string) {
	manager.data.Status = status
	manager.Save()
}

// UpdateExec description
func (manager *ConfigManager) UpdateExec() string {
	return manager.data.UpdateExec
}

// SetUpdateExec description
func (manager *ConfigManager) SetUpdateExec(exec string) {
	manager.data.UpdateExec = exec
	manager.Save()
}

// teardown description
func (manager *ConfigManager) teardown() {
}
