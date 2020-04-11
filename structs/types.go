package structs

import "github.com/bwmarrin/discordgo"

type Voice struct {
	VoiceConnection *discordgo.VoiceConnection
	Channel         string
	Guild           string
	PlayerStatus    int
}

type Configuration struct {
	Token           string `json:"token"`
	Prefix          string `json:"prefix"`
	SoundcloudToken string `json:"soundcloud_token"`
	YoutubeToken    string `json:"youtube_token"`
}

type Song struct {
	Link    string
	Type    string
	Guild   string
	Channel string
}

type SoundcloudResponse struct {
	Link  string `json:"stream_url"`
	Title string `json:"title"`
}

type YoutubeRoot struct {
	Items []YoutubeVideo `json:"items"`
}

type YoutubeVideo struct {
	Snippet YoutubeSnippet `json:"snippet"`
}

type YoutubeSnippet struct {
	Resource ResourceID `json:"resourceId"`
}

type ResourceID struct {
	VideoID string `json:"videoId"`
}
