package chat

import (
	"github.com/gorilla/websocket"
	"io"
)

type Client struct {
	ws       *websocket.Conn
	Username string
	servers  map[string]*Server
	writeCh  chan *message
	stopCh   chan bool
}

type message struct {
	Server string `json:"server"`
	Method string `json:"method"`
	Params interface{} `json:"params"`
}

func NewClient(c *websocket.Conn, uname string) {
	client := &Client{
		c,
		uname,
		make(map[string]*Server),
		make(chan *message),
		make(chan bool),
	}
	client.Listen()
}

func (c *Client) Listen() {
	go c.write()
	for {
		msg := new(message)
		err := c.ws.ReadJSON(msg)
		if err == io.EOF {
			c.stopCh <- true
			break
		} else if err != nil {
			continue
		}

		s, ok := c.servers[msg.Server]
		if ok {
			if msg.Method == "GetNicks" {
				c.writeCh <- &message{s.name, "Nicks", s.GetNicks()}
			} else if msg.Method == "Say" {
				s.Broadcast(c.Username, msg.Params.(string))
			} else if msg.Method == "Leave" {
				delete(c.servers, s.name)
				s.delClient(c)
			}
		} else if msg.Method == "Join" {
			s, ok = servers[msg.Server]
			if ok {
				c.servers[s.name] = s
				s.addClient(c)
			}
		}
	}
	c.ws.Close()
}

func (c *Client) write() {
	for {
		select {
		case <-c.stopCh:
			return
		case msg := <-c.writeCh:
			err := c.ws.WriteJSON(msg)
			if err == io.EOF {
				//Read will get it's own EOF soon enough
				break
			}
		}
	}
}

func (c *Client) NewMessage(server, who, what string) {
	c.writeCh <- &message{server, "NewMessage", []string{who, what}}
}

func (c *Client) Joined(server, who string) {
	c.writeCh <- &message{server, "Joined", who}
}

func (c *Client) Left(server, who string) {
	c.writeCh <- &message{server, "Left", who}
}
