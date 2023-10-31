package config

import (
	"flag"
)

const (
	ServerAddress = "localhost:8080"
	ShortURLBase  = "http://localhost:8080/"
)

type ServerParametrs struct {
	ServerAddress string
	ShortURLBase  string
}

func InitConfig() *ServerParametrs {
	serverParametrs := &ServerParametrs{}

	flag.StringVar(&serverParametrs.ServerAddress, "a", ServerAddress, "HTTP server start address")
	flag.StringVar(&serverParametrs.ShortURLBase, "b", ShortURLBase, "Base address of the resulting shortened URL")

	flag.Parse()

	return serverParametrs
}
