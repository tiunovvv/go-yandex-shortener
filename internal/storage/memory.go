package storage

import (
	"context"
	"fmt"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
)

type URLInfo struct {
	fullURL     string
	userID      string
	DeletedFlag bool
}
type Memory struct {
	urls map[string]URLInfo
}

func NewMemory() Store {
	return &Memory{map[string]URLInfo{}}
}

func (i *Memory) GetShortURL(ctx context.Context, fullURL string) string {
	for key, value := range i.urls {
		if value.fullURL == fullURL {
			return key
		}
	}
	return ""
}

func (i *Memory) GetFullURL(ctx context.Context, shortURL string) (string, bool, error) {
	if urlInfo, found := i.urls[shortURL]; found {
		return urlInfo.fullURL, urlInfo.DeletedFlag, nil
	}
	return "", false, fmt.Errorf("URL `%s` not found", shortURL)
}

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

func (i *Memory) SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error {
	for k, v := range urls {
		if err := i.SaveURL(ctx, k, v, userID); err != nil {
			return myErrors.ErrKeyAlreadyExists
		}
	}

	return nil
}

func (i *Memory) GetURLByUserID(ctx context.Context, userID string) map[string]string {
	urls := make(map[string]string)
	for key, value := range i.urls {
		if value.userID == userID {
			urls[key] = value.fullURL
		}
	}
	return urls
}

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

func (i *Memory) GetPing(ctx context.Context) error {
	return nil
}

func (i *Memory) Close() error {
	return nil
}
