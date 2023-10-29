package main

import (
	"github.com/tiunovvv/go-yandex-shortener/pkg/handlers"
	"github.com/tiunovvv/go-yandex-shortener/pkg/server"
	"github.com/tiunovvv/go-yandex-shortener/pkg/storage"
)

func main() {

	storage := storage.CreateStorage()
	handler := handlers.NewHandler(storage)
	server := server.NewServer(handler)

	err := server.Run()
	if err != nil {
		panic(err)
	}
}
