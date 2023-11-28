package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("error building logger: %w", err)
	}

	defer func() {
		err = fmt.Errorf("error sync logger: %w", logger.Sync())
	}()

	return logger, err
}
