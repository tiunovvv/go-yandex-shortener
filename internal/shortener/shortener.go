package shortener

import (
	"context"
	"errors"
	"fmt"
	"time"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
	"github.com/tiunovvv/go-yandex-shortener/internal/models"
	"github.com/tiunovvv/go-yandex-shortener/internal/storage"
)

type Shortener struct {
	store storage.Store
}

func NewShortener(store storage.Store) *Shortener {
	return &Shortener{
		store: store,
	}
}

func (sh *Shortener) GetShortURL(ctx context.Context, fullURL string) (string, error) {
	if shortURL := sh.store.GetShortURL(ctx, fullURL); shortURL != "" {
		return shortURL, myErrors.ErrURLAlreadySaved
	}
	shortURL := sh.store.GenerateShortURL()
	for errors.Is(sh.store.SaveURL(ctx, shortURL, fullURL), myErrors.ErrKeyAlreadyExists) {
		shortURL = sh.store.GenerateShortURL()
	}

	return shortURL, nil
}

func (sh *Shortener) GetShortURLBatch(ctx context.Context, fullURL []models.ReqAPIBatch) ([]models.ResAPIBatch, error) {
	var cancelCtx context.CancelFunc
	const seconds = time.Second * 10
	ctx, cancelCtx = context.WithTimeout(ctx, seconds)
	defer cancelCtx()

	res, err := sh.store.GetShortURLBatch(ctx, fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get short URLs: %w", err)
	}

	return res, nil
}

func (sh *Shortener) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	var cancelCtx context.CancelFunc
	const seconds = time.Second * 10
	ctx, cancelCtx = context.WithTimeout(ctx, seconds)
	defer cancelCtx()
	fullURL, err := sh.store.GetFullURL(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("failed to get fullURL from filestore: %w", err)
	}
	return fullURL, nil
}

func (sh *Shortener) CheckConnect(ctx context.Context) error {
	if err := sh.store.GetPing(ctx); err != nil {
		return fmt.Errorf("failed to connect store: %w", err)
	}
	return nil
}
