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

func (s *MemoryStore) GetFullURL(shortURL string) string {
	if fullURL, found := s.Urls[shortURL]; found {
		return fullURL
	}
	return ""
}

func (s *MemoryStore) FindByFullURL(url string) string {
	for key, value := range s.Urls {
		if value == url {
			return key
		}
	}
	return ""
}

func (s *MemoryStore) FindByShortURL(url string) bool {
	for key := range s.Urls {
		if key == url {
			return true
		}
	}
	return false
}
