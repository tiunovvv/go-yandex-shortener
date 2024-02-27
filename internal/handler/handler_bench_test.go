package handler

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
	"github.com/tiunovvv/go-yandex-shortener/internal/storage/mocks"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	chars    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	segments = 2
	segLen   = 5
)

func BenchmarkPostHandler(b *testing.B) {
	config := config.GetConfig()
	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	logger, err := logCfg.Build()
	assert.NoError(b, err)
	log := logger.Sugar()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	m := mocks.NewMockStore(ctrl)
	shortener := shortener.NewShortener(m, log)
	handler := NewHandler(config, shortener, log)
	router := handler.InitRoutes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		url := generateRandomURL()
		m.EXPECT().SaveURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		b.StartTimer()
		request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/", bytes.NewReader([]byte(url)))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, request)
		result := w.Result()
		err = result.Body.Close()
		assert.NoError(b, err)
		assert.Equal(b, http.StatusCreated, result.StatusCode)
	}
}

func BenchmarkPostAPI(b *testing.B) {
	config := config.GetConfig()
	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	logger, err := logCfg.Build()
	assert.NoError(b, err)
	log := logger.Sugar()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	m := mocks.NewMockStore(ctrl)
	shortener := shortener.NewShortener(m, log)
	handler := NewHandler(config, shortener, log)
	router := handler.InitRoutes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		urlForSave := generateRandomURL()
		body := `{"url":"` + urlForSave + `"}`
		m.EXPECT().SaveURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		b.StartTimer()
		request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten",
			bytes.NewReader([]byte(body)))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, request)
		result := w.Result()
		err = result.Body.Close()
		assert.NoError(b, err)
		assert.Equal(b, http.StatusCreated, result.StatusCode)
	}
}

func BenchmarkPostAPIBatch(b *testing.B) {
	config := config.GetConfig()
	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	logger, err := logCfg.Build()
	assert.NoError(b, err)
	log := logger.Sugar()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	m := mocks.NewMockStore(ctrl)
	shortener := shortener.NewShortener(m, log)
	handler := NewHandler(config, shortener, log)
	router := handler.InitRoutes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		firstURL := generateRandomURL()
		secondURL := generateRandomURL()
		body := `[{"correlation_id": "1","original_url": "` + firstURL
		body += `"}, {"correlation_id": "2","original_url": "` + secondURL + `"}]`
		m.EXPECT().SaveURLBatch(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		b.StartTimer()
		request := httptest.NewRequest(http.MethodPost,
			"http://localhost:8080/api/shorten/batch", bytes.NewReader([]byte(body)))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, request)
		result := w.Result()
		err = result.Body.Close()
		assert.NoError(b, err)
		assert.Equal(b, http.StatusCreated, result.StatusCode)
	}
}

func BenchmarkPostAPIUserURLs(b *testing.B) {
	config := config.GetConfig()
	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	logger, err := logCfg.Build()
	assert.NoError(b, err)
	log := logger.Sugar()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	m := mocks.NewMockStore(ctrl)
	shortener := shortener.NewShortener(m, log)
	handler := NewHandler(config, shortener, log)
	router := handler.InitRoutes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		m.EXPECT().GetURLByUserID(gomock.Any(),
			gomock.Any()).Return(map[string]string{
			"KCcJLOWm": "http://axIsW.NAPvB",
			"yetoic5G": "http://wJUDw.rHNqd"})
		b.StartTimer()
		request := httptest.NewRequest(http.MethodGet,
			"http://localhost:8080/api/user/urls", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, request)
		result := w.Result()
		err = result.Body.Close()
		assert.NoError(b, err)
		assert.Equal(b, http.StatusOK, result.StatusCode)
	}
}

func BenchmarkGetHandler(b *testing.B) {
	config := config.GetConfig()
	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	logger, err := logCfg.Build()
	assert.NoError(b, err)
	log := logger.Sugar()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	m := mocks.NewMockStore(ctrl)
	shortener := shortener.NewShortener(m, log)
	handler := NewHandler(config, shortener, log)
	router := handler.InitRoutes()

	key := generateRandomSegment(segLen)
	target, err := url.JoinPath("http://localhost:8080/", key)
	assert.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		url := generateRandomURL()
		m.EXPECT().GetFullURL(gomock.Any(), gomock.Any()).Return(url, false, nil)
		b.StartTimer()
		request := httptest.NewRequest(http.MethodGet, target, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, request)
		result := w.Result()
		err = result.Body.Close()
		assert.NoError(b, err)
		assert.Equal(b, http.StatusTemporaryRedirect, result.StatusCode)
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
