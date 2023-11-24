package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	size int
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, fmt.Errorf("error calculating size: %w", err)
}

func GinLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		blw := &bodyLogWriter{ResponseWriter: c.Writer, size: 0}
		c.Writer = blw
		c.Next()
		duration := time.Since(start)

		log.Sugar().Infoln(
			"uri", c.Request.RequestURI,
			"method", c.Request.Method,
			"statusCode", c.Writer.Status(),
			"duration", duration,
			"size", blw.size,
		)
	}
}
