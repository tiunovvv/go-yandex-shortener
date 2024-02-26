package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
	"github.com/tiunovvv/go-yandex-shortener/internal/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	baseURL  = "http://localhost:8080/"
	chars    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	segments = 2
	segLen   = 5
)

func initRouter() (*gin.Engine, storage.Store, error) {
	config := config.GetConfig()

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)

	logger, err := cfg.Build()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	log := logger.Sugar()

	store := storage.NewMemory()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create storage: %w", err)
	}

	shortener := shortener.NewShortener(store, log)
	handler := NewHandler(config, shortener, log)
	router := handler.InitRoutes()
	return router, store, nil
}

func BenchmarkPostHandler(b *testing.B) {
	router, _, err := initRouter()
	assert.Equal(b, err, nil)

	url := generateRandomURL()

	request := httptest.NewRequest(http.MethodPost, baseURL, bytes.NewReader([]byte(url)))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := httptest.NewRequest(http.MethodPost, baseURL, bytes.NewReader([]byte(url)))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, request)
		result := w.Result()
		err = result.Body.Close()
		assert.Equal(b, err, nil)
		assert.Equal(b, http.StatusConflict, result.StatusCode)
	}
}

func BenchmarkGetHandler(b *testing.B) {
	router, store, err := initRouter()
	assert.Equal(b, err, nil)

	const seconds = 10 * time.Second
	ctx, cancelCtx := context.WithTimeout(context.Background(), seconds)
	defer cancelCtx()

	urlForSave := generateRandomURL()
	key := generateRandomSegment(segLen)

	err = store.SaveURL(ctx, key, urlForSave, "")
	assert.Equal(b, err, nil)

	target, err := url.JoinPath(baseURL, key)
	assert.Equal(b, err, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := httptest.NewRequest(http.MethodGet, target, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, request)
		result := w.Result()
		err = result.Body.Close()
		assert.Equal(b, err, nil)
		assert.Equal(b, http.StatusTemporaryRedirect, result.StatusCode)
		assert.Equal(b, urlForSave, result.Header.Get("Location"))
	}
}

func BenchmarkPostAPI(b *testing.B) {
	router, store, err := initRouter()
	assert.Equal(b, err, nil)

	const seconds = 10 * time.Second
	ctx, cancelCtx := context.WithTimeout(context.Background(), seconds)
	defer cancelCtx()
	key := generateRandomSegment(segLen)
	urlForSave := generateRandomURL()
	body := `{"url":"` + urlForSave + `"}`

	err = store.SaveURL(ctx, key, urlForSave, "")
	assert.Equal(b, err, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten",
			bytes.NewReader([]byte(body)))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, request)
		result := w.Result()
		err = result.Body.Close()
		assert.Equal(b, err, nil)
		assert.Equal(b, http.StatusConflict, result.StatusCode)
	}
}

func BenchmarkPostAPIBatch(b *testing.B) {
	router, _, err := initRouter()
	assert.Equal(b, err, nil)

	firstURL := generateRandomURL()
	secondURL := generateRandomURL()

	body := `[{"correlation_id": "1","original_url": "` + firstURL
	body += `"}, {"correlation_id": "2","original_url": "` + secondURL + `"}]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := httptest.NewRequest(http.MethodPost,
			"http://localhost:8080/api/shorten/batch", bytes.NewReader([]byte(body)))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, request)
		result := w.Result()
		err = result.Body.Close()
		assert.Equal(b, err, nil)
		assert.Equal(b, http.StatusCreated, result.StatusCode)
	}
}

func generateRandomURL() string {
	url := "http://"
	for i := 0; i < segments; i++ {
		segment := generateRandomSegment(segLen)
		url += segment
		if i < segments-1 {
			url += "."
		}
	}
	return url
}

func generateRandomSegment(length int) string {
	segment := make([]byte, length)
	for i := range segment {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			panic(err) // handle error
		}
		segment[i] = chars[n.Int64()]
	}
	return string(segment)
}
