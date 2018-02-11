package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/foszor/duder/helpers/rugutils"
	"github.com/go-fsnotify/fsnotify"
)

// VERSION contains the current version
const VERSION string = "0.0.1"

// instance struct describes the bot
type instance struct {
	configPath      string
	config          config
	session         *discordgo.Session
	me              *discordgo.User
	owner           *discordgo.User
	debug           bool
	shutdownSignal  chan os.Signal
	permissionsPath string
	permissions     permissions
	rugHotload      bool
	rugHotloader    *fsnotify.Watcher
	avatarPath      string
}

// Duder contains the bot instance
var Duder = &instance{}

func init() {
	flag.StringVar(&Duder.configPath, "config", "config.toml", "Location of the configuration file, if not found it will be generated (default config.toml)")
	flag.StringVar(&Duder.permissionsPath, "permissions", "permissions.json", "Location of the permissions file (default permissions.json)")
	flag.BoolVar(&Duder.debug, "debug", true, "Enable debug mode")
	flag.BoolVar(&Duder.rugHotload, "hotload", true, "Enable or disable rug hotloading")
	flag.StringVar(&Duder.avatarPath, "avatarPath", "avatars", "Location of avatars (default avatars)")
	flag.Parse()

	log.Printf("Duder version %s", VERSION)
}

func main() {
	// intitialize shutdown channel.
	Duder.shutdownSignal = make(chan os.Signal, 1)

	// load the configuration file
	if err := loadConfig(Duder.configPath); err != nil {
		log.Fatal("Failed to load configuration file, ", err)
	}

	// load the permissions file
	if err := loadPermissions(Duder.permissionsPath); err != nil {
		log.Fatal("Failed to load permissions file, ", err)
	}

	// load the rugs
	if err := loadRugs(Duder.config.RugPath); err != nil {
		log.Fatal("Failed to load rugs, ", err)
	}

	// observe rugs for hotloading
	if Duder.rugHotload {
		if rugHotloader, err := observeRugs(Duder.config.RugPath); err != nil {
			Duder.dprint("Failed to monitor rugs for hotload, ", err)
		} else {
			Duder.rugHotloader = rugHotloader
		}
	}

	// create the Discord session
	log.Printf("Creating Discord session with token '%v'", Duder.config.BotToken)
	session, err := discordgo.New(Duder.config.BotToken)
	if err != nil {
		log.Fatal("Error creating discord session, ", err)
	}
	Duder.session = session

	// obtain bot account details
	log.Println("Obtaining bot account details")
	me, err := Duder.session.User("@me")
	if err != nil {
		log.Fatal("Error obtaining bot account details, ", err)
	}
	Duder.me = me
	log.Println("\tBot client ID: ", Duder.me.ID)

	// obtain owner account details
	log.Println("Obtaining owner account details")
	owner, err := Duder.session.User(Duder.config.OwnerID)
	if err != nil {
		log.Fatal("Error obtaining owner account details, ", err)
	}
	Duder.owner = owner
	log.Print("\tOwner client ID: ", Duder.owner.ID)

	// register callback for messageCreate
	Duder.session.AddHandler(onMessageCreate)

	// open the Discord connection
	log.Println("Opening Discord connection")
	err = Duder.session.Open()
	if err != nil {
		log.Fatal("Error opening discord connection,", err)
	}

	Duder.setStatus("with Maude")
	log.Println("Bot is now running.")

	// register bot sg.shutdown channel to receive shutdown signals.
	signal.Notify(Duder.shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	// wait for shutdown signal
	<-Duder.shutdownSignal

	log.Println("termination signal received; shutting down...")

	// gracefully shut down the bot
	Duder.teardown()

	return
}

// dprint calls Output to print to the standard logger when debug mode is enabled. Arguments are handled in the manner of fmt.Print.
func (duder *instance) dprint(v ...interface{}) {
	if duder.debug {
		c := color.New(color.FgYellow)
		c.Println(v...)
	}
}

// dprintf calls Output to print to the standard logger when debug mode is enabled. Arguments are handled in the manner of fmt.Printf.
func (duder *instance) dprintf(format string, v ...interface{}) {
	if duder.debug {
		c := color.New(color.FgYellow)
		c.Printf(format, v...)
		fmt.Println("")
	}
}

// wprint calls Output to print to the standard logger when debug mode is enabled. Arguments are handled in the manner of fmt.Printf.
func (duder *instance) wprint(v ...interface{}) {
	if duder.debug {
		c := color.New(color.FgHiYellow)
		c.Println(v...)
	}
}

// wprintf calls Output to print to the standard logger when debug mode is enabled. Arguments are handled in the manner of fmt.Printf.
func (duder *instance) wprintf(format string, v ...interface{}) {
	if duder.debug {
		c := color.New(color.FgHiYellow)
		c.Printf(format, v...)
		fmt.Println("")
	}
}

// shutdown sends Shutdown signal to the bot's Shutdown channel.
func (duder *instance) shutdown() {
	duder.shutdownSignal <- os.Interrupt
}

// teardown gracefully releases all resources and saves data before Shutdown.
func (duder *instance) teardown() (err error) {
	// Perform teardown for commands.
	//sg.rootCommand.teardown(sg)

	// close discord session.
	err = duder.session.Close()
	if err != nil {
		return
	}

	if duder.rugHotloader != nil {
		defer duder.rugHotloader.Close()
	}

	return
}

// sendMessageToChannel description
func (duder *instance) sendMessageToChannel(channelID string, content string) {
	duder.session.ChannelMessageSend(channelID, content)
}

// setStatus description
func (duder *instance) setStatus(status string) error {
	if err := duder.session.UpdateStatus(0, status); err != nil {
		return errors.New("unable to set status")
	}
	return nil
}

func getMessageGuild(session *discordgo.Session, message *discordgo.MessageCreate) (*discordgo.Guild, error) {
	channel, err := getMessageChannel(session, message)
	if err != nil {
		return nil, err
	}
	guild, err := session.Guild(channel.GuildID)
	if err != nil {
		return nil, errors.New("Unable to get guild")
	}

	return guild, nil
}

func getMessageChannel(session *discordgo.Session, message *discordgo.MessageCreate) (*discordgo.Channel, error) {
	channel, err := session.Channel(message.ChannelID)
	if err != nil {
		return nil, errors.New("Unable to get channel")
	}

	return channel, nil
}

// onMessageCreate description
func onMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if strings.HasPrefix(message.Content, fmt.Sprintf("%s ", Duder.config.Prefix)) {
		Duder.dprint("Proccessing command", message.Content)
		runCommand(session, message)
	}
}

