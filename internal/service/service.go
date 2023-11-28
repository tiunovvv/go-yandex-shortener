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
	fileStorage := storage.NewFileStore(config, log)
	shortener := shortener.NewShortener(fileStorage)
	handler := handler.NewHandler(config, shortener, log)

	srv := new(server.Server)
	if err := srv.Run(config.ServerAddress, handler.InitRoutes()); err != nil {
		log.Sugar().Errorf("error occured while running http server: %v", err)
	}
}
