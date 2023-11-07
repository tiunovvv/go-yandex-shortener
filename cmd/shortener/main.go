package main

import (
	"fmt"

	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/handler"
	"github.com/tiunovvv/go-yandex-shortener/internal/server"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
	"github.com/tiunovvv/go-yandex-shortener/internal/storage"
)

func main() {
	config := config.NewConfig()
	storage := storage.NewStorage(config)
	shortener := shortener.NewShortener(storage)
	handler := handler.NewHandler(shortener)

	srv := new(server.Server)
	if err := srv.Run(config.ServerAddress, handler.InitRoutes()); err != nil {
		fmt.Printf("error occured while running http server: %v", err)
	}
}
