package storage

import (
	"fmt"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
)

type InMemoryStore struct {
	urls map[string]string
}

func (i *InMemoryStore) SaveURL(shortURL string, fullURL string) error {
	if _, exists := i.urls[shortURL]; exists {
		return fmt.Errorf("failed to save shortURL %s: %w", shortURL, myErrors.ErrKeyAlreadyExists)
	}
	i.urls[shortURL] = fullURL

	return nil
}

func (i *InMemoryStore) GetFullURL(shortURL string) (string, error) {
	if fullURL, found := i.urls[shortURL]; found {
		return fullURL, nil
	}
	return "", fmt.Errorf("URL `%s` not found", shortURL)
}

func (i *InMemoryStore) GetShortURL(fullURL string) string {
	for key, value := range i.urls {
		if value == fullURL {
			return key
		}
	}
	return ""
}
