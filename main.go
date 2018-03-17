package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
)

// VERSION contains the current version
const VERSION string = "1.0-a1"

// Log channels
type LogChannel uint8

const (
	// LogGeneral description
	LogGeneral = 0
	// LogWarning description
	LogWarning = 1
	// LogVerbose description
	LogVerbose = 2
)

func init() {
	Duder.Logf(LogGeneral, "Duder version %s", VERSION)

	flag.BoolVar(&Duder.debug, "debug", true, "Enable debug mode")
	flag.StringVar(&Duder.Config.path, "config", "config.json", "Configuration file (default config.json)")
	flag.Parse()
}

func main() {
	logFile := "duder.log"
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		if _, err := os.Create(logFile); err != nil {
			log.Fatal("Failed to create log file; ", err)
		}
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file; ", err)
	}

	logWriter := io.MultiWriter(os.Stdout, file)

	log.SetOutput(logWriter)
	log.SetFlags(log.Ldate | log.Ltime)

	// intitialize shutdown channel.
	Duder.shutdownSignal = make(chan os.Signal, 1)

	// load the configuration file
	if err := Duder.Config.Load(); err != nil {
		log.Fatal("Failed to load configuration file; ", err)
	}

	// load the permissions file
	if err := Duder.Permissions.Load(); err != nil {
		log.Fatal("Failed to load permissions; ", err)
	}

	// load the rugs
	if err := Duder.Rugs.Load(); err != nil {
		log.Fatal("Failed to load rugs; ", err)
	}

	// connect to Discord
	if err := Duder.Discord.Connect(); err != nil {
		log.Fatal("Failed to connect to Discord; ", err)
	}

	Duder.Log(LogGeneral, "Bot is now running.")

	// register bot sg.shutdown channel to receive shutdown signals.
	signal.Notify(Duder.shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	// wait for shutdown signal
	<-Duder.shutdownSignal

	Duder.Log(LogGeneral, "termination signal received; shutting down...")

	// gracefully shut down the bot
	Duder.teardown()

	return
}

// DuderBot struct describes the DuderBot
type DuderBot struct {
	Config         ConfigManager
	Discord        DiscordManager
	Permissions    PermissionsManager
	Rugs           RugManager
	debug          bool
	shutdownSignal chan os.Signal
}

// Duder contains the bot instance
var Duder = &DuderBot{}

// Log description
func (duder *DuderBot) Log(channel uint8, v ...interface{}) {
	if channel == LogVerbose && !duder.debug {
		return
	}

	switch channel {
	case LogVerbose:
		color.Set(color.FgYellow)
		log.Println(v...)
		color.Unset()
	case LogWarning:
		color.Set(color.FgHiYellow)
		log.Println(v...)
		color.Unset()
	default:
		log.Print(v...)
	}
}

// Logf description
func (duder *DuderBot) Logf(channel uint8, format string, v ...interface{}) {
	if channel == LogVerbose && !duder.debug {
		return
	}

	msg := fmt.Sprintf(format, v...)

	switch channel {
	case LogVerbose:
		color.Set(color.FgYellow)
		log.Println(msg)
		color.Unset()
	case LogWarning:
		color.Set(color.FgHiYellow)
		log.Println(msg)
		color.Unset()
	default:
		log.Println(msg)
	}
}

// GetUserInput description
func (duder *DuderBot) GetUserInput(prompt string, required bool) string {
	var input string
	for {
		fmt.Print(fmt.Sprintf("%s: ", prompt))
		fmt.Scanln(&input)
		if len(input) > 0 || !required {
			return input
		}
	}
}

// Update description
func (duder *DuderBot) Update(message *discordgo.MessageCreate) {
	if duder.Config.OwnerID() == message.Author.ID {
		if len(duder.Config.UpdateExec()) == 0 {
			duder.Discord.SendMessageToChannel(message.ChannelID, fmt.Sprintf("%s, the update script isn't defined", message.Author.Username))
		}
	} else {
		duder.Discord.SendMessageToChannel(message.ChannelID, fmt.Sprintf("%s, you don't have permissions for that.", message.Author.Username))
	}

	duder.Logf(LogGeneral, "Running update command '%s'", duder.Config.UpdateExec())
	output, err := exec.Command(duder.Config.UpdateExec()).CombinedOutput()
	if err != nil {
		duder.Logf(LogWarning, "Error running update command; %s", err.Error())
	} else {
		duder.Logf(LogGeneral, "Update command exited with '%s'", string(output))
	}
}

// Shutdown sends Shutdown signal to the bot's Shutdown channel.
func (duder *DuderBot) Shutdown(message *discordgo.MessageCreate) {
	if duder.Config.OwnerID() != message.Author.ID {
		duder.Discord.SendMessageToChannel(message.ChannelID, fmt.Sprintf("%s, you don't have permissions for that.", message.Author.Username))
		return
	}

	duder.Discord.SendMessageToChannel(message.ChannelID, "Goodbye.")

	duder.teardown()

	duder.shutdownSignal <- os.Interrupt
}

// teardown gracefully releases all resources and saves data before Shutdown.
func (duder *DuderBot) teardown() (err error) {
	duder.Discord.teardown()
	duder.Permissions.teardown()
	duder.Rugs.teardown()

	return
}
