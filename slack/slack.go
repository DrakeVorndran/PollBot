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
			default:
				slackClient.SendMessage(slackClient.NewOutgoingMessage("I don't know what you want, try @PollBot commands", ev.Channel))
			}
			// ===============================================================
			// END SLACKBOT CUSTOM CODE
		default:

		}
	}
}

type Poll struct {
	MessageId string
	Name      string
	Items     map[string]string
}

func NewPoll(Name string, pollOptions map[string]string) Poll {
	return Poll{
		Name:      Name,
		MessageId: "",
		Items:     pollOptions,
	}
}

func sendCreate(slackClient *slack.RTM, message, slackChannel string, client *mgo.Database) {
	message = strings.ToLower(message)
	commands := strings.Split(message, " ")

	// headerText := slack.NewTextBlockObject("mrkdwn", "You have a new request:\n*<fakeLink.toEmployeeProfile.com|Fred Enriquez - New device request>*", false, false)
	// headerSection := slack.NewSectionBlock(headerText, nil, nil)
	// fmt.Println(headerSection)
	emojiString := "😀😁😂🤣😃😄😅😆😉😊😋😎😍😘🥰😗😙😚☺️🙂🤗🤩🤔🤨😐😑😶🙄😏😣😥😮🤐😯😪😫😴😌😛😜😝🤤😒😓😔😕🙃🤑😲☹️🙁😖😞😟😤😢😭😦😧😨😩🤯😬😰😱🥵🥶😳🤪😵😡😠🤬😷🤒🤕🤢🤮🤧😇🤠🤡🥳🥴🥺🤥🤫🤭🧐🤓😈👿👹👺💀👻👽🤖💩😺😸😹😻😼😽🙀😿😾"
	// alf := "abcdefghijklmnopqrstuvwxyz"
	pollOptions := make(map[string]string)

	if len(commands) < 3 {
		slackClient.SendMessage(slackClient.NewOutgoingMessage("You need at least 3 arguments to create a poll!", slackChannel))
		return
	}
	for i := range commands[2:] {
		fmt.Println(i)
		pollOptions[commands[i+2]] = string(emojiString[i])
	}
	fmt.Println(pollOptions)
	UserPoll := NewPoll(commands[1], pollOptions)
	collection := client.C("polls")
	UserJson, _ := json.Marshal(UserPoll)
	fmt.Println(string(UserJson))
	collection.Insert(UserJson)

	slackClient.SendMessage(slackClient.NewOutgoingMessage("", slackChannel))

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
