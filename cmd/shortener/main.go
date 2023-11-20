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
	config := config.NewConfig()
	storage := storage.NewStorage(config)
	shortener := shortener.NewShortener(storage)
	logger, err := logger.NewLogger()
	if err != nil {
		log.Printf("error occured while initializing logger: %v", err)
		return
	}
	compressor := compressor.NewCompressor()
	handler := handler.NewHandler(shortener, logger, compressor)

	srv := new(server.Server)
	if err := srv.Run(config.ServerAddress, handler.InitRoutes()); err != nil {
		logger.Sugar().Errorf("error occured while running http server: %v", err)
	}
}
