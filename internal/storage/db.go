package storage

import (
	"context"

	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/generator"
	"github.com/tiunovvv/go-yandex-shortener/internal/models"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
	"go.uber.org/zap"
)

type DataBase struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	*generator.Generator
}

func NewDB(
	ctx context.Context,
	config *config.Config,
	logger *zap.Logger,
	queryTracer *queryTracer,
) (shortener.Store, error) {
	if config.DSN == "" {
		return nil, fmt.Errorf("DSN is empty")
	}

	poolCfg, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}

	poolCfg.ConnConfig.Tracer = queryTracer
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	createSchemaURLs :=
		`CREATE TABLE IF NOT EXISTS urls(
			short_url CHAR(8) PRIMARY KEY NOT NULL,
			full_url VARCHAR(200) NOT NULL
		)`

	_, err = pool.Exec(ctx, createSchemaURLs)

	if err != nil {
		return nil, fmt.Errorf("failed to creati table URLS: %w", err)
	}

	dataBase := &DataBase{pool: pool, logger: logger}
	return dataBase, nil
}

func (db *DataBase) GetPing(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	return nil
}

func (db *DataBase) SaveURL(ctx context.Context, shortURL string, fullURL string) error {
	insertSchemaURLs :=
		`INSERT INTO urls (short_url, full_url) 
			VALUES ($1, $2);`

	_, err := db.pool.Exec(ctx, insertSchemaURLs, db.stringToChar8(shortURL), fullURL)

	if err != nil {
		return fmt.Errorf("failed to insert row: %w", err)
	}

	return nil
}

func (db *DataBase) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	selectSchemaFullURL :=
		`SELECT full_url FROM urls 
			WHERE short_url = $1;`

	row := db.pool.QueryRow(ctx, selectSchemaFullURL, shortURL)

	var fullURL string
	if err := row.Scan(&fullURL); err != nil {
		return "", fmt.Errorf("failed to find shortURL=%s in database: %w", shortURL, err)
	}

	return fullURL, nil
}

func (db *DataBase) GetShortURL(ctx context.Context, fullURL string) string {
	selectSchemaShortURL :=
		`SELECT short_url FROM urls 
			WHERE full_url = $1;`

	row := db.pool.QueryRow(ctx, selectSchemaShortURL, fullURL)

	var shortURL string
	if err := row.Scan(&shortURL); err != nil {
		return ""
	}

	return shortURL
}

func (db *DataBase) GetShortURLBatch(ctx context.Context, reqSlice []models.ReqAPIBatch) ([]models.ResAPIBatch, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin batch: %w", err)
	}

	insertSchemaURLs :=
		`INSERT INTO urls (short_url, full_url)
			VALUES ($1, $2);`

	var resSlice []models.ResAPIBatch
	for _, req := range reqSlice {
		if shortURL := db.GetShortURL(ctx, req.FullURL); shortURL != "" {
			res := models.ResAPIBatch{ID: req.ID, ShortURL: shortURL}
			resSlice = append(resSlice, res)
			continue
		}

		res := models.ResAPIBatch{ID: req.ID, ShortURL: db.GenerateShortURL()}
		result, err := tx.Exec(ctx, insertSchemaURLs, db.stringToChar8(res.ShortURL), req.FullURL)
		if err != nil {
			if tx.Rollback(ctx) != nil {
				return nil, fmt.Errorf("failed to rollback: %w", err)
			}
			return nil, fmt.Errorf("failed to insert row: %w", err)
		}

		for result.RowsAffected() == 0 {
			res = models.ResAPIBatch{ID: req.ID, ShortURL: db.GenerateShortURL()}
			result, err = tx.Exec(ctx, insertSchemaURLs, db.stringToChar8(res.ShortURL), req.FullURL)
			if err != nil {
				if tx.Rollback(ctx) != nil {
					return nil, fmt.Errorf("failed to rollback: %w", err)
				}
				return nil, fmt.Errorf("failed to insert row: %w", err)
			}
			continue
		}
		resSlice = append(resSlice, res)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return resSlice, nil
}

func (db *DataBase) CloseStore() error {
	db.pool.Close()
	return nil
}

func (db *DataBase) stringToChar8(input string) string {
	const length = 8
	char8Bytes := make([]byte, length)
	copy(char8Bytes, input)
	char8Value := string(char8Bytes)
	return char8Value
}
