package main

import (
	"log"

	"github.com/tiunovvv/go-yandex-shortener/internal/logger"
	"github.com/tiunovvv/go-yandex-shortener/internal/service"
)

func main() {
	logger, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("error occured while initializing logger: %v", err)
	}

	service.Start(logger)
}
