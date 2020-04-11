package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mongodb/mongo-go-driver/mongo"

	markovHandler "github.com/PhilipBridges/flaggbot-go/markov"
	"github.com/PhilipBridges/flaggbot-go/structs"

	handlers "github.com/PhilipBridges/flaggbot-go/handlers"
)

// Variables used for command line parameters
var (
	Token    string
	client   *mongo.Client
	memes    *mongo.Collection
	profiles *mongo.Collection
	mark     *mongo.Collection
	dg       *discordgo.Session

	count            int
	token            string
	prefix           string
	youtubeToken     string
	voiceConnections []structs.Voice
	queue            []structs.Song
	stopChannel      chan bool
)

func init() {
	// Token is passed as a flag, or you can just put it here
	flag.StringVar(&Token, "t", "NTQ2MDIwMzkwNzg3NDgxNjMw.D0iJvQ.6Kkw_rYrr1Us5hABRxyhJFz_3tE", "Bot Token")
	flag.Parse()
}

func main() {
	if len(os.Args) >= 4 {
		token = os.Args[1]
		youtubeToken = os.Args[3]
		prefix = os.Args[4]
		fmt.Println("Configuration loaded from params")
	} else if handlers.LoadConfiguration() {
		fmt.Println("Configuration loaded from JSON config file")
	} else {
		fmt.Println("Please enter a token, a youtube token and a prefix or add a config.json file")
		return
	}

	stopChannel = make(chan bool)
	// START MONGODB STUFF *****
	// Put your mongo DB address here
	client, err := mongo.Connect(context.TODO(), "mongodb://admin:admin1@ds141631.mlab.com:41631/botdb")

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())
	fmt.Println("Connected to mongoDB")

	memes = client.Database("botdb").Collection("memes")
	profiles = client.Database("botdb").Collection("profiles")
	mark = client.Database("botdb").Collection("markov")
	// END MONGODB STUFF *****

	// Creates a new Discord session using -t flag
	dg, err = discordgo.New("Bot " + Token)
	defer dg.Close()
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageHandler)
	dg.AddHandler(memeHandler)
	dg.AddHandler(profileHandler)
	go dg.AddHandler(markHandler)

	// Seperate from other handlers because it uses voice stuff
	go dg.AddHandler(handlers.CommandHandler)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		panic(err)
	}

	// This deletes command spam every tick
	go func() {
		c := time.Tick(time.Minute * 5)
		for range c {
			handlers.AutoClean(dg)
		}
	}()

	go func() {
		c := time.Tick(time.Minute * 160)
		for range c {
			handlers.AutoHelp(dg)
		}
	}()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

// Tweak order or something
func markHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Content) >= 1 && m.Author.ID != s.State.User.ID {
		markovHandler.AddToModel(2, s, m, mark)
	}

	if m.Content == "!mark" {
		markovHandler.BuildModel(1, s, m, mark)
	}

	if m.Content == "!gen" {
		markovHandler.CallGen(2, s, m, mark)
	}
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID && m.Content != "!clear" {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "!help" {
		s.ChannelMessageSend(m.ChannelID,
			">Commands are: \n!stop (stop all audio)\n!create (profile)\n!bet (x)\n!clean/!clear (delete last 100 bot related messages)\n!fbux\n!broke (irreversibly resets flaggbux to 100)\n!meme\n!memecount\n!u (sound)\n!gear (sound)\n!mark (create markov chain data, only used once)\n!gen (generate random sentence)x")
	}

	if m.Content == "!clean" || m.Content == "!clear" {
		handlers.Clean(s, m)
	}
}

func profileHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!create" {
		handlers.CreateProfile(s, m, profiles)
	}

	if strings.Contains(m.Content, "!bet") {
		handlers.HandleBet(s, m, profiles)
	}

	if m.Content == "!fbux" {
		handlers.ShowMoney(s, m, profiles)
	}

	if m.Content == "!broke" {
		handlers.Reset(s, m, profiles)
	}

	if m.Content == "!top" {
		handlers.Top(s, m, profiles)
	}

}

func memeHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!meme" {
		handlers.AddMeme(s, m, memes)
	}

	if m.Content == "!memecount" {
		handlers.MemeCount(s, m, memes)
	}
}

// *** Only for init meme count
// meme := structs.Meme{true, 0}

// insertResult, err := memes.InsertOne(context.TODO(), meme)
// if err != nil {
// 	log.Fatal(err)
// }
// ***
