package main

import (
	"github.com/volnistii11/URL-shortener/internal/app/server"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
)

func main() {

	repo := storage.NewRepository()

	server.RunServer(repo)
}
