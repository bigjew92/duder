package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	configPath  string
	botID       string
	botUsername string
	botToken    string
	debugMode   bool
)

func init() {
	flag.StringVar(&configPath, "config", "duder.toml", "location of the config file, if not found it will be generated (default duder.toml)")
	flag.BoolVar(&debugMode, "debug", false, "enable debug mode")
	flag.Parse()
}

func main() {
	if err := LoadConfig(configPath); err != nil {
		log.Fatal("error loading config,", err)
	}

	botToken := Config.BotToken
	if !strings.HasPrefix(botToken, "Bot ") {
		botToken = fmt.Sprintf("Bot %s", botToken)
	}

	log.Println("creating discord session with token", botToken)

	dg, err := discordgo.New(botToken)
	if err != nil {
		log.Fatal("error creating discord session, ", err)
	}

	log.Println("obtaining discord account details")
	u, err := dg.User("@me")
	if err != nil {
		log.Fatal("error obtaining account details, ", err)
	}

	botID = u.ID
	botUsername = u.Username

	dg.AddHandler(messageCreate)

	log.Println("opening discord connection")
	err = dg.Open()
	if err != nil {
		log.Fatal("error opening discord connection,", err)
	}

	log.Println("Bot is now running.")

	<-make(chan struct{})
	return
}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Content == "!d quit" {
		if err := session.Close(); err != nil {
			log.Fatal("error closing session,", err)
		}
		os.Exit(0)
	}
}
