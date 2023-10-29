package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tiunovvv/go-yandex-shortener/pkg/handlers"
)

type Server struct {
	handler *handlers.Handler
}

const port = ":8080"

func NewServer(handler *handlers.Handler) *Server {
	return &Server{handler: handler}
}

func (s *Server) Run() error {
	routers := mux.NewRouter()

	routers.HandleFunc("/", s.handler.PostHandler).Methods("POST")
	routers.HandleFunc("/{id}", s.handler.GetHandler).Methods("GET")

	err := http.ListenAndServe(port, routers)
	if err != nil {
		return err
	}

	return nil
}
