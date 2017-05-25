package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// VERSION contains the current version
const VERSION string = "0.0.1"

const (
	// PermissionNone is given to everyone
	PermissionNone = 0
	// PermissionOwner only the bot owner can use
	PermissionOwner = 1
)

// Instance struct describes the bot
type Instance struct {
	ConfigPath string
	RugPath    string
	Config     config
	Session    *discordgo.Session
	Me         *discordgo.User
	Owner      *discordgo.User
	DebugMode  bool
	shutdown   chan os.Signal
}

// Duder contains the bot instance
var Duder = &Instance{}

func init() {
	flag.StringVar(&Duder.ConfigPath, "config", "duder.toml", "location of the config file, if not found it will be generated (default duder.toml)")
	flag.StringVar(&Duder.RugPath, "rugpath", "rugs", "directory of the rug files (default rugs)")
	flag.BoolVar(&Duder.DebugMode, "debug", true, "enable debug mode")
	flag.Parse()
}

func main() {
	// intitialize shutdown channel.
	Duder.shutdown = make(chan os.Signal, 1)

	// load the rugs
	log.Print("loading rugs from folder: ", Duder.RugPath)
	LoadRugs(Duder.RugPath)
	RunCommand("testcmd1")
	RunCommand("testcmd2")

	os.Exit(0)

	// load the configuration file
	log.Print("loading configuration file: ", Duder.ConfigPath)
	if err := LoadConfig(Duder.ConfigPath); err != nil {
		log.Fatal("failed to load configuration file,", err)
	}

	// create the Discord session
	log.Printf("creating Discord session with token '%v'", Duder.Config.BotToken)
	session, err := discordgo.New(Duder.Config.BotToken)
	if err != nil {
		log.Fatal("error creating discord session, ", err)
	}
	Duder.Session = session

	// obtain bot account details
	log.Println("obtaining bot account details")
	me, err := Duder.Session.User("@me")
	if err != nil {
		log.Fatal("error obtaining bot account details, ", err)
	}
	Duder.Me = me
	log.Print("bot client ID: ", Duder.Me.ID)

	// obtain owner account details
	log.Println("obtaining owner account details")
	owner, err := Duder.Session.User(Duder.Config.OwnerID)
	if err != nil {
		log.Fatal("error obtaining owner account details, ", err)
	}
	Duder.Owner = owner
	log.Print("owner client ID: ", Duder.Owner.ID)

	// register callback for messageCreate
	Duder.Session.AddHandler(onMessageCreate)

	// open the Discord connection
	log.Println("opening Discord connection")
	err = Duder.Session.Open()
	if err != nil {
		log.Fatal("error opening discord connection,", err)
	}

	log.Println("bot is now running.")

	// register bot sg.shutdown channel to receive shutdown signals.
	signal.Notify(Duder.shutdown, syscall.SIGINT, syscall.SIGTERM)

	// wait for shutdown signal
	<-Duder.shutdown

	log.Println("termination signal received; shutting down...")

	// gracefully shut down the bot
	Duder.teardown()

	return
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

// DPrint calls Output to print to the standard logger when debug mode is enabled. Arguments are handled in the manner of fmt.Print.
func (duder *Instance) DPrint(v ...interface{}) {
	if duder.DebugMode {
		log.Print(v...)
	}
}

// DPrintf calls Output to print to the standard logger when debug mode is enabled. Arguments are handled in the manner of fmt.Printf.
func (duder *Instance) DPrintf(format string, v ...interface{}) {
	if duder.DebugMode {
		log.Printf(format, v...)
	}
}

func onMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Content == "!d quit" {
		Duder.Shutdown()
	}
}
