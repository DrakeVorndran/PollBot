package slack

import ( //packages

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/nlopes/slack"
	"go.mongodb.org/mongo-driver/bson"
)

/*
   TODO: Change @BOT_NAME to the same thing you entered when creating your Slack application.
   NOTE: command_arg_1 and command_arg_2 represent optional parameteras that you define
   in the Slack API UI
*/

type Command struct {
	Input       string
	Description string
}

func newCommand(In, Desc string) Command {
	return Command{Description: Desc, Input: In}
}

var CreateCommand = Command{Description: "Creates a new poll | `create <pollName> [...poll values]`", Input: "create"}
var ReadCommand = Command{Description: "Gives the current standing of the poll | `read <pollName>`", Input: "read"}
var EndCommand = Command{Description: "Stops a given poll from being voted on | `end <pollName>`", Input: "end"}
var CommandsCommand = Command{Description: "Gives a list of commands, or the details of a speific commands | `commands [command]` ", Input: "commands"}

var Commands = map[string]Command{
	"create":   CreateCommand,
	"read":     ReadCommand,
	"end":      EndCommand,
	"commands": CommandsCommand,
}

/*
   CreateSlackClient sets up the slack RTM (real-timemessaging) client library,
   initiating the socket connection and returning the client.
   DO NOT EDIT THIS FUNCTION. This is a fully complete implementation.
*/
func CreateSlackClient(apiKey string) *slack.RTM { // Functions
	api := slack.New(apiKey)
	rtm := api.NewRTM()
	go rtm.ManageConnection() // goroutine!
	return rtm
}

/*
   RespondToEvents waits for messages on the Slack client's incomingEvents channel,
   and sends a response when it detects the bot has been tagged in a message with @<botTag>.

   EDIT THIS FUNCTION IN THE SPACE INDICATED ONLY!
*/
func RespondToEvents(slackClient *slack.RTM) {
	MongoUri := os.Getenv("MONGO_URI")
	session, err := mgo.Dial(MongoUri)
	client := session.DB("pollbot")

	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	for msg := range slackClient.IncomingEvents {
		fmt.Println("Event Received: ", msg.Type)
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			botTagString := fmt.Sprintf("<@%s> ", slackClient.GetInfo().User.ID)
			if !strings.Contains(ev.Msg.Text, botTagString) {
				continue
			}
			message := strings.Replace(ev.Msg.Text, botTagString, "", -1)

			// TODO: Make your bot do more than respond to a help command. See notes below.
			// Make changes below this line and add additional funcs to support your bot's functionality.
			// sendHelp is provided as a simple example. Your team may want to call a free external API
			// in a function called sendResponse that you'd create below the definition of sendHelp,
			// and call in this context to ensure execution when the bot receives an event.

			// START SLACKBOT CUSTOM CODE
			// ===============================================================
			switch strings.Split(message, " ")[0] {
			case "help":
				sendHelp(slackClient, message, ev.Channel)
			case "commands":
				sendCommands(slackClient, message, ev.Channel)
			case "create":
				sendCreate(slackClient, message, ev.Channel, client)
			case "read":
				sendRead(slackClient, message, ev.Channel, client)
			case "end":
				sendEnd(slackClient, message, ev.Channel, client)
			default:
				slackClient.SendMessage(slackClient.NewOutgoingMessage("I don't know what you want, try @PollBot commands", ev.Channel))
			}
			// ===============================================================
			// END SLACKBOT CUSTOM CODE
		case *slack.ReactionAddedEvent:
			handleReactionUpdate(slackClient, ev.Item.Timestamp, ev.Item.Channel)
		default:

		}
	}
}

func sendEnd(slackClient *slack.RTM, message, slackChannel string, client *mgo.Database) {
	sendRead(slackClient, message, slackChannel, client)
	message = strings.ToLower(message)
	commands := strings.Split(message, " ")
	collection := client.C("polls")
	err := collection.Remove(bson.M{"name": commands[1]})
	if err != nil {
		fmt.Println(err)
	} else {
		slackClient.SendMessage(slackClient.NewOutgoingMessage(fmt.Sprintf("The %s poll has been ended, these are the final results", commands[1]), slackChannel))

	}
}

