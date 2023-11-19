package main

import (
	"log"

	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/handler"
	"github.com/tiunovvv/go-yandex-shortener/internal/logger"
	"github.com/tiunovvv/go-yandex-shortener/internal/server"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
	"github.com/tiunovvv/go-yandex-shortener/internal/storage"
)

func main() {
	logger, err := logger.InitLogger()
	if err != nil {
		log.Printf("error occured while initializing logger: %v", err)
		return
	}

	config := config.NewConfig()
	storage := storage.NewStorage(config)
	shortener := shortener.NewShortener(storage)
	handler := handler.NewHandler(shortener, logger)

	srv := new(server.Server)
	if err := srv.Run(config.ServerAddress, handler.InitRoutes()); err != nil {
		log.Printf("error occured while running http server: %v", err)
	}
}
