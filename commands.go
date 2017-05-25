package main

// Command struct describes commands
type Command struct {
	Trigger     string
	Permissions []int
	SubCommands []*Command
}
