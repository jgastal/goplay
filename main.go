package main

import (
	"fmt"
	"net/http"
	"os"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello world!")
}

func main() {
	http.HandleFunc("/", hello)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
