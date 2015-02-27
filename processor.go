package slack

import (
	"log"
	"fmt"
	"encoding/json"
	"regexp"
)

type messageProcessor func(*Message)

func EventProcessor(con *Connection, respond messageProcessor, hear messageProcessor) {	
	for {
		msg := <-con.In
		
		log.Printf("%s", msg)
		
		var data map[string]interface{}
		err := json.Unmarshal(msg, &data)
	
		if err != nil {
			fmt.Printf("%T\n%s\n%#v\n", err, err, err)
			switch v := err.(type) {
				case *json.SyntaxError:
					log.Println(string(msg[v.Offset-40:v.Offset]))
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
			switch data["type"] {
				case "message":
					filterMessage(con, data, respond, hear)
			}
		}
	}
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

func filterMessage(con *Connection, data map[string]interface{}, respond messageProcessor, hear messageProcessor) {	
	var userFullName string
	var userId string
	
	user, ok := data["user"]	
	if ok {
		userId = user.(string)
		userFullName, _ = findUser(con.config, userId)
	}
	
	// process messages directed at Talbot
	r, _ := regexp.Compile("^(<@" + con.config.Self.Id + ">|@?" + con.config.Self.Name + "):? (.+)")
					
	matches := r.FindStringSubmatch(data["text"].(string))
				
	if len(matches) == 3 {
		if respond != nil {
			m := &Message{con: con, responseStrategy: reply, Text: matches[2], From: userFullName, fromId: userId, channel: data["channel"].(string)}
			respond(m)	
		}
	} else if data["channel"].(string)[0] == 'D' {
		if respond != nil {
			// process direct messages
			m := &Message{con: con, responseStrategy: send, Text: data["text"].(string), From: userFullName, fromId: userId, channel: data["channel"].(string)}
			respond(m)
		}
	} else {
		if hear != nil {
			m := &Message{con: con, responseStrategy: reply, Text: data["text"].(string), From: userFullName, fromId: userId, channel: data["channel"].(string)}
			hear(m)
		}
	}
}