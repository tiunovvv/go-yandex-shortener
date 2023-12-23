package storage

import (
	"context"
	"embed"
	"errors"

	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const insertSchemaURLs = `INSERT INTO urls (short_url, full_url, user_id) VALUES ($1, $2, $3);`

type DB struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewDB(ctx context.Context, dsn string, logger *zap.Logger) (Store, error) {
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

	dataBase := &DB{pool: pool, logger: logger}
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

func (db *DB) GetPing(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	return nil
}

func (db *DB) SaveURL(ctx context.Context, shortURL string, fullURL string, userID string) error {
	_, err := db.pool.Exec(ctx, insertSchemaURLs, []byte(shortURL), fullURL, userID)

	if err != nil {
		return fmt.Errorf("failed to insert row: %w", err)
	}

	return nil
}

func (db *DB) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	const selectSchemaFullURL = `SELECT full_url FROM urls WHERE short_url = $1;`

	row := db.pool.QueryRow(ctx, selectSchemaFullURL, shortURL)

	var fullURL string
	if err := row.Scan(&fullURL); err != nil {
		return "", fmt.Errorf("failed to find shortURL=%s in database: %w", shortURL, err)
	}

	return fullURL, nil
}

func (db *DB) GetShortURL(ctx context.Context, fullURL string) string {
	const selectSchemaShortURL = `SELECT short_url FROM urls WHERE full_url = $1;`

	row := db.pool.QueryRow(ctx, selectSchemaShortURL, fullURL)

	var shortURL string
	if err := row.Scan(&shortURL); err != nil {
		return ""
	}

	return shortURL
}

func (db *DB) SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for k, v := range urls {
		if _, err = tx.Exec(ctx, insertSchemaURLs, []byte(k), v, userID); err != nil {
			if tx.Rollback(ctx) != nil {
				return fmt.Errorf("failed to rollback: %w", err)
			}
			return fmt.Errorf("failed to insert row: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (db *DB) GetURLByUserID(ctx context.Context, userID string) map[string]string {
	const selectSchemaURLsByUserID = `SELECT short_url, full_url FROM urls WHERE user_id = $1;`

	rows, err := db.pool.Query(ctx, selectSchemaURLsByUserID, userID)
	if err != nil {
		db.logger.Sugar().Errorf("failed to select by user_id: %w", err)
		return nil
	}

	urls := make(map[string]string)
	for rows.Next() {
		var shortURL, fullURL string
		err := rows.Scan(&shortURL, &fullURL)
		if err != nil {
			db.logger.Sugar().Errorf("failed to get rows from select by user_id: %w", err)
		}
		urls[shortURL] = fullURL
	}

	return urls
}

func (db *DB) Close() error {
	db.pool.Close()
	return nil
}
