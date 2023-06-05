package main

import (
	"github.com/gin-gonic/gin"
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
	logger, err := telemetry.NewLogger()
	if err != nil {
		log.Printf("Error : %v\n", err)
	}
	defer logger.Sync()

	r := server.NewRouter(logger)
	var s *gin.Engine
	if cfg.GetFileStoragePath() != "" {
		fileStorage := file.NewRestorer(repo, cfg)
		s = r.Router(fileStorage.RestoreDataFromJSONFileToStructure(), cfg)
	} else {
		s = r.Router(repo, cfg)
	}
	s.Run(cfg.GetServer())
}
