package slack

// eventer interface represents a pipe along which messages can be sent
type eventer interface {
	// Write the message to the specified channel
	Write(string, string) error
}

// Message type represents a message received from Slack
type Message struct {
	// a pipe for sending responses to this message
	eventStream eventer

	// the strategy for sending the response - depending upon how the message was received e.g. a reply if
	// addressed specifically to the bot or a send if not
	responseStrategy func(*Message, string) error

	// the text content of the message
	Text string

	// the name of the user whom the message is from
	From string

	// id of the user the message is from
	fromId string

	// channel on which the message was received
	channel string
}

// Send a new message on the specified channel
func (m *Message) Tell(channel string, text string) error {
	return m.eventStream.Write(channel, text)
}

// Send a new message on the channel this message was received on
func (m *Message) Send(text string) error {
	return m.eventStream.Write(m.channel, text)
}

// Send a reply to the user who sent this message on the same channel it was received on
func (m *Message) Reply(text string) error {
	return m.Send("<@" + m.fromId + ">: " + text)
}

// Send a message in a way that matches the way in which this message was received e.g.
// if this message was addressed then send a reply back to person who sent the message.
func (m *Message) Respond(text string) error {
	return m.responseStrategy(m, text)
}

// response strategy for replying
func reply(m *Message, text string) error {
	return m.Reply(text)
}

// response strategy for sending
func send(m *Message, text string) error {
	return m.Send(text)
}
