package rugutils

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// ParseArguments description
func ParseArguments(content string) []string {
	content = strings.TrimSpace(content)
	inQuote := false
	var args []string
	var arg string

	for _, c := range content {
		if inQuote {
			if c == '"' {
				inQuote = false
				args = append(args, arg)
				arg = ""
			} else {
				arg = arg + string(c)
			}
		} else {
			if c == ' ' {
				if len(arg) > 0 {
					args = append(args, arg)
					arg = ""
				}
			} else if c == '"' {
				inQuote = true
			} else {
				arg = arg + string(c)
			}
		}
	}

	if len(arg) > 0 {
		args = append(args, arg)
	}

	return args
}

// ConvertMentions description
func ConvertMentions(message *discordgo.MessageCreate) string {
	var buffer bytes.Buffer

	buffer.WriteString("new Array( ")
	for i, mention := range message.Mentions {
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(fmt.Sprintf("new DuderUser(\"%s\",\"%s\")", mention.ID, mention.Username))
	}
	buffer.WriteString(")")
	return buffer.String()
}

// ConvertArgs description
func ConvertArgs(args []string) string {
	var buffer bytes.Buffer

	buffer.WriteString("new Array( ")
	for i, arg := range args {
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(fmt.Sprintf("\"%s\"", arg))
	}
	buffer.WriteString(")")
	return buffer.String()
}

// ConvertUserPermission description
func ConvertUserPermission(perms []int) string {
	var buffer bytes.Buffer

	buffer.WriteString("new Array( ")
	for i, p := range perms {
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(fmt.Sprintf("\"%v\"", p))
	}
	buffer.WriteString(")")
	return buffer.String()
}
