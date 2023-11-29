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
	*http.Server
}

func NewServer() (*Server, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("error building logger: %w", err)
	}
	config := config.NewConfig(logger)
	fileStorage := storage.NewFileStore(config, logger)
	shortener := shortener.NewShortener(fileStorage)
	handler := handler.NewHandler(config, shortener, logger)
	errorLog := zap.NewStdLog(logger)
	const (
		seconds = 10 * time.Second
		bytes   = 20
	)
	s := http.Server{
		Addr:           config.ServerAddress,
		Handler:        handler.InitRoutes(),
		ErrorLog:       errorLog,
		MaxHeaderBytes: 1 << bytes,
		ReadTimeout:    seconds,
		WriteTimeout:   seconds,
	}

	return &Server{logger, &s}, nil
}

func (s *Server) Start() error {
	var err error
	defer func() {
		err = fmt.Errorf("error sync logger: %w", s.logger.Sync())
	}()
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal("Could not listen on", zap.String("addr", s.Addr), zap.Error(err))
		}
	}()
	s.logger.Info("Server is ready to handle requests", zap.String("addr", s.Addr))
	s.gracefulShutdown()
	return err
}

func (s *Server) gracefulShutdown() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)
	sig := <-quit
	s.logger.Info("Server is shutting down", zap.String("reason", sig.String()))
	const seconds = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), seconds)
	defer cancel()

	s.SetKeepAlivesEnabled(false)
	if err := s.Shutdown(ctx); err != nil {
		s.logger.Fatal("Could not gracefully shutdown the server", zap.Error(err))
	}
	s.logger.Info("Server stopped")
}
