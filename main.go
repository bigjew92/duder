package main

import (
	"flag"
	"fmt"
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

// LogChannel description
var LogChannel = struct {
	General uint8
	Warning uint8
	Verbose uint8
}{
	General: 0,
	Warning: 1,
	Verbose: 2,
}

func init() {
	Duder.Logf(LogChannel.General, "Duder version %s", VERSION)

	flag.BoolVar(&Duder.debug, "debug", true, "Enable debug mode")
	flag.StringVar(&Duder.Config.path, "config", "config.json", "Configuration file (default config.json)")
	flag.Parse()
}

func main() {
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

	Duder.Log(LogChannel.General, "Bot is now running.")

	// register bot sg.shutdown channel to receive shutdown signals.
	signal.Notify(Duder.shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	// wait for shutdown signal
	<-Duder.shutdownSignal

	Duder.Log(LogChannel.General, "termination signal received; shutting down...")

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
	if channel == LogChannel.Verbose && !duder.debug {
		return
	}

	switch channel {
	case LogChannel.Verbose:
		c := color.New(color.FgYellow)
		c.Println(v...)
	case LogChannel.Warning:
		c := color.New(color.FgHiYellow)
		c.Println(v...)
	default:
		fmt.Println(v...)
	}
}

// Logf description
func (duder *DuderBot) Logf(channel uint8, format string, v ...interface{}) {
	if channel == LogChannel.Verbose && !duder.debug {
		return
	}

	msg := fmt.Sprintf(format, v...)

	switch channel {
	case LogChannel.Verbose:
		c := color.New(color.FgYellow)
		c.Printf(format, v...)
		c.Println("")
	case LogChannel.Warning:
		c := color.New(color.FgHiYellow)
		c.Printf(format, v...)
		c.Println("")
	default:
		fmt.Println(msg)
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

	duder.Logf(LogChannel.General, "Running update command '%s'", duder.Config.UpdateExec())
	output, err := exec.Command(duder.Config.UpdateExec()).CombinedOutput()
	if err != nil {
		duder.Logf(LogChannel.Warning, "Error running update command; %s", err.Error())
	} else {
		duder.Logf(LogChannel.General, "Update command exited with '%s'", string(output))
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
