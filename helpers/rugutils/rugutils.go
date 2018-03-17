package rugutils

import (
	"bytes"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// ConvertMentions description
func ConvertMentions(guildID string, message *discordgo.MessageCreate) string {
	var buffer bytes.Buffer

	buffer.WriteString("new Array( ")
	for i, mention := range message.Mentions {
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(fmt.Sprintf("new DuderUser(\"%s\",\"%s\",\"%s\")", guildID, mention.ID, mention.Username))
	}
	buffer.WriteString(" )")
	return buffer.String()
}

// ConvertArguments description
func ConvertArguments(args []string) string {
	var buffer bytes.Buffer

	buffer.WriteString("new Array( ")
	for i, arg := range args {
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(fmt.Sprintf("\"%s\"", arg))
	}
	buffer.WriteString(" )")
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
	buffer.WriteString(" )")
	return buffer.String()
}
