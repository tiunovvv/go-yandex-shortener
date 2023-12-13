package storage

import (
	"context"
	"embed"
	"errors"
	"math/rand"
	"time"

	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tiunovvv/go-yandex-shortener/internal/models"
	"go.uber.org/zap"
)

const insertSchemaURLs = `INSERT INTO urls (short_url, full_url) VALUES ($1, $2);`

type DataBase struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewDatabaseStore(ctx context.Context, dsn string, logger *zap.Logger) (Store, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}

	queryTracer := NewQueryTracer(logger)
	poolCfg.ConnConfig.Tracer = queryTracer
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	dataBase := &DataBase{pool: pool, logger: logger}
	return dataBase, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func (db *DataBase) GetPing(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	return nil
}

func (db *DataBase) SaveURL(ctx context.Context, shortURL string, fullURL string) error {
	_, err := db.pool.Exec(ctx, insertSchemaURLs, []byte(shortURL), fullURL)

	if err != nil {
		return fmt.Errorf("failed to insert row: %w", err)
	}

	return nil
}

func (db *DataBase) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	const selectSchemaFullURL = `SELECT full_url FROM urls WHERE short_url = $1;`

	row := db.pool.QueryRow(ctx, selectSchemaFullURL, shortURL)

	var fullURL string
	if err := row.Scan(&fullURL); err != nil {
		return "", fmt.Errorf("failed to find shortURL=%s in database: %w", shortURL, err)
	}

	return fullURL, nil
}

func (db *DataBase) GetShortURL(ctx context.Context, fullURL string) string {
	const selectSchemaShortURL = `SELECT short_url FROM urls WHERE full_url = $1;`

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

	resSlice := make([]models.ResAPIBatch, 0, len(reqSlice))
	for _, req := range reqSlice {
		if shortURL := db.GetShortURL(ctx, req.FullURL); shortURL != "" {
			res := models.ResAPIBatch{ID: req.ID, ShortURL: shortURL}
			resSlice = append(resSlice, res)
			continue
		}

		res := models.ResAPIBatch{ID: req.ID, ShortURL: db.GenerateShortURL()}
		result, err := tx.Exec(ctx, insertSchemaURLs, []byte(res.ShortURL), req.FullURL)
		if err != nil {
			if tx.Rollback(ctx) != nil {
				return nil, fmt.Errorf("failed to rollback: %w", err)
			}
			return nil, fmt.Errorf("failed to insert row: %w", err)
		}

		for result.RowsAffected() == 0 {
			res = models.ResAPIBatch{ID: req.ID, ShortURL: db.GenerateShortURL()}
			result, err = tx.Exec(ctx, insertSchemaURLs, []byte(res.ShortURL), req.FullURL)
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

func (db *DataBase) Close() error {
	db.pool.Close()
	return nil
}

func (db *DataBase) GenerateShortURL() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	const length = 8
	str := make([]byte, length)

	charset := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for i := range str {
		str[i] = charset[rand.Intn(len(charset))]
	}

	return string(str)
}
