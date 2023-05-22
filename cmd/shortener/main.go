package main

import (
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/server"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
)

func main() {
	repo := storage.NewRepository()
	cfg := config.NewFlags()

	srv := server.NewRouter()
	srv.RunServer(repo, cfg)
}
