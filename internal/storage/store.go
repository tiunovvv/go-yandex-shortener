package storage

import (
	"context"

	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/models"
	"go.uber.org/zap"
)

type Store interface {
	GetShortURL(ctx context.Context, fullURL string) string
	GetFullURL(ctx context.Context, shortURL string) (string, error)
	GetShortURLBatch(ctx context.Context, fullURL []models.ReqAPIBatch) ([]models.ResAPIBatch, error)
	SaveURL(ctx context.Context, shortURL string, fullURL string) error
	GetPing(ctx context.Context) error
	Close() error
	GenerateShortURL() string
}

func NewStore(ctx context.Context, config *config.Config, logger *zap.Logger) (Store, error) {
	if len(config.DSN) != 0 {
		return NewDatabaseStore(ctx, config.DSN, logger)
	}

	if len(config.FilePath) != 0 {
		return NewFileStore(config.FilePath, logger)
	}

	return NewInMemoryStore(), nil
}
