package slack

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
)

const (
	slackEventTypeMessage = "message"

	maxMessageSize = 4000
)

// Processor type processes inbound events from Slack
type Processor struct {
	// Connection to Slack
	con *Connection

	// Slack user information relating to the bot account
	self User

	// a sequence number to uniquely identify sent messages and correlate with acks from Slack
	sequence int

	// map of event handler functions to handle types of Slack event
	eventHandlers map[string]func(*Processor, map[string]interface{})
}

// event type represents an event sent to Slack e.g. messages
type event struct {
	Id      int    `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// send Event to Slack
func (p *Processor) sendEvent(eventType string, channel string, text string) error {
	p.sequence++

	response := &event{Id: p.sequence, Type: eventType, Channel: channel, Text: text}

	responseJson, err := json.Marshal(response)
	if err != nil {
		return err
	}

	p.con.Write(responseJson)

	return nil
}

// Write the message on the specified channel to Slack
func (p *Processor) Write(channel string, text string) error {
	if len(text) <= maxMessageSize {
		return p.sendEvent(slackEventTypeMessage, channel, text)
	}

	for len(text) > 0 {
		if len(text) <= maxMessageSize {
			if err := p.sendEvent(slackEventTypeMessage, channel, text); err != nil {
				return err
			}
			text = ""
		} else {
			// split message at a convenient place
			maxSizeChunk := text[:maxMessageSize]

			var breakIndex int
			if lastLineBreak := strings.LastIndex(maxSizeChunk, "\n"); lastLineBreak > -1 {
				breakIndex = lastLineBreak
			} else if lastWordBreak := strings.LastIndexAny(maxSizeChunk, "\n\t .,/\\-(){}[]|=+*&"); lastWordBreak > -1 {
				breakIndex = lastWordBreak
			} else {
				breakIndex = maxMessageSize
			}

			if err := p.sendEvent(slackEventTypeMessage, channel, text[:breakIndex]); err != nil {
				return err
			}

			if breakIndex != maxMessageSize {
				breakIndex++
			}

			text = text[breakIndex:]
		}
	}

	return nil
}

// Start processing events from Slack
func (p *Processor) Start() {
	for {
		msg := p.con.Read()

		log.Printf("%s", msg)

		var data map[string]interface{}
		err := json.Unmarshal(msg, &data)

		if err != nil {
			fmt.Printf("%T\n%s\n%#v\n", err, err, err)
			switch v := err.(type) {
			case *json.SyntaxError:
				log.Println(string(msg[v.Offset-40 : v.Offset]))
			}
			log.Printf("%s", msg)
			continue
		}

		// if reply_to attribute is present the event is an ack' for a sent message
		_, isReply := data["reply_to"]
		subtype, ok := data["subtype"]
		var isMessageChangedEvent bool

		if ok {
			isMessageChangedEvent = (subtype.(string) == "message_changed" || subtype.(string) == "message_deleted")
		}

		if !isReply && !isMessageChangedEvent {
			handler, ok := p.eventHandlers[data["type"].(string)]

			if ok {
				handler(p, data)
			}
		}
	}
}

// type for callbacks to receive messages from Slack
type messageProcessor func(*Message)

// Starts processing events on the connection from Slack and passes any messages to the hear callback and only
// messages addressed to the bot to the respond callback
func EventProcessor(con *Connection, respond messageProcessor, hear messageProcessor) {
	p := Processor{
		con:  con,
		self: con.config.Self,
		eventHandlers: map[string]func(*Processor, map[string]interface{}){
			slackEventTypeMessage: func(p *Processor, event map[string]interface{}) {
				filterMessage(p, event, respond, hear)
			},
		},
	}

	p.Start()
}

// finds the full name of the Slack user for the specified user ID
func findUser(config Config, user string) (string, bool) {
	var users []User

	users = config.Users

	for i := 0; i < len(users); i++ {
		if users[i].Id == user {
			return users[i].RealName, true
		}
	}

	return "", false
}

// Invoke one of the specified callbacks for the message if appropriate
func filterMessage(p *Processor, data map[string]interface{}, respond messageProcessor, hear messageProcessor) {
	var userFullName string
	var userId string

	user, ok := data["user"]
	if ok {
		userId = user.(string)
		userFullName, _ = findUser(p.con.config, userId)
	}

	// process messages directed at Talbot
	r, _ := regexp.Compile("^(<@" + p.self.Id + ">|@?" + p.self.Name + "):? (.+)")

	text, ok := data["text"]
	if !ok {
		return
	}

	matches := r.FindStringSubmatch(text.(string))

	if len(matches) == 3 {
		if respond != nil {
			m := &Message{eventStream: p, responseStrategy: reply, Text: matches[2], From: userFullName, fromId: userId, channel: data["channel"].(string)}
			respond(m)
		}
	} else if data["channel"].(string)[0] == 'D' {
		if respond != nil {
			// process direct messages
			m := &Message{eventStream: p, responseStrategy: send, Text: text.(string), From: userFullName, fromId: userId, channel: data["channel"].(string)}
			respond(m)
		}
	} else {
		if hear != nil {
			m := &Message{eventStream: p, responseStrategy: send, Text: text.(string), From: userFullName, fromId: userId, channel: data["channel"].(string)}
			hear(m)
		}
	}
}
