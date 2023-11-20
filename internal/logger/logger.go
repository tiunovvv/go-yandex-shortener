package logger

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Logger struct {
	*zap.Logger
}

type (
	bodyLogWriter struct {
		gin.ResponseWriter
		size int
	}
)

func NewLogger() (*Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("error initializing logger: %w", err)
	}
	logger.Sync()

	return &Logger{logger}, nil
}

func (l *Logger) GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		blw := &bodyLogWriter{ResponseWriter: c.Writer, size: 0}
		c.Writer = blw
		c.Next()
		duration := time.Since(start)

		l.Sugar().Infoln(
			"uri", c.Request.RequestURI,
			"method", c.Request.Method,
			"statusCode", c.Writer.Status(),
			"duration", duration,
			"size", blw.size,
		)
	}
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, fmt.Errorf("error calculating size: %w", err)
}
