package main

import (
	"log"

	"github.com/tiunovvv/go-yandex-shortener/internal/server"
)

func main() {
	server, err := server.NewServer()
	if err != nil {
		log.Fatalf("error building server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
