package storage

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
	"github.com/tiunovvv/go-yandex-shortener/internal/models"
)

type InMemoryStore struct {
	urls map[string]string
}

func NewInMemoryStore() Store {
	return &InMemoryStore{map[string]string{}}
}

func (i *InMemoryStore) GetShortURL(ctx context.Context, fullURL string) string {
	for key, value := range i.urls {
		if value == fullURL {
			return key
		}
	}
	return ""
}

func (i *InMemoryStore) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	if fullURL, found := i.urls[shortURL]; found {
		return fullURL, nil
	}
	return "", fmt.Errorf("URL `%s` not found", shortURL)
}

func (i *InMemoryStore) SaveURL(ctx context.Context, shortURL string, fullURL string) error {
	if _, exists := i.urls[shortURL]; exists {
		return fmt.Errorf("failed to save shortURL %s: %w", shortURL, myErrors.ErrKeyAlreadyExists)
	}
	i.urls[shortURL] = fullURL

	return nil
}

func (i *InMemoryStore) GetShortURLBatch(
	ctx context.Context,
	reqSlice []models.ReqAPIBatch,
) ([]models.ResAPIBatch, error) {
	resSlice := make([]models.ResAPIBatch, 0, len(reqSlice))
	for _, req := range reqSlice {
		if shortURL := i.GetShortURL(ctx, req.FullURL); shortURL != "" {
			res := models.ResAPIBatch{ID: req.ID, ShortURL: shortURL}
			resSlice = append(resSlice, res)
			continue
		}
		res := models.ResAPIBatch{ID: req.ID, ShortURL: i.GenerateShortURL()}
		for errors.Is(i.SaveURL(ctx, res.ShortURL, req.FullURL), myErrors.ErrKeyAlreadyExists) {
			res.ShortURL = i.GenerateShortURL()
			continue
		}
		resSlice = append(resSlice, res)
	}
	return resSlice, nil
}

func (i *InMemoryStore) GetPing(ctx context.Context) error {
	return fmt.Errorf("failed to connect to database")
}

func (i *InMemoryStore) Close() error {
	return nil
}

func (i *InMemoryStore) GenerateShortURL() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	const length = 8
	str := make([]byte, length)

	charset := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for i := range str {
		str[i] = charset[rand.Intn(len(charset))]
	}

	return string(str)
}
