package handler_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/handler"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
	"github.com/tiunovvv/go-yandex-shortener/internal/storage"
	"go.uber.org/zap"
)

func ExampleHandler_PostHandler() {
	config := &config.Config{
		BaseURL:       "http://localhost:8080/",
		ServerAddress: "localhost:8080",
		FilePath:      "",
		DSN:           "",
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("failed to initialize logger: %v", err)
		return
	}
	log := logger.Sugar()

	store := storage.NewMemory()
	shortener := shortener.NewShortener(store, log)
	handler := handler.NewHandler(config, shortener, log)
	err = store.SaveURL(context.Background(), "qwerasdf", "http://www.example.com", "userID")
	if err != nil {
		fmt.Printf("failed to save URL: %v", err)
		return
	}
	router := handler.InitRoutes()
	request := httptest.NewRequest(http.MethodPost,
		"http://localhost:8080/", bytes.NewReader([]byte("http://www.example.com")))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)
	result := w.Result()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		fmt.Printf("failed to read body: %v", err)
		return
	}
	fmt.Println(string(body))
	err = result.Body.Close()
	if err != nil {
		fmt.Printf("failed to close body: %v", err)
		return
	}

	// Output:
	// http://localhost:8080/qwerasdf
}
