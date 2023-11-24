package service

import (
	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/handler"
	"github.com/tiunovvv/go-yandex-shortener/internal/server"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
	"github.com/tiunovvv/go-yandex-shortener/internal/storage"
	"go.uber.org/zap"
)

func Start(log *zap.Logger) {
	config := config.NewConfig(log)

	memoryStorage := storage.NewMemoryStorage()

	fileStorage := storage.NewFileStorage(config.FileStoragePath, memoryStorage)

	if err := fileStorage.LoadURLs(); err != nil {
		log.Sugar().Errorf("error loading URLs from temp file: %v", err)
		return
	}

	shortener := shortener.NewShortener(memoryStorage)

	handler := handler.NewHandler(config, shortener, log, fileStorage)

	srv := new(server.Server)
	if err := srv.Run(config.ServerAddress, handler.InitRoutes()); err != nil {
		log.Sugar().Errorf("error occured while running http server: %v", err)
	}
}
