package chat

import (
	"io"
)

type JSONReadWriteCloser interface {
	io.Closer
	ReadJSON(interface{}) error
	WriteJSON(interface{}) error
}

type Client struct {
	ws       JSONReadWriteCloser
	Username string
	servers  map[string]*Server
	writeCh  chan *message
	stopCh   chan struct{}
}

type message struct {
	Server string `json:"server"`
	Method string `json:"method"`
	Params interface{} `json:"params"`
}

func NewClient(c JSONReadWriteCloser, uname string) {
	client := &Client{
		c,
		uname,
		make(map[string]*Server),
		make(chan *message),
		make(chan struct{}),
	}
	client.Listen()
}

func (c *Client) Listen() {
	go c.write()
	for {
		msg := new(message)
		err := c.ws.ReadJSON(msg)
		if err == io.EOF {
			c.stopCh <- struct{}{}
			break
		} else if err != nil {
			continue
		}

		s, ok := c.servers[msg.Server]
		if ok {
			switch msg.Method {
				case "GetNicks":
					c.writeCh <- &message{s.name, "Nicks", s.GetNicks()}
				case "Say":
					s.Broadcast(c.Username, msg.Params.(string))
				case "Leave":
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
	for _, s := range c.servers {
		s.delClient(c)
	}
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
