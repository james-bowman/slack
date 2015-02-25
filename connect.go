package slack

import (
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"io/ioutil"
	"log"
	"encoding/json"
)

func handshake(apiUrl string, token string) (*Config, error) {
	resp, err := http.PostForm(apiUrl, url.Values{"token":{token}})

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}
	
	var data Config
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("%T\n%s\n%#v\n", err, err, err)
		switch v := err.(type) {
			case *json.SyntaxError:
				log.Println(string(body[v.Offset-40:v.Offset]))
		}
		log.Printf("%s", body)
		return nil, err
	}

	return &data, nil
}

func connectAndUpgrade(url string, token string) (*Config, *websocket.Conn, error) {
	config, err := handshake(url, token)
	
	if err != nil {
		return nil, nil, err
	}
	
	conn, _, err := websocket.DefaultDialer.Dial(config.Url, http.Header{})
	
	if err != nil {
		return nil, nil, err
	}
	
	return config, conn, nil
}

func Connect(token string) (*Connection, error) {
	apiStartUrl := "https://slack.com/api/rtm.start"
	
	config, conn, err := connectAndUpgrade(apiStartUrl, token)
		
	if err != nil {
		return nil, err
	}
	
	c := Connection{ws: conn, out: make(chan []byte, 256), In: make(chan []byte, 256), config: *config}
	
	go c.start(apiStartUrl, token) 
	
	return &c, nil
}