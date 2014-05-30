package chat

import (
	"log"
)

var servers = make(map[string]*Server)

type Server struct {
	clients map[string]*Client
	addCh   chan *Client
	delCh   chan *Client
	name    string
}

func NewServer(name string) {
	s := Server{}
	s.addCh = make(chan *Client)
	s.delCh = make(chan *Client)
	s.clients = make(map[string]*Client)
	s.name = name

	servers[name] = &s
	go s.run()
}

func (s *Server) run() {
listen:
	for {
		select {
		case c := <-s.addCh:
			log.Println(c.Username, "joined", s.name)
			for _, v := range s.clients {
				v.joined(s.name, c.Username)
			}
			s.clients[c.Username] = c

		case c := <-s.delCh:
			log.Println(c.Username, "left", s.name)
			delete(s.clients, c.Username)
			for _, v := range s.clients {
				v.left(s.name, c.Username)
			}
			if len(s.clients) == 0 {
				break listen
			}
		}
	}
	delete(servers, s.name)
}

func (s *Server) addClient(c *Client) {
	s.addCh <- c
}

func (s *Server) delClient(c *Client) {
	s.delCh <- c
}

func (s *Server) getNicks() []string {
	reply := make([]string, len(s.clients))
	i := 0
	for k := range s.clients {
		reply[i] = k
		i++
	}
	return reply
}

func (s *Server) broadcast(who string, what string) {
	for _, v := range s.clients {
		v.newMessage(s.name, who, what)
	}
}
