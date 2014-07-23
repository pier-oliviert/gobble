package main

import (
  "net"
  "encoding/json"
  "sync"
  )

type Client struct {
  Conn net.Conn
  Notify chan string
  decoder *json.Decoder
}

type Action struct {
  Name string `json:"name"`
  Id int `json:"id"`
} 

var clients []*Client
var mutex sync.RWMutex

func AddClient(conn net.Conn) (*Client) {
  c := &Client{
    Conn: conn,
    Notify: make(chan string),
    decoder: json.NewDecoder(conn),
  }

  mutex.Lock()
  clients = append(clients, c)
  mutex.Unlock()

  go c.Listen()
  return c
}

func RemoveClient(c *Client) {
  mutex.Lock()
  defer mutex.Unlock()
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

  if idx != len(clients) - 1 {
    clients[idx] = clients[len(clients) - 1]
  }

  clients = clients[:len(clients) -1]

}

func (c *Client) Listen() {
  defer RemoveClient(c)

  for {
    var data map[string]Action

    if err := c.decoder.Decode(&data); err != nil {
      continue
    }

    action, ok := data["action"]
    if ok {
      c.execute(action)
    }
  }
}

func (c *Client) execute(action Action) {
  switch action.Name {
    case "open": OpenGPIO(int64(action.Id))
    case "close": CloseGPIO(int64(action.Id))
    default: ListGPIO()
  }
}
