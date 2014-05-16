package main

import (
	"github.com/gorilla/context"
	"net/http"
)

type forbidAnonymousHandler struct {
	handler func(w http.ResponseWriter, r *http.Request)
}

func (h forbidAnonymousHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := sstore.Get(r, "session")
	u, exists := session.Values["user"]
	if !exists {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	context.Set(r, "user", u)

	h.handler(w, r)
}
