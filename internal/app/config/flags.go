package config

import "flag"

var Addresses struct {
	Server           string
	RespondingServer string
}

func ParseFlags() {

	flag.StringVar(&Addresses.Server, "a", ":8080", "server address")
	flag.StringVar(&Addresses.RespondingServer, "b", "", "responding server address")
	flag.Parse()
}
