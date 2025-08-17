package main

import (
	"net/http"
)

func main() {
	const port = "8080"
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.Handle("/assets", http.FileServer(http.Dir("/assets/logo.png")))

	srv := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	srv.ListenAndServe()
}
