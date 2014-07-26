package main

import (
	"encoding/json"
	"log"
	"net"
)

type Client struct {
	Conn    net.Conn
	Update  chan bool
	decoder *json.Decoder
}

type Action struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

var clients []*Client

func AddClient(conn net.Conn) *Client {
	c := &Client{
		Conn:    conn,
		Update:  make(chan bool),
		decoder: json.NewDecoder(conn),
	}

	mutex.Lock()
	clients = append(clients, c)
	mutex.Unlock()

	go c.Listen()
	go c.update()
	return c
}

func RemoveClient(c *Client) {
	c.Conn.Close()
	idx := -1
	for i := 0; i < len(clients); i++ {
		obj := clients[i]
		if obj == c {
			idx = i
			break
		}
	}

	if idx < 0 {
		return
	}

	if idx != len(clients)-1 {
		clients[idx] = clients[len(clients)-1]
	}

	clients = clients[:len(clients)-1]

}

func (c *Client) Listen() {
	defer RemoveClient(c)

	for {
		var data map[string]Action

		if err := c.decoder.Decode(&data); err != nil {
			break
		}

		action, ok := data["action"]
		if ok {
			c.execute(action)
		}
	}
}

func (c *Client) update() {
	for {
		<-c.Update
		data, err := json.Marshal(pins)
		if err != nil {
			log.Fatal(err)
		}
		c.Conn.Write(data)
	}
}

func (c *Client) execute(action Action) {
	log.Println("Received action: ", action)
	mutex.Lock()

	switch action.Name {
	case "open":
		pin := GetPin(action.Id)
		if pin != nil {
			pin.Open()
		}
	case "close":
		pin := GetPin(action.Id)
		if pin != nil {
			pin.Close()
		}
	}

	for _, client := range clients {
		client.Update <- true
	}
	mutex.Unlock()
}
