package slack

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
	"encoding/json"
	"sync"
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

	wg sync.WaitGroup
	finish chan struct{}
	
	// Buffered channel of outbound messages.
	out chan []byte
	
	In chan []byte
	
	config Config
}

// write writes a message with the given message type and payload.
func (c *Connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// socketWriter writes messages to the websocket connection.
func (c *Connection) socketWriter() {
	c.wg.Add(1)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Println("Closing socket writer")
		ticker.Stop()
		c.ws.Close()
		c.wg.Done()
	}()
	for {
		select {
		case message, ok := <-c.out:
			if !ok {
				// channel closed so close the websocket
				c.write(websocket.CloseMessage, []byte{})
				log.Println("Closing socket")
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				// error writing to websocket
				log.Printf("Error writing to slack websocket: %s", err)
				return
			}
		case <-ticker.C:
			// if idle send a ping
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("Error sending ping on slack websocket: %s", err)
				return
			}
		case <-c.finish:
			return
		}
	}
}

// socketReader reads messages from the websocket connection.
func (c *Connection) socketReader() {
	c.wg.Add(1)
	defer func() {
		log.Println("Closing socket reader")
		c.ws.Close()
		c.wg.Done()
	}()
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading from slack websocket: %s",err)
			break
		}
		c.In <- message
	}
}

type event struct {
	Id int `json:"id"`
	Type string `json:"type"`
	Channel string `json:"channel"`
	Text string `json:"text"`
}

func (c *Connection) sendEvent(eventType string, channel string, text string) error {
	c.sequence++
	
	response := &event{Id: c.sequence, Type: eventType, Channel: channel, Text: text}

	responseJson, err := json.Marshal(response)
	if err != nil {
		return err
	}
		
	c.out <- responseJson
	
	return nil
}

func (c *Connection) SendMessage(channel string, text string) error {
	return c.sendEvent("message", channel, text)
}

func (c *Connection) start(reconnectionHandler func() (*Config, *websocket.Conn, error)) {
	for {
		c.finish = make(chan struct{})
		
		go c.socketWriter()
		c.socketReader()
		
		close(c.finish)
		c.wg.Wait()
		
		connected := false
		var ws *websocket.Conn
		var config *Config
		
		for i := 1; !connected; i = i * 2 {
			log.Printf("Lost connection to Slack - reconnecting in %d seconds", i)
		
			// wait 10 seconds before trying to reconnect
			time.Sleep(time.Duration(i)*time.Second)
		
			var err error
			config, ws, err = reconnectionHandler()
		
			if err != nil {
				log.Printf("Error reconnecting: %s", err)
			} else {
				log.Printf("Connected to Slack")
				connected = true
			}
		}
		
		c.ws = ws
		c.config = *config
	}
}