package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type compressWriter struct {
	io.Writer
	gin.ResponseWriter
}

func (c *compressWriter) Write(p []byte) (int, error) {
	w, err := c.Writer.Write(p)
	return w, fmt.Errorf("failed to write by compressor: %w", err)
}

func GinGzip(log *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		const (
			gZip            = "gzip"
			contentEncoding = "Content-Encoding"
		)

		if strings.Contains(c.GetHeader("Accept-Encoding"), gZip) {
			writer := c.Writer
			newWriter := gzip.NewWriter(writer)
			c.Writer = &compressWriter{Writer: newWriter, ResponseWriter: writer}
			defer func() {
				if err := newWriter.Close(); err != nil {
					log.Errorf("failed to close writer: %w", err)
					c.AbortWithStatus(http.StatusBadRequest)
				}
			}()
			writer.Header().Set(contentEncoding, gZip)
		}

		if strings.Contains(c.GetHeader(contentEncoding), gZip) {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				log.Errorf("failed to read body: %w", err)
				c.AbortWithStatus(http.StatusBadRequest)
			}
			c.Request.Body = io.NopCloser(reader)
			if err := reader.Close(); err != nil {
				log.Errorf("failed to close reader: %w", err)
				c.AbortWithStatus(http.StatusBadRequest)
			}
		}
		c.Next()
	}
}
