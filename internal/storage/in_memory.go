package storage

import (
	"fmt"

	"github.com/tiunovvv/go-yandex-shortener/internal/errors"
)

type MemoryStore struct {
	Urls map[string]string
}

func (s *MemoryStore) SaveURL(url string, shortURL string) error {
	if _, exists := s.Urls[shortURL]; exists {
		return fmt.Errorf("can`t save shortURL %s: %w", shortURL, errors.ErrKeyAlreadyExists)
	}
	s.Urls[shortURL] = url

	return nil
}

func (s *MemoryStore) GetFullURL(shortURL string) (string, error) {
	if fullURL, found := s.Urls[shortURL]; found {
		return fullURL, nil
	}
	return "", fmt.Errorf("URL `%s` not found", shortURL)
}

func (s *MemoryStore) GetShortURL(fullURL string) string {
	for key, value := range s.Urls {
		if value == fullURL {
			return key
		}
	}
	return ""
}
