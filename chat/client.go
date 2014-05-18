package chat

import (
	"github.com/gorilla/websocket"
	"io"
)

type Client struct {
	ws       *websocket.Conn
	Username string
	server   *Server
	writeCh  chan *message
	stopCh   chan bool
}

type message struct {
	Method string `json:"method"`
	Params interface{} `json:"params"`
}

func newClient(c *websocket.Conn, uname string, s *Server) *Client {
	return &Client{
		c,
		uname,
		s,
		make(chan *message),
		make(chan bool),
	}
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

		if msg.Method == "GetNicks" {
			c.replyNicks()
		} else if msg.Method == "Say" {
			c.say(msg.Params.(string))
		}
	}
	c.server.Del(c)
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

func (c *Client) replyNicks() {
	c.writeCh <- &message{"Nicks", c.server.GetNicks()}
}

func (c *Client) say(msg string) {
	c.server.Broadcast(c.Username, msg)
}

func (c *Client) NewMessage(who string, what string) {
	c.writeCh <- &message{"NewMessage", []string{who, what}}
}

func (c *Client) Joined(who string) {
	c.writeCh <- &message{"Joined", who}
}

func (c *Client) Left(who string) {
	c.writeCh <- &message{"Left", who}
}
