package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
)

// VERSION contains the current version
const VERSION string = "0.0.1"

// Instance struct describes the bot
type Instance struct {
	ConfigPath      string
	Config          config
	Session         *discordgo.Session
	Storage         storage
	Me              *discordgo.User
	Owner           *discordgo.User
	DebugMode       bool
	shutdown        chan os.Signal
	PermissionsPath string
	Permissions     permissionsRoot
}

// Duder contains the bot instance
var Duder = &Instance{}

func init() {
	flag.StringVar(&Duder.ConfigPath, "config", "duder.toml", "Location of the configuration file, if not found it will be generated (default duder.toml)")
	flag.StringVar(&Duder.PermissionsPath, "permissions", "duder.json", "Location of the permissions file (default duder.json)")
	flag.BoolVar(&Duder.DebugMode, "debug", true, "Enable debug mode")
	flag.Parse()

	log.Printf("Duder version %s", VERSION)
}

func main() {
	// intitialize shutdown channel.
	Duder.shutdown = make(chan os.Signal, 1)

	// load the configuration file
	if err := LoadConfig(Duder.ConfigPath); err != nil {
		log.Fatal("Failed to load configuration file, ", err)
	}

	// load the permissions file
	if err := LoadPermissions(Duder.PermissionsPath); err != nil {
		log.Fatal("Failed to load permissions file, ", err)
	}

	// load the rugs
	if err := LoadRugs(Duder.Config.RugPath); err != nil {
		log.Fatal("Failed to load rugs, ", err)
	}

	// create the Discord session
	log.Printf("Creating Discord session with token '%v'", Duder.Config.BotToken)
	session, err := discordgo.New(Duder.Config.BotToken)
	if err != nil {
		log.Fatal("Error creating discord session, ", err)
	}
	Duder.Session = session

	// obtain bot account details
	log.Println("Obtaining bot account details")
	me, err := Duder.Session.User("@me")
	if err != nil {
		log.Fatal("Error obtaining bot account details, ", err)
	}
	Duder.Me = me
	log.Print("\tBot client ID: ", Duder.Me.ID)

	// obtain owner account details
	log.Println("Obtaining owner account details")
	owner, err := Duder.Session.User(Duder.Config.OwnerID)
	if err != nil {
		log.Fatal("Error obtaining owner account details, ", err)
	}
	Duder.Owner = owner
	log.Print("\tOwner client ID: ", Duder.Owner.ID)

	// register callback for messageCreate
	Duder.Session.AddHandler(onMessageCreate)

	// open the Discord connection
	log.Println("Opening Discord connection")
	err = Duder.Session.Open()
	if err != nil {
		log.Fatal("Error opening discord connection,", err)
	}

	log.Println("Bot is now running.")

	// register bot sg.shutdown channel to receive shutdown signals.
	signal.Notify(Duder.shutdown, syscall.SIGINT, syscall.SIGTERM)

	// wait for shutdown signal
	<-Duder.shutdown

	log.Println("termination signal received; shutting down...")

	// gracefully shut down the bot
	Duder.teardown()

	return
}

// DPrint calls Output to print to the standard logger when debug mode is enabled. Arguments are handled in the manner of fmt.Print.
func (duder *Instance) DPrint(v ...interface{}) {
	if duder.DebugMode {
		c := color.New(color.FgYellow)
		c.Println(v...)
	}
}

// DPrintf calls Output to print to the standard logger when debug mode is enabled. Arguments are handled in the manner of fmt.Printf.
func (duder *Instance) DPrintf(format string, v ...interface{}) {
	if duder.DebugMode {
		c := color.New(color.FgYellow)
		c.Printf(format, v...)
		fmt.Println("")
	}
}

// Shutdown sends Shutdown signal to the bot's Shutdown channel.
func (duder *Instance) Shutdown() {
	duder.shutdown <- os.Interrupt
}

// teardown gracefully releases all resources and saves data before Shutdown.
func (duder *Instance) teardown() (err error) {
	// Perform teardown for commands.
	//sg.rootCommand.teardown(sg)

	// close discord session.
	err = duder.Session.Close()
	if err != nil {
		return
	}
	return
}

// SendMessageToChannel description
func (duder *Instance) SendMessageToChannel(channelID string, content string) {
	duder.Session.ChannelMessageSend(channelID, content)
}

func onMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if strings.HasPrefix(message.Content, fmt.Sprintf("%s ", Duder.Config.Prefix)) {
		Duder.DPrint("Proccessing command", message.Content)
		RunCommand(session, message)
	}
}