// runCommand description
func runCommand(session *discordgo.Session, message *discordgo.MessageCreate) {
	// strip the command prefix from the message content
	content := message.Content[len(Duder.config.Prefix)+1 : len(message.Content)]
	content = strings.TrimSpace(content)

	// get the root command
	args := rugutils.ParseArguments(content)
	if len(args) == 0 {
		return
	}
	cmd := strings.ToLower(args[0])
	Duder.dprintf("Root command '%s'", cmd)

	// hardcoded commands
	switch cmd {
	case "reload":
		if Duder.config.OwnerID == message.Author.ID {
			loadRugs(Duder.config.RugPath)
			if len(rugLoadErrors) > 0 {
				session.ChannelMessageSend(message.ChannelID, ":octagonal_sign: Rugs reloaded with errors.")
			} else {
				session.ChannelMessageSend(message.ChannelID, ":ok_hand: Rugs successfully reloaded.")
			}
		} else {
			session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%s, you don't have permissions for that.", message.Author.Username))
		}
		return
	case "shutdown":
		if Duder.config.OwnerID == message.Author.ID {
			session.ChannelMessageSend(message.ChannelID, "Goodbye.")
			Duder.shutdown()
		} else {
			session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%s, you don't have permissions for that.", message.Author.Username))
		}
		return
	}

	// check each rug to find the matching command
	for _, rug := range rugMap {
		for _, rugCmd := range rug.commands {
			if rugCmd.trigger == cmd {
				execRugCommand(rug, rugCmd, session, message, args)
				return
			}
		}
	}
}
