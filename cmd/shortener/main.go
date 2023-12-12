package main

import (
	"context"
	"log"
	"time"

	"github.com/tiunovvv/go-yandex-shortener/internal/server"
)

func main() {
	const seconds = 10 * time.Second
	ctx, cancelCtx := context.WithTimeout(context.Background(), seconds)

	server, err := server.NewServer(ctx)
	if err != nil {
		log.Fatalf("error building server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
	cancelCtx()
}
