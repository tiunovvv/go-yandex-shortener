package main

import (
	"log"

	"github.com/tiunovvv/go-yandex-shortener/internal/compressor"
	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/handler"
	"github.com/tiunovvv/go-yandex-shortener/internal/logger"
	"github.com/tiunovvv/go-yandex-shortener/internal/server"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
	"github.com/tiunovvv/go-yandex-shortener/internal/storage"
)

func main() {
	logger, err := logger.NewLogger()
	if err != nil {
		log.Printf("error occured while initializing logger: %v", err)
		return
	}

	config := config.NewConfig(logger)

	storage, err := storage.NewStorage(config)
	if err != nil {
		logger.Sugar().Errorf("error occured while initializing storage: %v", err)
		return
	}

	shortener := shortener.NewShortener(storage)
	compressor := compressor.NewCompressor()
	handler := handler.NewHandler(shortener, logger, compressor)

	srv := new(server.Server)
	if err := srv.Run(config.ServerAddress, handler.InitRoutes()); err != nil {
		logger.Sugar().Errorf("error occured while running http server: %v", err)
	}
}
