package storage

import (
	"context"
	"fmt"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
)

// URLInfo structure of in-memory map.
type URLInfo struct {
	fullURL     string
	userID      string
	DeletedFlag bool
}

// Memory is in-memory storage.
type Memory struct {
	urls map[string]URLInfo
}

// NewMemory creates new in-memory storage.
func NewMemory() Store {
	return &Memory{map[string]URLInfo{}}
}

// GetShortURL shortURL by fullURL from in-memmory.
func (i *Memory) GetShortURL(ctx context.Context, fullURL string) string {
	for key, value := range i.urls {
		if value.fullURL == fullURL {
			return key
		}
	}
	return ""
}

// GetFullURL fullURL by shortURL from in-memmory.
func (i *Memory) GetFullURL(ctx context.Context, shortURL string) (string, bool, error) {
	if urlInfo, found := i.urls[shortURL]; found {
		return urlInfo.fullURL, urlInfo.DeletedFlag, nil
	}
	return "", false, fmt.Errorf("URL `%s` not found", shortURL)
}

// SaveURL saves all info about new URL in local map.
func (i *Memory) SaveURL(ctx context.Context, shortURL string, fullURL string, userID string) error {
	for _, value := range i.urls {
		if value.fullURL == fullURL {
			return myErrors.ErrURLAlreadySaved
		}
	}

	if _, exists := i.urls[shortURL]; exists {
		return fmt.Errorf("failed to save shortURL %s: %w", shortURL, myErrors.ErrKeyAlreadyExists)
	}
	i.urls[shortURL] = URLInfo{fullURL, userID, false}

	return nil
}

// SaveURLBatch saves all info about new URL list in local map.
func (i *Memory) SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error {
	for k, v := range urls {
		if err := i.SaveURL(ctx, k, v, userID); err != nil {
			return myErrors.ErrKeyAlreadyExists
		}
	}

	return nil
}

// GetURLByUserID returns list of user URLs from in-memmory.
func (i *Memory) GetURLByUserID(ctx context.Context, userID string) map[string]string {
	urls := make(map[string]string)
	for key, value := range i.urls {
		if value.userID == userID {
			urls[key] = value.fullURL
		}
	}
	return urls
}

// SetDeletedFlag sets deleted flag for short URL in local map.
func (i *Memory) SetDeletedFlag(ctx context.Context, userID string, shortURL string) error {
	url, exists := i.urls[shortURL]

	if !exists {
		return fmt.Errorf("failed to update deleted flag for short_url=%s", shortURL)
	}

	if url.userID != userID {
		return fmt.Errorf("failed to update deleted flag for short_url=%s, different userID ", shortURL)
	}
	url.DeletedFlag = true
	i.urls[shortURL] = url
	return nil
}

// GetPing method needed for implementation of Store interface.
func (i *Memory) GetPing(ctx context.Context) error {
	return nil
}

// Close method needed for implementation of Store interface.
func (i *Memory) Close() error {
	return nil
}
