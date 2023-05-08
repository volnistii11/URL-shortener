package config

import (
	"flag"
	"os"
)

var Addresses struct {
	Server           string
	RespondingServer string
}

func ParseFlags() {
	flag.StringVar(&Addresses.Server, "a", ":8080", "server address")
	flag.StringVar(&Addresses.RespondingServer, "b", "", "responding server address")
	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		Addresses.Server = envServerAddress
	}
	if envRespondingServerAddress := os.Getenv("RESPONDING_SERVER_ADDRESS"); envRespondingServerAddress != "" {
		Addresses.RespondingServer = envRespondingServerAddress
	}
}
