package slack

type eventer interface {
	Write(string, string) error
}

type Message struct {
	eventStream eventer

	responseStrategy func(*Message, string) error

	Text string
	From string

	fromId  string
	channel string
}

//type responder func(string)

func (m *Message) Tell(channel string, text string) error {
	return m.eventStream.Write(channel, text)
}

func (m *Message) Send(text string) error {
	return m.eventStream.Write(m.channel, text)
}

func (m *Message) Reply(text string) error {
	return m.Send("<@" + m.fromId + ">: " + text)
}

func (m *Message) Respond(text string) error {
	return m.responseStrategy(m, text)
}

func reply(m *Message, text string) error {
	return m.Reply(text)
}

func send(m *Message, text string) error {
	return m.Send(text)
}
