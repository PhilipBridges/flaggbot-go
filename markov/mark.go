package markov

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mongodb/mongo-go-driver/mongo"

	"github.com/mb-14/gomarkov"
	"github.com/mongodb/mongo-go-driver/bson"

	structs "gobot/structs"
)

var (
	mark *mongo.Collection
)

func main() {

}

func AddToModel(order int, s *discordgo.Session, m *discordgo.MessageCreate, mark *mongo.Collection) {
	var foundFile []string
	// Collect messages to insert to chain
	var args []string
	messages, err := s.ChannelMessages(m.ChannelID, 1, "", "", "")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, msg := range messages {
		if msg.Author.ID != s.State.User.ID &&
			(strings.HasPrefix(msg.Content, "!") == false) &&
			(strings.HasPrefix(msg.Content, "http") == false) &&
			(strings.Contains(msg.Content, "https") == false) &&
			(strings.HasPrefix(msg.Content, "<@") == false) &&
			(strings.HasPrefix(msg.Content, "youtube") == false) &&
			(strings.HasPrefix(msg.Content, "@") == false) {
			args = append(args, msg.Content)
		}

	}
	foundFile = GetDataset(order, s, m, mark)

	joinedFile := strings.Join(foundFile, " ")
	update := bson.D{
		{"$set", bson.D{
			{"text", joinedFile + fmt.Sprintf(" %s", strings.Join(args, " "))},
		}},
	}

	if err != nil {
		fmt.Println(err)
	}

	filter := bson.D{{
		"chain", false,
	}}

	_, err = mark.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		fmt.Println(err)
	}
}

// BuildModel builds the mode from Mongo data
func BuildModel(order int, s *discordgo.Session, m *discordgo.MessageCreate, mark *mongo.Collection) {
	var foundFile []string
	var args []string

	// Collect messages to insert to chain
	messages, err := s.ChannelMessages(m.ChannelID, 100, "", "", "")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, msg := range messages {
		if msg.Author.ID != s.State.User.ID &&
			(strings.HasPrefix(msg.Content, "!") == false) &&
			(strings.HasPrefix(msg.Content, "http") == false) &&
			(strings.HasPrefix(msg.Content, "<@") == false) &&
			(strings.HasPrefix(msg.Content, "@") == false) {
			args = append(args, msg.Content)
		}
	}
	//

	foundFile = GetDataset(order, s, m, mark)
	if len(foundFile) < 2 {
		insert := bson.D{
			{
				"chain", false,
			},
			{
				"text", (fmt.Sprintf(" %s", strings.Join(args, " "))),
			},
		}
		mark.InsertOne(context.TODO(), insert)
		s.ChannelMessageSend(m.ChannelID, "Getting initial data...")
		return
	} else {
		s.ChannelMessageSend(m.ChannelID, "Seed data already created")
		return
	}

}

func CallGen(order int, s *discordgo.Session, m *discordgo.MessageCreate, mark *mongo.Collection) {
	chain := gomarkov.NewChain(order)
	var foundFile []string
	var botMessage string

	foundFile = GetDataset(order, s, m, mark)

	chain.Add(foundFile)

	// Randomize the slice to generate a sentence
	rand.Seed(time.Now().UnixNano())

	botText := generateMarkov(chain)

	randStart := randomInt(int(botText[0]), len(botText)-1)

	if len(botText) > 281 {
		if randStart+281 < len(botText) {
			botMessage = botText[randStart : randStart+181]
		} else {
			botMessage = botText[randStart : len(botText)-1]
		}
	}
	if len(botText) < 281 && len(botText) > 50 {
		botMessage = "/tts" + botText[randStart:len(botText)-1]
	}

	s.ChannelMessageSend(m.ChannelID, botMessage)
}

func generateMarkov(chain *gomarkov.Chain) string {
	order := chain.Order
	tokens := make([]string, 0)
	for i := 0; i < order; i++ {
		tokens = append(tokens, gomarkov.StartToken)
	}
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		next, _ := chain.Generate(tokens[(len(tokens) - order):])
		tokens = append(tokens, next)
	}
	str := strings.Join(tokens[order:len(tokens)-1], " ")
	return str
}

// GetDataset fetches the stored data
func GetDataset(order int, s *discordgo.Session, m *discordgo.MessageCreate, mark *mongo.Collection) []string {
	var foundFile structs.File

	filter := bson.D{{
		"chain", false,
	}}

	err := mark.FindOne(context.TODO(), filter).Decode(&foundFile)

	if err != nil {
		fmt.Println(err)
	}
	return strings.Split(foundFile.Text, " ")
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func split(str string) []string {
	return strings.Split(str, "")
}

// func SaveModel(chain *gomarkov.Chain, mark *mongo.Collection) {
// 	var foundFile structs.MarkChain

// 	res, err := bson.Marshal(chain)

// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	insert := bson.D{
// 		{
// 			"chain", true,
// 		},
// 		{
// 			"data", res,
// 		},
// 	}

// 	filter := bson.D{{
// 		"chain", true,
// 	}}

// 	err = mark.FindOne(context.TODO(), filter).Decode(&foundFile)
// 	if foundFile.Chain != true {
// 		_, err := mark.InsertOne(context.TODO(), insert)

// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		return
// 	}

// 	test := generateMarkov(chain)

// 	fmt.Println(test)

// 	update := bson.D{
// 		{"$set", bson.D{
// 			{
// 				"chain", true,
// 			},
// 			{
// 				"data", test,
// 			},
// 		}},
// 	}

// 	_, err = mark.UpdateOne(context.TODO(), filter, update)

// 	if err != nil {
// 		fmt.Println(err)
// 	}

// }

// func loadModel() (*gomarkov.Chain, error) {
// 	var chain gomarkov.Chain
// 	fmt.Println("loading...")
// 	data, err := ioutil.ReadFile("model.json")
// 	if err != nil {
// 		return &chain, err
// 	}
// 	err = json.Unmarshal(data, &chain)
// 	if err != nil {
// 		return &chain, err
// 	}
// 	return &chain, nil
// }
