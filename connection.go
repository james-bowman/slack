package slack

import (
	"github.com/gorilla/websocket"
	"log"
	"fmt"
	"time"
	"encoding/json"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)


type Connection struct {
	// The websocket connection.
	ws *websocket.Conn
	
	sequence int

	// Buffered channel of outbound messages.
	Out chan []byte
	
	In chan []byte
	
	Config Config
}

// write writes a message with the given message type and payload.
func (c *Connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// socketWriter writes messages to the websocket connection.
func (c *Connection) socketWriter() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.Out:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				log.Println("Closing socket")
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				log.Println(err)
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

// socketReader reads messages from the websocket connection.
func (c *Connection) socketReader() {
	defer func() {
		c.ws.Close()
	}()
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		c.In <- message
	}
}

func (c *Connection) Send(response *Event) error {
	c.sequence++
	response.Id = c.sequence
	responseJson, err := json.Marshal(response)
	if err != nil {
		return err
	}
		
	c.Out <- responseJson
	
	return nil
}


type Event struct {
	Id int `json:"id"`
	ReplyTo int `json:"reply_to"`
	Type string `json:"type"`
	User string `json:"user"`
	Channel string `json:"channel"`
	Text string `json:"text"`
}

type messageProcessor func(Event) *Event

func (c *Connection) process(processMessage messageProcessor) {	
	for {
		msg := <-c.In
		
		var data Event
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
	
		response := processMessage(data)
		
		if response != nil {
			err := c.Send(response)
			
			if err != nil {
				log.Println(err)
				log.Println(response)
			}
		}
	}
}

func (c *Connection) Start(processMessage messageProcessor) {
	go c.process(processMessage)

	go c.socketWriter()
	c.socketReader()
}