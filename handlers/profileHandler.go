package handlers

import (
	"context"
	"fmt"
	"gobot/structs"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/mongodb/mongo-go-driver/mongo/options"

	"github.com/bwmarrin/discordgo"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

func CreateProfile(s *discordgo.Session, m *discordgo.MessageCreate, profiles *mongo.Collection) {
	var foundProfile structs.Profile

	profileFilter := bson.D{{
		"username", m.Author.Username,
	}}

	err := profiles.FindOne(context.TODO(), profileFilter).Decode(&foundProfile)

	if foundProfile.Enabled == true {
		s.ChannelMessageSend(m.ChannelID, "Profile already created.")
		return
	}

	data := bson.D{{"enabled", true}, {"username", m.Author.Username}, {"flaggbux", 100}}

	_, err = profiles.InsertOne(context.TODO(), data)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not create profile.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Profile has been created."))
}

// HandleBet ...handles the bet
func HandleBet(s *discordgo.Session, m *discordgo.MessageCreate, profiles *mongo.Collection) {

	var profile structs.Profile
	var bet int
	var filter bson.D
	var update bson.D
	var split []string

	profileFilter := bson.D{{
		"username", m.Author.Username,
	}}

	profileErr := profiles.FindOne(context.TODO(), profileFilter).Decode(&profile)

	if profileErr != nil {
		s.ChannelMessageSend(m.ChannelID, "No profile found. Type '!create' to create your profile.")
		return
	}

	rand.Seed(time.Now().UnixNano())

	if len(m.Content) > 4 {
		split = strings.Split(m.Content, " ")
	} else {
		s.ChannelMessageSend(m.ChannelID, "You must put an amount to bet.")
		return
	}
	if len(split) == 1 {
		s.ChannelMessageSend(m.ChannelID, "You made a typo.")
		return
	}
	
	bet, err := strconv.Atoi(split[1])

	if bet > profile.Flaggbux {
		s.ChannelMessageSend(m.ChannelID, "You don't have enough flaggbux.")
		return
	}

	if bet > 2500 || bet < 1 {
		s.ChannelMessageSend(m.ChannelID, "You must bet between 1-2500 flaggbux.")
		return
	}

	if bet <= profile.Flaggbux && bet >= 1 {
		roll := randomInt(1, 10)

		if roll >= 5 {
			filter = bson.D{{"username", m.Author.Username}}
			update = bson.D{
				{"$inc", bson.D{
					{"flaggbux", -bet},
				}},
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You lost %d!", bet))
		} else {
			filter = bson.D{{"username", m.Author.Username}}
			update = bson.D{
				{"$inc", bson.D{
					{"flaggbux", bet * 2},
				}},
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You won %d!", bet*2))
		}

		_, err = profiles.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not update.")
			return
		}
	}
}

// ShowMoney displays your current Flaggbux
func ShowMoney(s *discordgo.Session, m *discordgo.MessageCreate, profiles *mongo.Collection) {
	var profile structs.Profile
	profileFilter := bson.D{{
		"username", m.Author.Username,
	}}

	profileErr := profiles.FindOne(context.TODO(), profileFilter).Decode(&profile)

	if profileErr != nil {
		s.ChannelMessageSend(m.ChannelID, "You don't have a profile. Use the '!create' command to set one up automatically.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You have %d flaggbux.", profile.Flaggbux))
}

func Top(s *discordgo.Session, m *discordgo.MessageCreate, profiles *mongo.Collection) {
	var lmt int64
	var profileSlice []structs.Profile

	filter := bson.D{{}}
	byTop := bson.D{{"flaggbux", -1}}

	result, profileErr := profiles.Find(context.TODO(), filter, &options.FindOptions{Limit: &lmt, Sort: byTop})

	if profileErr != nil {
		fmt.Println(profileErr)
		s.ChannelMessageSend(m.ChannelID, "Could not find top flaggbux owners.")
		return
	}

	ctx := context.Background()
	defer result.Close(ctx)

	for result.Next(nil) {
		item := structs.Profile{}
		err := result.Decode(&item)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not find top flaggbux owners.")
			return
		}
		profileSlice = append(profileSlice, item)
	}
	var strArr []string

	for i, profile := range profileSlice {
		strArr = append(strArr, fmt.Sprintf("%d: %s: %s flaggbux", i+1, profile.Username, strconv.Itoa(profile.Flaggbux)))
	}
	newStr := strings.Join(strArr, "\n")
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v", newStr))

}

// Reset sets your flaggbux back to 100
func Reset(s *discordgo.Session, m *discordgo.MessageCreate, profiles *mongo.Collection) {

	filter := bson.D{{"username", m.Author.Username}}
	update := bson.D{
		{"$inc", bson.D{
			{"flaggbux", 100},
		}},
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Flaggbux reset to 100."))

	_, err := profiles.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Command failed. :()")
		return
	}
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}
