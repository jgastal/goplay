package main

import (
	"net/http"
)

type formHandler struct {
	handler func(w http.ResponseWriter, r *http.Request)
}

func (h formHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		internal_error(w, r)
		return
	}
	h.handler(w, r)
}
