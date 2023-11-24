package storage

import (
	"fmt"

	"github.com/tiunovvv/go-yandex-shortener/internal/errors"
)

type MemoryStorage struct {
	Urls map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Urls: make(map[string]string),
	}
}

func (s *MemoryStorage) SaveURL(url string, shortURL string) error {
	if _, exists := s.Urls[shortURL]; exists {
		return fmt.Errorf("can`t save shortURL %s: %w", shortURL, errors.ErrKeyAlreadyExists)
	}
	s.Urls[shortURL] = url

	return nil
}

func (s *MemoryStorage) GetFullURL(shortURL string) string {
	if fullURL, found := s.Urls[shortURL]; found {
		return fullURL
	}
	return ""
}

func (s *MemoryStorage) FindByFullURL(url string) string {
	for key, value := range s.Urls {
		if value == url {
			return key
		}
	}
	return ""
}

func (s *MemoryStorage) FindByShortURL(url string) bool {
	for key := range s.Urls {
		if key == url {
			return true
		}
	}
	return false
}
