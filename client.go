package main

import (
	"encoding/json"
	"log"
	"net"
	"time"
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
	defer c.Conn.Close()
	idx := -1
	mutex.Lock()
	defer mutex.Unlock()
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
			go c.execute(action)
		}
	}
	log.Print("Deconnecting a Client")
}

func (c *Client) update() {
	defer RemoveClient(c)
	Loop:
		for {
			<-c.Update
			ch := make(chan int, 1)
			data, err := json.Marshal(pins)
			if err != nil {
				log.Fatal(err)
			}

			select {
				case ch <- c.write(data):
				case <- time.After(5 * time.Second):
					log.Print("Couldn't write on client's socket")
					break Loop
				
			}
		}
}

func (c *Client) write(data []byte) int {
	size, err := c.Conn.Write(data)
	if err != nil {
		log.Print(err)
	}
	return size
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
