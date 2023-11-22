package storage

import (
	"fmt"
	"strconv"

	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/errors"
)

type Storage struct {
	Config   *config.Config
	Urls     map[string]string
	Producer *Producer
	Consumer *Consumer
}

func NewStorage(cfg *config.Config) (*Storage, error) {
	urls := make(map[string]string)

	if cfg.FileStoragePath == "" {
		return &Storage{
			Config:   cfg,
			Urls:     urls,
			Consumer: nil,
			Producer: nil,
		}, nil
	}

	producer, err := NewProducer(cfg.FileStoragePath)
	if err != nil {
		return nil, fmt.Errorf("can`t create producer %s: %w", cfg.FileStoragePath, err)
	}

	consumer, err := NewConsumer(cfg.FileStoragePath)
	if err != nil {
		return nil, fmt.Errorf("can`t create consumer %s: %w", cfg.FileStoragePath, err)
	}

	if err := consumer.ReadURLs(urls); err != nil {
		return nil, fmt.Errorf("can`t read URLs from file %s: %w", cfg.FileStoragePath, err)
	}

	return &Storage{
		Config:   cfg,
		Urls:     urls,
		Consumer: consumer,
		Producer: producer}, nil
}

func (s *Storage) SaveURL(url string, shortURL string) error {
	if _, exists := s.Urls[shortURL]; exists {
		return fmt.Errorf("can`t save shortURL %s: %w", shortURL, errors.ErrKeyAlreadyExists)
	}
	s.Urls[shortURL] = url

	if s.Producer == nil {
		return nil
	}

	u := URLsJSON{UUID: strconv.Itoa(len(s.Urls)), ShortURL: shortURL, OriginalURL: url}
	if err := s.Producer.WriteURL(u); err != nil {
		return err
	}

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
