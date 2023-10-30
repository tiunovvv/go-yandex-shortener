package main

import (
	"github.com/sirupsen/logrus"

	"github.com/tiunovvv/go-yandex-shortener/pkg/handler"
	"github.com/tiunovvv/go-yandex-shortener/pkg/server"
	"github.com/tiunovvv/go-yandex-shortener/pkg/storage"
)

const port = "8080"

func main() {

	storage := storage.CreateStorage()
	handler := handler.NewHandler(storage)

	srv := new(server.Server)
	if err := srv.Run(string(port), handler.InitRoutes()); err != nil {
		logrus.Fatalf("error occured while running http server: %s", err.Error())
	}
}
