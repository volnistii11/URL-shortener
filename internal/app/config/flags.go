package config

import (
	"flag"
	"os"
)

type Flags interface {
	ParseFlags()
	GetServer() string
	GetRespondingServer() string
	GetFileStoragePath() string
	GetDatabaseDSN() string
}

func NewFlags() Flags {
	return &address{
		server:           "",
		respondingServer: "",
		fileStoragePath:  "",
		databaseDSN:      "",
	}
}

type address struct {
	server           string
	respondingServer string
	fileStoragePath  string
	databaseDSN      string
}

func (addr *address) GetServer() string {
	return addr.server
}

func (addr *address) GetRespondingServer() string {
	return addr.respondingServer
}

func (addr *address) GetFileStoragePath() string {
	return addr.fileStoragePath
}

func (addr *address) GetDatabaseDSN() string {
	return addr.databaseDSN
}

func (addr *address) ParseFlags() {
	flag.StringVar(&addr.server, "a", ":8080", "server address")
	flag.StringVar(&addr.respondingServer, "b", "", "responding server address")
	flag.StringVar(&addr.fileStoragePath, "f", "./tmp/short-url-db.json", "file storage path")
	flag.StringVar(&addr.databaseDSN, "d", "", "database dsn")
	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		addr.server = envServerAddress
	}
	if envRespondingServerAddress := os.Getenv("RESPONDING_SERVER_ADDRESS"); envRespondingServerAddress != "" {
		addr.respondingServer = envRespondingServerAddress
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		addr.fileStoragePath = envFileStoragePath
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		addr.databaseDSN = envDatabaseDSN
	}
}
