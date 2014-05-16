package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/gob"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"html/template"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"os"
)

var sstore = sessions.NewCookieStore(
	[]byte("tDqZYv^\"?Qn2r|GgP!':rjY.naX!zLZBHSw8:2(pm`8G#?:utS!fBxd,9S-^\"D=D"),
	[]byte("_cB2t~ss,V/XIl^41ppWRYB6=PrJ\\\\U2"),
)

var decoder = schema.NewDecoder()

var db *mgo.Database

type user struct {
	Email    string
	Password string
}

type login user

type signup struct {
	Email           string
	Password        string
	Password_repeat string
}

func internal_error(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/html/internal-error.html")
}

func templateResponse(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	t, err := template.ParseFiles(name)
	if err != nil {
		log.Println("Template error: ", err)
		internal_error(w, r)
		return
	}
	t.Execute(w, data)
}

func login_get(w http.ResponseWriter, r *http.Request) {
	// FIXME redirect if already logged in

	templateResponse(w, r, "template/login.html", nil)
}

func login_post(w http.ResponseWriter, r *http.Request) {
	// FIXME redirect if already logged in

	cred := new(login)
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
		internal_error(w, r)
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
		internal_error(w, r)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(cred.Password)) != nil {
		formErr := map[string]string{"login_form_error": "Incorrect email or password."}
		templateResponse(w, r, "template/login.html", formErr)
		return
	}

	// Putting the password on a cookie just opens up an attack vector, even if the password is hashed and the cookie encrypted
	u.Password = ""
	session, _ := sstore.Get(r, "session")
	session.Values["user"] = u
	err = session.Save(r, w)
	if err != nil {
		log.Println("Session error: ", err)
		internal_error(w, r)
		return
	}

	http.Redirect(w, r, "/profile", 302)
}

func signup_post(w http.ResponseWriter, r *http.Request) {
	// FIXME redirect if already logged in

	cred := new(signup)
	err := decoder.Decode(cred, r.PostForm)
	if err != nil {
		formErr := map[string]string{"signup_form_error": "Extra parameters received. Please send only email and password."}
		templateResponse(w, r, "template/login.html", formErr)
		return
	}
	if cred.Password != cred.Password_repeat {
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
		internal_error(w, r)
		return
	}
	col := db.C("users")
	err = col.Insert(u)
	if err != nil {
		log.Println("Mongodb error: ", err)
		internal_error(w, r)
		return
	}

	session, _ := sstore.Get(r, "session")
	session.Values["user"] = u
	err = session.Save(r, w)
	if err != nil {
		log.Println("Session error: ", err)
		http.Redirect(w, r, "/login", 302)
		return
	}

	http.Redirect(w, r, "/profile", 302)
}

func profile(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "PLACEHOLDER")
}

func main() {
	router := mux.NewRouter()

	router.Methods("GET").Path("/login").HandlerFunc(login_get)
	router.Methods("GET").Path("/profile").HandlerFunc(profile)

	//Form handlers
	form_router := router.Methods("POST").Subrouter()
	form_router.Handle("/login", formHandler{login_post})
	form_router.Handle("/signup", formHandler{signup_post})

	gob.Register(&user{})

	s, err := mgo.Dial(os.Getenv("MONGOHQ_URL"))
	if err != nil {
		panic(err)
	}
	db = s.DB("")

	db.C("users").EnsureIndex(mgo.Index{Key: []string{"email"}, Unique: true})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		panic(err)
	}
}
