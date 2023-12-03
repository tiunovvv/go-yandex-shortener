package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"go.uber.org/zap"
)

type DataBase struct {
	*pgx.Conn
}

func ConnectDB(config *config.Config, logger *zap.Logger) *DataBase {
	conn, err := pgx.Connect(context.Background(), config.DatabaseDsn)
	if err != nil {
		logger.Sugar().Errorf("error connecting to db: %v", err)
	}

	dataBase := &DataBase{conn}
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
