package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type (
	bodyLogWriter struct {
		gin.ResponseWriter
		size int
	}
)

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

func InitLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	defer logger.Sync()
	return logger, nil
}

func WithLogging(s *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		blw := &bodyLogWriter{ResponseWriter: c.Writer, size: 0}
		c.Writer = blw
		uri := c.Request.RequestURI
		method := c.Request.Method

		c.Next()

		statusCode := c.Writer.Status()
		duration := time.Since(start)

		s.Infoln(
			"uri", uri,
			"method", method,
			"statusCode", statusCode,
			"duration", duration,
			"size", blw.size,
		)
	}
}
