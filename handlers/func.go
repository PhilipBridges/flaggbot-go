package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/PhilipBridges/flaggbot-go/structs"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

var (
	token            string
	prefix           string
	youtubeToken     string
	voiceConnections []structs.Voice
	queue            []structs.Song
	stopChannel      chan bool
)

const (
	IS_PLAYING     = iota
	IS_NOT_PLAYING = iota
)

func LoadConfiguration() bool {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return false
	}
	var config structs.Configuration
	err = json.Unmarshal(file, &config)
	if err != nil {
		return false
	}
	token = config.Token
	prefix = config.Prefix
	youtubeToken = config.YoutubeToken
	return true
}

func CommandHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	var commandArgs []string = strings.Split(m.Content, " ")
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		fmt.Println(err)
	}
	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		fmt.Println(err)
	}
	voiceChannel := findVoiceChannelID(guild, m)
	if commandArgs[0] == prefix+"connect" {
		voiceConnections = append(voiceConnections, connectToVoiceChannel(s, channel.GuildID, voiceChannel))
	} else if commandArgs[0] == prefix+"disconnect" {
		disconnectFromVoiceChannel(channel.GuildID, voiceChannel)
	} else if commandArgs[0] == prefix+"play" {
		go playAudioFile(sanitizeLink(commandArgs[1]), channel.GuildID, voiceChannel, "web")
	} else if commandArgs[0] == prefix+"stop" {
		disconnectFromVoiceChannel(channel.GuildID, voiceChannel)
	}
}

// This function is used to close the connection to a voiceChannel. When called, it will crawl
// the list of opened voice connections and if one is found corresponding to the parameters, it
// will be closed
func disconnectFromVoiceChannel(guild string, channel string) {
	for index, voice := range voiceConnections {
		if voice.Guild == guild {
			_ = voice.VoiceConnection.Disconnect()
			stopChannel <- true
			voiceConnections = append(voiceConnections[:index], voiceConnections[index+1:]...)
		}
	}
}

// This function will sanitize a link that contains < and >, this is used to handle links with
// disabled embed in Discord
func sanitizeLink(link string) string {
	firstTreatment := strings.Replace(link, "<", "", 1)
	return strings.Replace(firstTreatment, ">", "", 1)
}

// This function is used to extract the id of a playlist given a youtube plaulist link
func parseYoutubePlaylistLink(link string) string {
	standardPlaylistSanitize := strings.Replace(link, "https://www.youtube.com/playlist?list=", "", 1)
	return standardPlaylistSanitize
}

// This function will crawl the voice connections and try to find and return a voice connection
// and its index if one is found
func findVoiceConnection(guild string, channel string) (structs.Voice, int) {
	var voiceConnection structs.Voice
	var index int
	for i, vc := range voiceConnections {
		if vc.Guild == guild {
			voiceConnection = vc
			index = i
		}
	}
	fmt.Println(voiceConnection, index)
	return voiceConnection, index

}

// This function will call the playAudioFile function in a new goroutine if songs are remaining in the
// queue. If there is no song left in the queue, the function return false
func nextSong() bool {
	if len(queue) > 0 {
		go playAudioFile(queue[0].Link, queue[0].Guild, queue[0].Channel, queue[0].Type)
		queue = append(queue[:0], queue[1:]...)
		return true
	} else {
		return false
	}
}

// This function is used to add an item to the queue
func addSong(song structs.Song) {
	queue = append(queue, song)
}

// This function is used to play every audio files, if the program is already playing, the function
// will add the song to the queue and call the nextSonng function when the current song is over
func playAudioFile(file string, guild string, channel string, linkType string) {
	voiceConnection, index := findVoiceConnection(guild, channel)
	switch voiceConnection.PlayerStatus {
	case IS_NOT_PLAYING:
		voiceConnections[index].PlayerStatus = IS_PLAYING
		dgvoice.PlayAudioFile(voiceConnection.VoiceConnection, file, stopChannel)
		voiceConnections[index].PlayerStatus = IS_NOT_PLAYING
	case IS_PLAYING:
		addSong(structs.Song{
			Link:    file,
			Type:    linkType,
			Guild:   guild,
			Channel: channel,
		})
	}
}

// This function allow the user to stop the current playing file
func stopAudioFile(guild string, channel string) {
	_, index := findVoiceConnection(guild, channel)
	voiceConnections[index].PlayerStatus = IS_NOT_PLAYING
	//dgvoice.KillPlayer()
}

// This function allow the bot to find the voice channel id of the user who called the connect command
func findVoiceChannelID(guild *discordgo.Guild, message *discordgo.MessageCreate) string {
	var channelID string

	for _, vs := range guild.VoiceStates {
		if vs.UserID == message.Author.ID {
			channelID = vs.ChannelID
		}
	}
	return channelID
}

// This function allow the user to connect the bot to a channel. It will ask the voice channel id
// of the user to the findVoiceChannelID function and will then call the ChannelVoiceJoin
// of the discordgo.Session instance. Then it checks if the voice connection already exist and
// return a new Voice object
func connectToVoiceChannel(bot *discordgo.Session, guild string, channel string) structs.Voice {
	vs, err := bot.ChannelVoiceJoin(guild, channel, false, true)

	checkForDoubleVoiceConnection(guild, channel)

	if err != nil {
		fmt.Println(err)
	}
	return structs.Voice{
		VoiceConnection: vs,
		Channel:         channel,
		Guild:           guild,
		PlayerStatus:    IS_NOT_PLAYING,
	}

}

// This function check if there is already an existing voice connection for the givent params
func checkForDoubleVoiceConnection(guild string, channel string) {
	for index, voice := range voiceConnections {
		if voice.Guild == guild {
			voiceConnections = append(voiceConnections[:index], voiceConnections[index+1:]...)
		}
	}
}
