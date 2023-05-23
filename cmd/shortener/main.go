package main

import (
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/server"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
)

func main() {
	repo := storage.NewRepository()
	cfg := config.NewFlags()
	cfg.ParseFlags()

	srv := server.NewRouter()
	s := srv.Router(repo, cfg)
	s.Run(cfg.GetServer())
}
