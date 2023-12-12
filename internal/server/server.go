package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/handler"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
	"github.com/tiunovvv/go-yandex-shortener/internal/storage"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger
	store  storage.Store
	*http.Server
}

func NewServer(ctx context.Context) (*Server, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	config := config.NewConfig(logger)
	store, err := storage.NewStore(ctx, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	shortener := shortener.NewShortener(store)
	handler := handler.NewHandler(config, shortener, logger)

	errorLog := zap.NewStdLog(logger)
	const (
		bytes   = 20
		seconds = 10 * time.Second
	)
	s := http.Server{
		Addr:           config.ServerAddress,
		Handler:        handler.InitRoutes(),
		ErrorLog:       errorLog,
		MaxHeaderBytes: 1 << bytes,
		ReadTimeout:    seconds,
		WriteTimeout:   seconds,
	}

	return &Server{logger, store, &s}, nil
}

func (s *Server) Start() error {
	var err error
	defer func() {
		if er := s.logger.Sync(); er != nil {
			err = fmt.Errorf("failed to sync logger: %w", er)
		}
	}()

	defer func() {
		if er := s.store.Close(); er != nil {
			err = fmt.Errorf("failed to close store: %w", er)
		}
	}()

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("could not listen on", zap.String("addr", s.Addr), zap.Error(err))
		}
	}()

	s.logger.Info("server is ready to handle requests", zap.String("addr", s.Addr))
	s.gracefulShutdown()
	return err
}

func (s *Server) gracefulShutdown() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)
	sig := <-quit
	s.logger.Info("server is shutting down", zap.String("reason", sig.String()))
	const seconds = 10 * time.Second
	ctx, cancelCtx := context.WithTimeout(context.Background(), seconds)
	defer cancelCtx()

	s.SetKeepAlivesEnabled(false)
	if err := s.Shutdown(ctx); err != nil {
		s.logger.Error("failed to gracefully shutdown the server", zap.Error(err))
	}
	s.logger.Info("server is stopped")
}
