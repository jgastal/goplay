package chat

import (
	"github.com/gorilla/context"
	"github.com/gorilla/websocket"
	"github.com/jgastal/goplay/handlers"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{}

type Server struct {
	handler func(http.ResponseWriter, *http.Request)
	clients map[string]*Client
	addCh   chan *Client
	delCh   chan *Client
	name    string
}

func NewServer(name string) *Server {
	s := Server{}
	s.addCh = make(chan *Client)
	s.delCh = make(chan *Client)
	s.clients = make(map[string]*Client)
	s.name = name

	go s.run()
	return &s
}

func (s *Server) GetHandler() func(http.ResponseWriter, *http.Request) {
	if s.handler == nil {
		s.handler = func(w http.ResponseWriter, r *http.Request) {
			u := context.Get(r, "username").(string)
			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				handlers.InternalErrorHandler(w, r)
				return
			}

			client := newClient(ws, u, s)
			s.addCh <- client
			client.Listen()
		}
	}
	return s.handler
}

func (s *Server) run() {
	for {
		select {
		case c := <-s.addCh:
			log.Println(c.Username, "connected")
			for _, v := range s.clients {
				v.Joined(c.Username)
			}
			s.clients[c.Username] = c

		case c := <-s.delCh:
			log.Println(c.Username, "disconnected")
			delete(s.clients, c.Username)
			for _, v := range s.clients {
				v.Left(c.Username)
			}
		}
	}
}

func (s *Server) Del(c *Client) {
	s.delCh <- c
}

func (s *Server) GetNicks() []string {
	reply := make([]string, len(s.clients))
	i := 0
	for k := range s.clients {
		reply[i] = k
		i++
	}
	return reply
}

func (s *Server) Broadcast(who string, what string) {
	for _, v := range s.clients {
		v.NewMessage(who, what)
	}
}
