package main

import (
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/server"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	file_storage "github.com/volnistii11/URL-shortener/internal/app/storage/file-storage"
	"github.com/volnistii11/URL-shortener/internal/telemetry"
	"log"
)

func main() {
	repo := storage.NewRepository()
	cfg := config.NewFlags()
	cfg.ParseFlags()
	file_storage := file_storage.NewSetter(repo, cfg)
	logger, err := telemetry.NewLogger()
	if err != nil {
		log.Printf("Error : %v\n", err)
	}
	defer logger.Sync()
	r := server.NewRouter(logger)
	s := r.Router(file_storage.RestoreDataFromJSONFileToStructure(), cfg)
	s.Run(cfg.GetServer())
}
