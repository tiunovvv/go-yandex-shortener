package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"go.uber.org/zap"
)

type DataBase struct {
	fileStore *FileStore
	*pgx.Conn
	logger *zap.Logger
}

func NewDB(config *config.Config, logger *zap.Logger) *DataBase {
	if config.DatabaseDsn == "" {
		fileStorage := NewFileStore(config, logger)
		dataBase := &DataBase{fileStorage, nil, logger}
		return dataBase
	}

	conn, err := pgx.Connect(context.Background(), config.DatabaseDsn)
	if err != nil {
		logger.Sugar().Errorf("error connecting to db: %v", err)
	}

	_, err = conn.Exec(context.Background(),
		"CREATE TABLE IF NOT EXISTS urls (short_url TEXT,full_url TEXT,PRIMARY KEY (short_url))")

	if err != nil {
		logger.Sugar().Errorf("error creating table URLS: %v", err)
	}

	fileStorage := NewFileStore(config, logger)
	dataBase := &DataBase{fileStorage, conn, logger}
	return dataBase
}

func (db *DataBase) CheckConnect() error {
	if db == nil {
		return fmt.Errorf("error connecting to db")
	}
	if err := db.Ping(context.Background()); err != nil {
		return fmt.Errorf("error connecting to db: %w", err)
	}
	return nil
}

func (db *DataBase) SaveURL(shortURL string, fullURL string) error {
	if db.Conn == nil {
		return db.fileStore.SaveURL(shortURL, fullURL)
	}

	_, err := db.Exec(context.Background(),
		"INSERT INTO urls (short_url, full_url) VALUES ($1, $2);", shortURL, fullURL)

	if err != nil {
		return fmt.Errorf("error inserting into table URLS: %w", err)
	}

	return nil
}

func (db *DataBase) GetFullURL(shortURL string) (string, error) {
	if db.Conn == nil {
		return db.fileStore.GetFullURL(shortURL)
	}

	var fullURL string
	row := db.QueryRow(context.Background(),
		"SELECT full_url FROM urls WHERE short_url = $1;", shortURL)

	if err := row.Scan(&fullURL); err != nil {
		return "", fmt.Errorf("no fullURL in db for shortURL=%s error: %w", shortURL, err)
	}

	return fullURL, nil
}

func (db *DataBase) GetShortURL(fullURL string) string {
	if db.Conn == nil {
		return db.fileStore.GetShortURL(fullURL)
	}

	var shortURL string
	row := db.QueryRow(context.Background(),
		"SELECT short_url FROM urls WHERE full_url = $1;", fullURL)

	if err := row.Scan(&shortURL); err != nil {
		db.logger.Sugar().Errorf("no shortURL in db for fullURL=%s error: %w", fullURL, err)
		return ""
	}

	return shortURL
}
