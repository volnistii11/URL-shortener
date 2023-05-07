package config

import "flag"

var Addresses struct {
	Server           string
	RespondingServer string
}

func ParseFlags() {

	flag.StringVar(&Addresses.Server, "a", "http://localhost:8080", "server address")
	flag.StringVar(&Addresses.RespondingServer, "b", "http://localhost:8080", "responding server address")
	flag.Parse()
}
