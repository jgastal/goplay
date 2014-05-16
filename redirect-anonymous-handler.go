package main

import (
	"github.com/gorilla/context"
	"net/http"
)

type redirectAnonymousHandler struct {
	handler func(w http.ResponseWriter, r *http.Request)
}

func (h redirectAnonymousHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := sstore.Get(r, "session")
	u, exists := session.Values["user"]
	if !exists {
		http.Redirect(w, r, "/login", 302)
		return
	}
	context.Set(r, "user", u)

	h.handler(w, r)
}
