package shortener

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
	"github.com/tiunovvv/go-yandex-shortener/internal/models"
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

func initShortener() (*Shortener, error) {
	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)

	logger, err := loggerCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	log := logger.Sugar()

	store := storage.NewMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	shortener := NewShortener(store, log)
	return shortener, nil
}

func BenchmarkGetShortURL(b *testing.B) {
	sh, err := initShortener()
	assert.Equal(b, err, nil)

	const seconds = 100 * time.Second
	ctx, cancelCtx := context.WithTimeout(context.Background(), seconds)
	defer cancelCtx()

	userID, err := generateUniqueUserID()
	assert.Equal(b, err, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fullURL := generateRandomURL()
		b.StartTimer()
		_, err := sh.GetShortURL(ctx, fullURL, userID)
		if err != nil {
			assert.NotErrorIs(b, err, myErrors.ErrURLAlreadySaved)
		}
	}
}

func BenchmarkGetShortURLBatch(b *testing.B) {
	sh, err := initShortener()
	assert.Equal(b, err, nil)

	const seconds = 100 * time.Second
	ctx, cancelCtx := context.WithTimeout(context.Background(), seconds)
	defer cancelCtx()

	userID, err := generateUniqueUserID()
	assert.Equal(b, err, nil)

	var req models.ReqAPIBatch
	reqSlice := make([]models.ReqAPIBatch, 0, 50)

	for i := 0; i < 50; i++ {
		shortURL := generateShortURL()
		fullShortURL, err := url.JoinPath(baseURL, shortURL)
		assert.Equal(b, err, nil)
		req.ID = strconv.Itoa(i + 1)
		req.FullURL = fullShortURL
		reqSlice = append(reqSlice, req)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sh.GetShortURLBatch(ctx, reqSlice, userID)
		assert.Equal(b, err, nil)
	}
}

func BenchmarkGetFullURL(b *testing.B) {
	sh, err := initShortener()
	assert.Equal(b, err, nil)

	const seconds = 100 * time.Second
	ctx, cancelCtx := context.WithTimeout(context.Background(), seconds)
	defer cancelCtx()

	userID, err := generateUniqueUserID()
	assert.Equal(b, err, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fullURLOrig := generateRandomURL()
		shortURL, err := sh.GetShortURL(ctx, fullURLOrig, userID)
		if err != nil {
			assert.NotErrorIs(b, err, myErrors.ErrURLAlreadySaved)
		}
		b.StartTimer()
		fullURL, _, err := sh.GetFullURL(ctx, shortURL)
		assert.Equal(b, err, nil)
		assert.Equal(b, fullURLOrig, fullURL)
	}
}

func BenchmarkGetURLByUserID(b *testing.B) {
	sh, err := initShortener()
	assert.Equal(b, err, nil)

	const seconds = 100 * time.Second
	ctx, cancelCtx := context.WithTimeout(context.Background(), seconds)
	defer cancelCtx()

	userID, err := generateUniqueUserID()
	assert.Equal(b, err, nil)

	for i := 0; i < 50; i++ {
		fullURL := generateRandomURL()
		_, err := sh.GetShortURL(ctx, fullURL, userID)
		if err != nil {
			assert.NotErrorIs(b, err, myErrors.ErrURLAlreadySaved)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sh.GetURLByUserID(ctx, baseURL, userID)
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

func generateUniqueUserID() (string, error) {
	uuid, err := uuid.NewV4()
	return uuid.String(), err
}
