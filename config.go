package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type config struct {
	BotToken string `toml:"bot_token"`
	OwnerID  string `toml:"owner_ID"`
	RugPath  string `toml:"rug_path"`
	Prefix   string `toml:"prefix"`
}

const defaultConfig = `# Duder configuration file

# Discord app bot token
bot_token = "BOT_TOKEN"

# Discord client ID for bot owner
owner_ID = "OWNER_ID"

# Path where Rugs are located
rug_path = "rugs"

# Message prefix to indicate it's a bot command
prefix = "!d"
`

// LoadConfig loads the configuration file
func LoadConfig(path string) error {
	// validate the config file
	path = strings.TrimSpace(path)
	if len(path) == 0 {
		return errors.New("configuration file is undefined")
	}
	log.Print("Loading configuration file: ", path)

	// check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Print("Configuration file not found; creating new one...\r\n")

		// prompt for bot token
		botToken := getInput("Bot token", true)
		// prompt for owner ID
		clientID := getInput("Owner ID", true)

		// populate the configuration data
		configData := strings.Replace(defaultConfig, "BOT_TOKEN", botToken, 1)
		configData = strings.Replace(configData, "OWNER_ID", clientID, 1)

		// create the configuration file
		if err := ioutil.WriteFile(path, []byte(configData), 0644); err != nil {
			return errors.New(fmt.Sprint("unable to create configuration file ", path, err.Error()))
		}
		log.Printf("Configuration file %v created", path)

		// load the configuration data
		if _, err := toml.Decode(configData, &Duder.Config); err != nil {
			return err
		}
	} else {
		// load the configuration file
		if _, err := toml.DecodeFile(path, &Duder.Config); err != nil {
			return err
		}
	}

	// validate the prefix
	Duder.Config.Prefix = strings.TrimSpace(Duder.Config.Prefix)
	if len(Duder.Config.Prefix) == 0 {
		return errors.New("'prefix' is undefined in configuration file")
	}

	// ensure the bot token has the 'Bot ' prefix
	if !strings.HasPrefix(Duder.Config.BotToken, "Bot ") {
		Duder.Config.BotToken = fmt.Sprintf("Bot %s", Duder.Config.BotToken)
	}

	return nil
}

func getInput(prompt string, required bool) string {
	var input string
	for {
		fmt.Print(fmt.Sprintf("%s: ", prompt))
		fmt.Scanln(&input)
		if len(input) > 0 || !required {
			return input
		}
	}
}
