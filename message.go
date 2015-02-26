package slack

import (
	"log"
)

type Message struct {
	con *Connection
	
	replier func(*Message, string)
	
	Text string
	From string
	
	fromId string
	channel string
}

type responder func(string)

func (m *Message) Send(text string) {
	m.con.SendMessage(m.channel, text)
}

func (m *Message) Reply(text string) {
	m.Send("<@" + m.fromId + ">: " + text)
}

func (m *Message) Respond(text string) {
	m.replier(m, text)
}


func reply(m *Message, text string) {
	log.Println("Replying")
	m.Reply(text)
}

func send(m *Message, text string) {
	log.Println("Sending")
	m.Send(text)
}