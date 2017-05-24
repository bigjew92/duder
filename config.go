package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type config struct {
	BotToken string `toml:"bot_token"`
}

// Config comment
var Config config

const defaultConfig = `# default config
bot_token = "REPLACE_WITH_TOKEN"
`

// LoadConfig loads the configuration file
func LoadConfig(path string) error {
	log.Print("loading configuration file ", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := ioutil.WriteFile(path, []byte(defaultConfig), 0644); err != nil {
			log.Print("unable to create default configuration file ", path)
			return err
		}
		log.Printf("configuration file %v created, edit the file and restart", path)
		os.Exit(0)
	} else {
		if _, err := toml.DecodeFile(path, &Config); err != nil {
			return err
		}
	}
	log.Print("loaded config")
	return nil
}
