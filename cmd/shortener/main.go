package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/handler"
	"github.com/tiunovvv/go-yandex-shortener/internal/server"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
	"github.com/tiunovvv/go-yandex-shortener/internal/storage"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

const (
	timeoutServerShutdown = time.Second * 5
	timeoutShutdown       = time.Second * 10
)

func run() error {
	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	})

	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
	}()

	cfg := config.GetConfig()

	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("failed to initialize logger %w", err)
	}

	log := logger.Sugar()
	store, err := storage.NewStore(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	sh := shortener.NewShortener(store, log)

	watch(ctx, wg, log, store)

	h := handler.NewHandler(cfg, sh, log)
	srv := server.InitServer(h, cfg, logger)

	componentsErrs := make(chan error, 1)

	manageServer(ctx, wg, srv, componentsErrs)

	select {
	case <-ctx.Done():
	case err := <-componentsErrs:
		log.Error(err)
		cancelCtx()
	}

	return nil
}

func watch(ctx context.Context, wg *sync.WaitGroup, log *zap.SugaredLogger, s storage.Store) {
	wg.Add(1)
	go func() {
		defer log.Info("closed DB and stoped Dispatcher")
		defer wg.Done()

		<-ctx.Done()

		if err := s.Close(); err != nil {
			log.Errorf("failed to close store: %w", err)
		}

		if err := log.Sync(); err != nil {
			log.Errorf("failed to sync logger: %w", err)
		}
	}()
}

func manageServer(ctx context.Context, wg *sync.WaitGroup, srv *http.Server, errs chan<- error) {
	go func(errs chan<- error) {
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			errs <- fmt.Errorf("listen and server has failed: %w", err)
		}
	}(errs)

	wg.Add(1)

	go func() {
		defer log.Print("server has been shutdown")
		defer wg.Done()
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()
		if err := srv.Shutdown(shutdownTimeoutCtx); err != nil {
			log.Printf("an error occurred during server shutdown: %v", err)
		}
	}()
}