func sendRead(slackClient *slack.RTM, message, slackChannel string, client *mgo.Database) {
	message = strings.ToLower(message)
	commands := strings.Split(message, " ")
	if len(commands) != 2 {
		slackClient.SendMessage(slackClient.NewOutgoingMessage("I don't understand what you want me to do, use `@pollbot commands read` for more information on this command", slackChannel))
		return
	}

	collection := client.C("polls")
	query := collection.Find(bson.M{"name": commands[1]})
	poll := Poll{}
	err := query.One(&poll)
	if err != nil {
		fmt.Println(err)
		slackClient.SendMessage(slackClient.NewOutgoingMessage("Oops, I don't have that poll", slackChannel))
	}
	msgRef := slack.NewRefToMessage(slackChannel, poll.Timestamp)
	msgReactions, err := slackClient.GetReactions(msgRef, slack.NewGetReactionsParameters())
	if err != nil {
		fmt.Printf("Error getting reactions: %s\n", err)
		return
	}
	reactionString := "The current poll standings - "
	total := 0
	for _, r := range msgReactions {
		total += (r.Count - 1)
	}
	for _, r := range msgReactions {
		percent := float64(r.Count - 1)
		fmt.Println(percent, total)
		percent /= float64(total)
		fmt.Println(percent)
		reactionString += fmt.Sprintf("|%s - :%s:, %d votes, %g%%| ", poll.Items[r.Name], r.Name, r.Count-1, percent*100)
	}
	slackClient.SendMessage(slackClient.NewOutgoingMessage(reactionString, slackChannel))
}

// handleAddReaction handles when a user adds a reaction to a message
func handleReactionUpdate(slackClient *slack.RTM, timestamp, slackChannel string) {
	msgRef := slack.NewRefToMessage(slackChannel, timestamp)
	msgReactions, err := slackClient.GetReactions(msgRef, slack.NewGetReactionsParameters())
	if err != nil {
		fmt.Printf("Error getting reactions: %s\n", err)
		return
	}
	fmt.Printf("\n")
	fmt.Printf("%d reactions to message...\n", len(msgReactions))
	for _, r := range msgReactions {
		fmt.Printf("  %d users say %s\n", r.Count, r.Name)
	}
}

// Poll struct is the struct used to store polls
type Poll struct {
	Name      string
	Timestamp string
	Items     map[string]string
}

// NewPoll creates a new poll struct and returns it
func NewPoll(Name, Timestamp string, pollOptions map[string]string) Poll {
	return Poll{
		Name:      Name,
		Timestamp: Timestamp,
		Items:     pollOptions,
	}
}

