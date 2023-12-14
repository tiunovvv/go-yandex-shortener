package storage

import (
	"context"
	"fmt"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
)

type Memory struct {
	urls map[string]string
}

func NewMemory() Store {
	return &Memory{map[string]string{}}
}

func (i *Memory) GetShortURL(ctx context.Context, fullURL string) string {
	for key, value := range i.urls {
		if value == fullURL {
			return key
		}
	}
	return ""
}

func (i *Memory) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	if fullURL, found := i.urls[shortURL]; found {
		return fullURL, nil
	}
	return "", fmt.Errorf("URL `%s` not found", shortURL)
}

func (i *Memory) SaveURL(ctx context.Context, shortURL string, fullURL string) error {
	if _, exists := i.urls[shortURL]; exists {
		return fmt.Errorf("failed to save shortURL %s: %w", shortURL, myErrors.ErrKeyAlreadyExists)
	}
	i.urls[shortURL] = fullURL

	return nil
}

func (i *Memory) SaveURLBatch(ctx context.Context, urls map[string]string) error {
	for k, v := range urls {
		if err := i.SaveURL(ctx, k, v); err != nil {
			return myErrors.ErrKeyAlreadyExists
		}
	}

	return nil
}

func (i *Memory) GetPing(ctx context.Context) error {
	return nil
}

func (i *Memory) Close() error {
	return nil
}
