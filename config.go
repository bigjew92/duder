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

rug_path = "rugs"
prefix = "!d"
`

// LoadConfig loads the configuration file
func LoadConfig(path string) error {
	// validate the config file
	path = strings.TrimSpace(path)
	if len(path) == 0 {
		return errors.New("config file is undefined")
	}
	log.Print("Loading configuration file: ", path)

	// check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Print("Configuration file not found; creating new one...\r\n")

		// prompt for bot token
		botToken := getInput("Bot token", true)
		// promot for owner ID
		clientID := getInput("Owner ID", true)

		// populate the configuration data
		configData := strings.Replace(defaultConfig, "BOT_TOKEN", botToken, 1)
		configData = strings.Replace(configData, "OWNER_ID", clientID, 1)

		// create the configuration file
		if err := ioutil.WriteFile(path, []byte(configData), 0644); err != nil {
			log.Print("Unable to create configuration file ", path)
			return err
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
