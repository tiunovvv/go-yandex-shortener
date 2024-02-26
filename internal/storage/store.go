package storage

import (
	"context"

	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"go.uber.org/zap"
)

type Store interface {
	GetShortURL(ctx context.Context, fullURL string) string
	GetFullURL(ctx context.Context, shortURL string) (string, bool, error)
	GetURLByUserID(ctx context.Context, userID string) map[string]string
	SaveURL(ctx context.Context, shortURL string, fullURL string, userID string) error
	SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error
	SetDeletedFlag(ctx context.Context, userID string, shortURL string) error
	GetPing(ctx context.Context) error
	Close() error
}

func NewStore(ctx context.Context, cfg *config.Config, log *zap.SugaredLogger) (Store, error) {
	if len(cfg.DSN) != 0 {
		store, err := NewDB(ctx, cfg.DSN, log)
		if err == nil {
			return store, nil
		}
		log.Errorf("failed to create storage using DB: %w", err)
	}

	if len(cfg.FilePath) != 0 {
		store, err := NewFile(cfg.FilePath, log)
		if err == nil {
			return store, nil
		}
		log.Errorf("failed to create storage using File: %w", err)
	}

	return NewMemory(), nil
}
