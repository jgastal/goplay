package handlers

import (
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"net/http"
)

var sstore = sessions.NewCookieStore(
	[]byte("tDqZYv^\"?Qn2r|GgP!':rjY.naX!zLZBHSw8:2(pm`8G#?:utS!fBxd,9S-^\"D=D"),
	[]byte("_cB2t~ss,V/XIl^41ppWRYB6=PrJ\\\\U2"),
)

type RedirectAnonymousHandler struct {
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (h RedirectAnonymousHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := sstore.Get(r, "session")
	u, exists := session.Values["user"]
	if !exists {
		http.Redirect(w, r, "/login", 302)
		return
	}
	context.Set(r, "user", u)

	h.Handler(w, r)
}

type FormHandler struct {
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (h FormHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		InternalErrorHandler(w, r)
		return
	}
	h.Handler(w, r)
}

type ForbidAnonymousHandler struct {
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (h ForbidAnonymousHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := sstore.Get(r, "session")
	u, exists := session.Values["user"]
	if !exists {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	context.Set(r, "user", u)

	h.Handler(w, r)
}

func InternalErrorHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/html/internal-error.html")
}
