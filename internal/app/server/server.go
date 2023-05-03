package server

import (
	"github.com/volnistii11/URL-shortener/internal/app/handlers"
	"net/http"
)

func RunServer() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handlers.MainHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
