package main

import (
	"github.com/sirupsen/logrus"

	"github.com/tiunovvv/go-yandex-shortener/cmd/config"
	"github.com/tiunovvv/go-yandex-shortener/pkg/handler"
	"github.com/tiunovvv/go-yandex-shortener/pkg/server"
	"github.com/tiunovvv/go-yandex-shortener/pkg/storage"
)

func main() {
	serverParametrs := config.InitConfig()
	storage := storage.CreateStorage()
	handler := handler.NewHandler(storage, serverParametrs.ShortURLBase)

	srv := new(server.Server)

	if err := srv.Run(serverParametrs.ServerAddress, handler.InitRoutes()); err != nil {
		logrus.Fatalf("error occured while running http server: %s", err.Error())
	}
}
