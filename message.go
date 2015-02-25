package slack

import (

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
	m.Send(text)
}

func (m *Message) Respond(text string) {
	m.replier(m, text)
}


func reply(m *Message, text string) {
	m.Reply(text)
}

func send(m *Message, text string) {
	m.Send(text)
}