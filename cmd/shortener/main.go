package main

import (
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/server"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"github.com/volnistii11/URL-shortener/internal/app/storage/file"
	"github.com/volnistii11/URL-shortener/internal/telemetry"
	"log"
)

func main() {
	repo := storage.NewRepository()
	cfg := config.NewFlags()
	cfg.ParseFlags()
	fileStorage := file.NewSetter(repo, cfg)
	logger, err := telemetry.NewLogger()
	if err != nil {
		log.Printf("Error : %v\n", err)
	}
	defer logger.Sync()
	r := server.NewRouter(logger)
	s := r.Router(fileStorage.RestoreDataFromJSONFileToStructure(), cfg)
	s.Run(cfg.GetServer())
}
