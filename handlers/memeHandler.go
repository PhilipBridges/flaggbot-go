package handlers

import (
	"context"
	"fmt"
	"gobot/structs"

	"github.com/bwmarrin/discordgo"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

var (
	collection *mongo.Collection
)

// AddMeme adds 1 meme to the total number of memes
func AddMeme(s *discordgo.Session, m *discordgo.MessageCreate, collection *mongo.Collection) {
	update := bson.D{
		{"$inc", bson.D{
			{"count", 1},
		}},
	}

	filter := bson.D{{"enabled", true}}

	_, err := collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return
	}

	var meme structs.Meme

	err = collection.FindOne(context.TODO(), filter).Decode(&meme)

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Meme count is now %d", meme.Count))
}

// MemeCount displays the total amount of times the !meme command is used
func MemeCount(s *discordgo.Session, m *discordgo.MessageCreate, memes *mongo.Collection) {
	var meme structs.Meme

	filter := bson.D{{"enabled", true}}
	err := memes.FindOne(context.TODO(), filter).Decode(&meme)

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Meme count could not be loaded.")
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Meme count is: %d", meme.Count))
}
