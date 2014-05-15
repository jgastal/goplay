package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

func login_get(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/html/login.html")
}

func login_post(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "PLACEHOLDER")
}

func signup_post(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "PLACEHOLDER")
}

func main() {
	router := mux.NewRouter()

	router.Methods("GET").Path("/login").HandlerFunc(login_get)

	//Form handlers
	form_router := router.Methods("POST").Subrouter()
	form_router.HandleFunc("/login", login_post)
	form_router.HandleFunc("/signup", signup_post)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		panic(err)
	}
}
