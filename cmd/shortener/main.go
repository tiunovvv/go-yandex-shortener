package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tiunovvv/go-yandex-shortener/pkg/handler"
	"github.com/tiunovvv/go-yandex-shortener/pkg/shortener"
)

const port = ":8080"

func main() {

	shorteners := shortener.CreateURLMap()
	handlers := handler.NewHandler(shorteners)
	routers := mux.NewRouter()

	routers.HandleFunc("/", handlers.PostHandler).Methods("POST")
	routers.HandleFunc("/{id}", handlers.GetHandler).Methods("GET")

	err := http.ListenAndServe(port, routers)
	if err != nil {
		panic(err)
	}
}
