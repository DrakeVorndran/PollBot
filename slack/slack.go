package slack

import ( //packages

	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/nlopes/slack"
)

/*
   TODO: Change @BOT_NAME to the same thing you entered when creating your Slack application.
   NOTE: command_arg_1 and command_arg_2 represent optional parameteras that you define
   in the Slack API UI
*/

// type Message struct { //Type
// 	Blocks []struct {
// 		Type string `json:"type"`
// 		Text struct {
// 			Type string `json:"type"`
// 			Text string `json:"text"`
// 		} `json:"text"`
// 		Accessory struct {
// 			Type string `json:"type"`
// 			Text struct {
// 				Type string `json:"type"`
// 				Text string `json:"text"`
// 			} `json:"text"`
// 			Value string `json:"value"`
// 		} `json:"accessory"`
// 	} `json:"blocks"`
// }

// {
// 	"blocks": [
// 		{
// 			"type": "section",
// 			"text": {
// 				"type": "mrkdwn",
// 				"text": "You can add a button alongside text in your message. "
// 			},
// 			"accessory": {
// 				"type": "button",
// 				"text": {
// 					"type": "plain_text",
// 					"text": "Button"
// 				},
// 				"value": "click_me_123"
// 			}
// 		}
// 	]
// }

// const helpStr = `{
// 	"blocks": [
// 		{
// 			"type": "section",
// 			"text": {
// 				"type": "mrkdwn",
// 				"text": "You can add a button alongside text in your message. "
// 			},
// 			"accessory": {
// 				"type": "button",
// 				"text": {
// 					"type": "plain_text",
// 					"text": "Button"
// 				},
// 				"value": "click_me_123"
// 			}
// 		}
// 	]
// }`

const helpStr = "To find all commands, use command @PollBot commands" //constant

const commandMessage = ""

const PublicConst = "This is a public const" //public const
var PublicVar = "This is a public const"     //public var

// Map of every command, with a description of what they do
type CommandArg interface {
	run()
}

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
			default:
				slackClient.SendMessage(slackClient.NewOutgoingMessage("I don't know what you want, try @PollBot commands", ev.Channel))
			}
			// ===============================================================
			// END SLACKBOT CUSTOM CODE
		default:

		}
	}
}

// sendHelp is a working help message, for reference.
func sendHelp(slackClient *slack.RTM, message, slackChannel string) {

	if strings.ToLower(message) != "help" {
		return
	}
	// var helpMessage Message
	// var _ = json.Unmarshal([]byte(helpStr), &helpMessage)
	attachment := slack.Attachment{
		Pretext:  "To view a list of commands, use the command `@pollbot commands` or press the commands button",
		Fallback: "We don't currently support your client",
		Color:    "#3AA3E3",
		Actions: []slack.AttachmentAction{
			slack.AttachmentAction{
				Name:  "commands",
				Text:  "Commands",
				Type:  "button",
				Value: "commands",
			},
		},
	}

	NewMessage := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := slackClient.PostMessage(slackChannel, slack.MsgOptionText("", false), NewMessage)
	if err != nil {
		fmt.Printf("Could not send message: %v", err)
	}
	fmt.Printf("Message with buttons sucessfully sent to channel %s at %s", channelID, timestamp)
	// slackClient.SendMessage(slackClient.NewOutgoingMessage(helpStr, slackChannel))
}

// sendResponse is NOT unimplemented --- write code in the function body to complete!

// func sendResponse(slackClient *slack.RTM, message, slackChannel string) {
// 	command := strings.ToLower(message)

// }
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
