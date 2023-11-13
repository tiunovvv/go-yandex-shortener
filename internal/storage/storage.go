package storage

import (
	"fmt"

	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/errors"
)

type Storage struct {
	Config *config.Config
	Urls   map[string]string
}

func NewStorage(cfg *config.Config) *Storage {
	return &Storage{
		Config: cfg,
		Urls:   make(map[string]string),
	}
}

func (s *Storage) SaveURL(url string, shortURL string) error {
	if _, exists := s.Urls[shortURL]; exists {
		return fmt.Errorf("cannot save shortURL %s: %w", shortURL, errors.ErrKeyAlreadyExists)
	}
	s.Urls[shortURL] = url
	return nil
}

func (s *Storage) GetFullURL(shortURL string) string {
	if fullURL, found := s.Urls[shortURL]; found {
		return fullURL
	}
	return ""
}

func (s *Storage) FindByFullURL(url string) string {
	for key, value := range s.Urls {
		if value == url {
			return key
		}
	}
	return ""
}

func (s *Storage) FindByShortURL(url string) bool {
	for key := range s.Urls {
		if key == url {
			return true
		}
	}
	return false
}
