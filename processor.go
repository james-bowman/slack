package slack

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
)

type eventReadWriter interface {
	Write([]byte)
	Read() []byte
}

type event struct {
	Id      int    `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type Processor struct {
	con *Connection

	self User

	sequence int

	eventHandlers map[string]func(*Processor, map[string]interface{})
}

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

func (p *Processor) Write(channel string, text string) error {
	return p.sendEvent("message", channel, text)
}

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
			handler, ok := p.eventHandlers["type"]

			if ok {
				handler(p, data)
			}
			/*
				switch data["type"] {
				case "message":
					filterMessage(con, data, respond, hear)
				}
			*/
		}
	}
}

type messageProcessor func(*Message)

func EventProcessor(con *Connection, respond messageProcessor, hear messageProcessor) {
	p := Processor{
		con:  con,
		self: con.config.Self,
		eventHandlers: map[string]func(*Processor, map[string]interface{}){
			"message": func(p *Processor, event map[string]interface{}) {
				filterMessage(p, event, respond, hear)
			},
		},
	}

	p.Start()
}

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
