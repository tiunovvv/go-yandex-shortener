package server

import (
	"net/http"
	"time"

	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/handler"
	"go.uber.org/zap"
)

// InitServer creates and returns new Server.
func InitServer(h *handler.Handler, cfg *config.Config, log *zap.Logger) *http.Server {
	const (
		maxHeaderBytes = 20
		handlerTimeout = 5 * time.Second
	)

	errorLog := zap.NewStdLog(log)

	return &http.Server{
		Addr:           cfg.ServerAddress,
		Handler:        h.InitRoutes(),
		MaxHeaderBytes: 1 << maxHeaderBytes,
		ErrorLog:       errorLog,
		ReadTimeout:    handlerTimeout,
		WriteTimeout:   handlerTimeout,
	}
}
