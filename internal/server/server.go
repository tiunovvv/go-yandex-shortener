package server

import (
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(serverAddress string, handler http.Handler) error {
	const (
		seconds = 10
		bytes   = 20
	)

	s.httpServer = &http.Server{
		Addr:           serverAddress,
		Handler:        handler,
		MaxHeaderBytes: 1 << bytes,
		ReadTimeout:    seconds * time.Second,
		WriteTimeout:   seconds * time.Second,
	}

	if err := s.httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("server ListenAndServe error: %w", err)
	}

	return nil
}
