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

// writePump pumps messages from the hub to the websocket connection.
func (c *Connection) writePump() {
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

// readPump pumps messages from the websocket connection to the hub.
func (c *Connection) readPump() {
	defer func() {
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
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
	sequence := 0
	
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
		}
	
		response := processMessage(data)
		
		if response != nil {
			sequence++
			response.Id = sequence
			responseJson, err := json.Marshal(response)
			if err != nil {
				log.Println(err)
			}
			
			c.Out <- responseJson
		}
	}
}

func (c *Connection) Start(processMessage messageProcessor) {
	go c.process(processMessage)

	go c.writePump()
	c.readPump()
}