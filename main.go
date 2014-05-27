package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/gob"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/jgastal/goplay/chat"
	"html/template"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"os"
)

var sstore sessions.Store

var decoder = schema.NewDecoder()

var db *mgo.Database

var upgrader = websocket.Upgrader{}

type user struct {
	Email    string
	Password string
}

type signup struct {
	Email           string
	Password        string
	PasswordRepeat  string
}

func templateResponse(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	t, err := template.ParseFiles(name)
	if err != nil {
		log.Println("Template error: ", err)
		InternalErrorHandler(w, r)
		return
	}
	t.Execute(w, data)
}

func loginGet(w http.ResponseWriter, r *http.Request) {
	session, _ := sstore.Get(r, "session")
	_, exists := session.Values["username"]
	if exists {
		http.Redirect(w, r, "/profile", 302)
		return
	}

	templateResponse(w, r, "template/login.html", nil)
}

func loginPost(w http.ResponseWriter, r *http.Request) {
	session, _ := sstore.Get(r, "session")
	_, exists := session.Values["username"]
	if exists {
		http.Redirect(w, r, "/profile", 302)
		return
	}

	cred := new(user)
	err := decoder.Decode(cred, r.PostForm)
	if err != nil {
		formErr := map[string]string{"login_form_error": "Extra parameters received. Please send only email and password."}
		templateResponse(w, r, "template/login.html", formErr)
		return
	}

	col := db.C("users")
	query := col.Find(map[string]string{"email": cred.Email})
	found, err := query.Count()
	if err != nil {
		log.Println("Mongodb error: ", err)
		InternalErrorHandler(w, r)
		return
	}
	if found <= 0 {
		formErr := map[string]string{"login_form_error": "Incorrect email or password."}
		templateResponse(w, r, "template/login.html", formErr)
		return
	}

	u := new(user)
	err = query.One(&u)
	if err != nil {
		log.Println("Mongodb error: ", err)
		InternalErrorHandler(w, r)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(cred.Password)) != nil {
		formErr := map[string]string{"login_form_error": "Incorrect email or password."}
		templateResponse(w, r, "template/login.html", formErr)
		return
	}

	session.Values["username"] = u.Email
	err = session.Save(r, w)
	if err != nil {
		log.Println("Session error: ", err)
		InternalErrorHandler(w, r)
		return
	}

	http.Redirect(w, r, "/profile", 302)
}

func signupPost(w http.ResponseWriter, r *http.Request) {
	session, _ := sstore.Get(r, "session")
	_, exists := session.Values["username"]
	if exists {
		http.Redirect(w, r, "/profile", 302)
		return
	}

	cred := new(signup)
	err := decoder.Decode(cred, r.PostForm)
	if err != nil {
		formErr := map[string]string{"signup_form_error": "Extra parameters received. Please send only email and password."}
		templateResponse(w, r, "template/login.html", formErr)
		return
	}
	if cred.Password != cred.PasswordRepeat {
		formErr := map[string]string{"signup_form_error": "Passwords don't match."}
		templateResponse(w, r, "template/login.html", formErr)
		return
	}
	// FIXME validate email and password

	u := new(user)
	u.Email = cred.Email
	pwd, err := bcrypt.GenerateFromPassword([]byte(cred.Password), bcrypt.DefaultCost)
	u.Password = string(pwd)
	if err != nil {
		log.Println("bcrypt error: ", err)
		InternalErrorHandler(w, r)
		return
	}
	col := db.C("users")
	err = col.Insert(u)
	if err != nil {
		log.Println("Mongodb error: ", err)
		InternalErrorHandler(w, r)
		return
	}

	session.Values["username"] = u.Email
	err = session.Save(r, w)
	if err != nil {
		log.Println("Session error: ", err)
		http.Redirect(w, r, "/login", 302)
		return
	}

	http.Redirect(w, r, "/profile", 302)
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	u := context.Get(r, "username").(string)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		InternalErrorHandler(w, r)
		return
	}

	chat.NewClient(ws, u)
}

func profile(w http.ResponseWriter, r *http.Request) {
	u := context.Get(r, "username")

	var endpoint string
	if proto, ok := r.Header["X-Forwarded-Proto"]; ok && proto[0] == "https" {
		endpoint = "wss://"
	} else {
		endpoint = "ws://"
	}
	endpoint += r.Host + "/chat"
	ctx := map[string]interface{}{
		"username":      u,
		"chat_endpoint": endpoint,
	}
	templateResponse(w, r, "template/chat.html", ctx)
}

func setupDb() {
	s, err := mgo.Dial(os.Getenv("MONGOHQ_URL"))
	if err != nil {
		panic(err)
	}
	db = s.DB("")

	db.C("users").EnsureIndex(mgo.Index{Key: []string{"email"}, Unique: true})
}

func setupHandlers() (router *mux.Router) {
	router = mux.NewRouter()

	router.Methods("GET").Path("/login").HandlerFunc(loginGet)
	router.Methods("GET").Path("/profile").Handler(RedirectAnonymousHandler{profile})

	//Form handlers
	formRouter := router.Methods("POST").Subrouter()
	formRouter.Handle("/login", FormHandler{loginPost})
	formRouter.Handle("/signup", FormHandler{signupPost})

	chat.NewServer("Lobby")
	//Chat websocket handler
	router.Path("/chat").Handler(ForbidAnonymousHandler{chatHandler})
	return
}

func main() {
	cryptoKey := os.Getenv("COOKIESTORE_CRYPTO_KEY")
	if cryptoKey == "" {
		sstore = sessions.NewCookieStore([]byte(os.Getenv("COOKIESTORE_AUTH_KEY")))
	} else {
		sstore = sessions.NewCookieStore([]byte(os.Getenv("COOKIESTORE_AUTH_KEY")), []byte(cryptoKey))
	}

	gob.Register(&user{})

	setupDb()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	err := http.ListenAndServe(":"+port, setupHandlers())
	if err != nil {
		panic(err)
	}
}
