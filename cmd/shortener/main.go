package main

import (
	"context"
	"log"

	"github.com/tiunovvv/go-yandex-shortener/internal/server"
)

func main() {
	ctx, cancelCtx := context.WithCancel(context.Background())

	server, err := server.NewServer(ctx)
	if err != nil {
		log.Fatalf("error building server: %v", err)
	}

	if err := server.Start(cancelCtx); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
