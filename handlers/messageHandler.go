package handlers

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	dg *discordgo.Session
)

// Clean clears out the last 100 messages from the bot + the message used to call this function
func Clean(s *discordgo.Session, m *discordgo.MessageCreate) {
	var args []string

	messages, err := s.ChannelMessages(m.ChannelID, 100, "", "", "")
	if err != nil {
		return
	}
	for _, msg := range messages {
		if msg.Author.ID == s.State.User.ID || (strings.HasPrefix(msg.Content, "!") == true) && (strings.HasPrefix(msg.Content, ">Commands") == false) {
			args = append(args, msg.ID)
		}

	}
	err = s.ChannelMessagesBulkDelete(m.ChannelID, args)

	if err != nil {
		fmt.Println(err)
		s.ChannelMessageSend(m.ChannelID, "Not enough messages to delete.")
	}
}

func AutoClean(dg *discordgo.Session) {
	var args []string

	messages, err := dg.ChannelMessages("119259349875949569", 100, "", "", "")
	if err != nil {
		return
	}
	for _, msg := range messages {
		if msg.Author.ID == dg.State.User.ID || (strings.HasPrefix(msg.Content, "!") == true) {
			args = append(args, msg.ID)
		}

	}
	err = dg.ChannelMessagesBulkDelete("119259349875949569", args)

	if err != nil {
		fmt.Println(err)
	}
}

func AutoHelp(dg *discordgo.Session) {
	_, err := dg.ChannelMessageSend("119259349875949569", "Commands are: \n!stop (stop all audio)\n!create (profile)\n!bet (x)\n!clean/!clear (delete last 100 bot related messages)\n!fbux\n!broke (irreversibly resets flaggbux to 100)\n!meme\n!memecount\n!u (sound)\n!gear (sound)\n!mark (create markov chain data, only used once)\n!gen (generate random sentence)")

	if err != nil {
		fmt.Println(err)
	}
}
