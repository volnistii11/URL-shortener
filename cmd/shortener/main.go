package main

import (
	"log"

	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/server"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"github.com/volnistii11/URL-shortener/internal/app/storage/database"
	"github.com/volnistii11/URL-shortener/internal/app/storage/file"
	"github.com/volnistii11/URL-shortener/internal/telemetry"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.NewFlags()
	cfg.ParseFlags()

	//example: "postgres://pguser:pgpwd4habr@localhost:5432/shortenerdb"
	db, err := database.NewConnection("pgx", cfg.GetDatabaseDSN())
	if err != nil {
		log.Printf("Error : %v\n", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)

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
