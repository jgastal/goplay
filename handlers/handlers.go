package handlers

import (
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"net/http"
	"os"
)

var sstore sessions.Store

type RedirectAnonymousHandler struct {
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (h RedirectAnonymousHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := sstore.Get(r, "session")
	u, exists := session.Values["username"]
	if !exists {
		http.Redirect(w, r, "/login", 302)
		return
	}
	context.Set(r, "username", u)

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
	u, exists := session.Values["username"]
	if !exists {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	context.Set(r, "username", u)

	h.Handler(w, r)
}

func InternalErrorHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/html/internal-error.html")
}

func init() {
	crypto_key := os.Getenv("COOKIESTORE_CRYPTO_KEY")
	if crypto_key == "" {
		sstore = sessions.NewCookieStore([]byte(os.Getenv("COOKIESTORE_AUTH_KEY")))
	} else {
		sstore= sessions.NewCookieStore([]byte(os.Getenv("COOKIESTORE_AUTH_KEY")), []byte(crypto_key))
	}
}
