package slack


type Message struct {
	con *Connection
	
	responseStrategy func(*Message, string)
	
	Text string
	From string
	
	fromId string
	channel string
}

type responder func(string)

func (m *Message) Tell(channel string, text string) {
	m.con.SendMessage(channel, text)
}

func (m *Message) Send(text string) {
	m.con.SendMessage(m.channel, text)
}

func (m *Message) Reply(text string) {
	m.Send("<@" + m.fromId + ">: " + text)
}

func (m *Message) Respond(text string) {
	m.responseStrategy(m, text)
}


func reply(m *Message, text string) {
	m.Reply(text)
}

func send(m *Message, text string) {
	m.Send(text)
}