func sendCreate(slackClient *slack.RTM, message, slackChannel string, client *mgo.Database) {
	message = strings.ToLower(message)
	commands := strings.Split(message, " ")

	// headerText := slack.NewTextBlockObject("mrkdwn", "You have a new request:\n*<fakeLink.toEmployeeProfile.com|Fred Enriquez - New device request>*", false, false)
	// headerSection := slack.NewSectionBlock(headerText, nil, nil)
	// fmt.Println(headerSection)
	// emojiList := []string{"😀", "😁", "😂", "🤣", "😃", "😄", "😅", "😆", "😉", "😊", "😋", "😎", "😍", "😘", "🥰", "😗", "😙", "😚", "☺", "️", "🙂", "🤗", "🤩", "🤔", "🤨", "😐", "😑", "😶", "🙄", "😏", "😣", "😥", "😮", "🤐", "😯", "😪", "😫", "😴", "😌", "😛", "😜", "😝", "🤤", "😒", "😓", "😔", "😕", "🙃", "🤑", "😲", "☹", "️", "🙁", "😖", "😞", "😟", "😤", "😢", "😭", "😦", "😧", "😨", "😩", "🤯", "😬", "😰", "😱", "🥵", "🥶", "😳", "🤪", "😵", "😡", "😠", "🤬", "😷", "🤒", "🤕", "🤢", "🤮", "🤧", "😇", "🤠", "🤡", "🥳", "🥴", "🥺", "🤥", "🤫", "🤭", "🧐", "🤓", "😈", "👿", "👹", "👺", "💀", "👻", "👽", "🤖", "💩", "😺", "😸", "😹", "😻", "😼", "😽", "🙀", "😿", "😾"}
	emojiList := []string{"grinning", "grin", "joy", "rolling_on_the_floor_laughing", "smiley", "smile", "sweat_smile", "laughing", "wink", "blush", "yum", "sunglasses", "heart_eyes", "kissing_heart", "kissing", "kissing_smiling_eyes", "kissing_closed_eyes", "relaxed", "slightly_smiling_face", "hugging_face", "star-struck", "thinking_face", "face_with_raised_eyebrow", "neutral_face", "expressionless", "no_mouth", "face_with_rolling_eyes", "smirk", "persevere", "disappointed_relieved", "open_mouth", "zipper_mouth_face", "hushed", "sleepy", "tired_face", "sleeping", "relieved", "stuck_out_tongue", "stuck_out_tongue_winking_eye", "stuck_out_tongue_closed_eyes", "drooling_face", "unamused", "sweat", "pensive", "confused", "upside_down_face", "money_mouth_face", "astonished", "white_frowning_face", "slightly_frowning_face", "confounded", "disappointed", "worried", "triumph", "cry", "sob", "frowning", "anguished", "fearful", "weary", "exploding_head", "grimacing", "cold_sweat", "scream", "flushed", "zany_face", "dizzy_face", "rage", "angry", "face_with_symbols_on_mouth", "mask", "face_with_thermometer", "face_with_head_bandage", "nauseated_face", "face_vomiting", "sneezing_face", "innocent", "face_with_cowboy_hat", "clown_face", "lying_face", "shushing_face", "face_with_hand_over_mouth", "face_with_monocle", "nerd_face", "smiling_imp", "imp", "japanese_ogre", "japanese_goblin", "skull", "ghost", "alien", "robot_face", "hankey", "smiley_cat", "smile_cat", "joy_cat", "heart_eyes_cat", "smirk_cat", "kissing_cat", "scream_cat", "crying_cat_face", "pouting_cat"}
	// alf := "abcdefghijklmnopqrstuvwxyz"
	pollOptions := make(map[string]string)

	if len(commands) < 3 {
		slackClient.SendMessage(slackClient.NewOutgoingMessage("You need at least 3 arguments to create a poll!", slackChannel))
		return
	}

	newMessage := "Use the reaction next to the option you want to vote for! - "
	for i := range commands[2:] {
		pollOptions[emojiList[i]] = commands[i+2]
	}
	for i := range pollOptions {
		newMessage += fmt.Sprintf("|:%s: - %s| ", i, pollOptions[i])
	}
	fmt.Println(newMessage)
	_, timestamp, err := slackClient.PostMessage(slackChannel, slack.MsgOptionText(newMessage, false))

	if err != nil {
		fmt.Println(err)
	}

	msgRef := slack.NewRefToMessage(slackChannel, timestamp)
	for i := range pollOptions {
		err := slackClient.AddReaction(i, msgRef)
		if err != nil {
			fmt.Println("Failed to add reaction ", pollOptions[i])
		}
	}

	UserPoll := NewPoll(commands[1], timestamp, pollOptions)
	collection := client.C("polls")
	collection.Insert(UserPoll)

}

// sendHelp is a working help message, for reference.
func sendHelp(slackClient *slack.RTM, message, slackChannel string) {

	if strings.ToLower(message) != "help" {
		return
	}
	// var helpMessage Message
	// var _ = json.Unmarshal([]byte(helpStr), &helpMessage)
	attachment := slack.Attachment{
		Pretext:  "To view a list of commands, use the command `@pollbot commands`",
		Fallback: "We don't currently support your client",
		Color:    "#3AA3E3",
	}

	NewMessage := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := slackClient.PostMessage(slackChannel, slack.MsgOptionText("", false), NewMessage)
	if err != nil {
		fmt.Printf("Could not send message: %v", err)
	}
	fmt.Printf("Message with buttons sucessfully sent to channel %s at %s", channelID, timestamp)
	// slackClient.SendMessage(slackClient.NewOutgoingMessage(helpStr, slackChannel))
}

func sendCommands(slackClient *slack.RTM, message, slackChannel string) {

	message = strings.ToLower(message)
	commands := strings.Split(message, " ")
	if len(commands) == 1 {
		commandsStr := "The commands you can use are "
		for key, _ := range Commands {
			commandsStr += key
			commandsStr += ", "
		}

		slackClient.SendMessage(slackClient.NewOutgoingMessage(commandsStr, slackChannel))
	} else if len(commands) == 2 {
		if val, ok := Commands[commands[1]]; ok {
			slackClient.SendMessage(slackClient.NewOutgoingMessage(val.Description, slackChannel))
		} else {
			slackClient.SendMessage(slackClient.NewOutgoingMessage("I don't have that command!", slackChannel))
		}
	} else {
		slackClient.SendMessage(slackClient.NewOutgoingMessage("Too many arguments, I don't know what you want!", slackChannel))
	}
}

func ActionHandler(w http.ResponseWriter, r *http.Request) {
	var payload slack.InteractionCallback
	err := json.Unmarshal([]byte(r.FormValue("payload")), &payload)
	if err != nil {
		fmt.Printf("Could not parse action response JSON: %v", err)
	}
	fmt.Printf("Message button pressed by user %s with value %s", payload.User.Name, payload.Value)
}
