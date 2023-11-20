package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Compressor struct {
}

func NewCompressor() *Compressor {
	return &Compressor{}
}

type compressWriter struct {
	io.Writer
	gin.ResponseWriter
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.Writer.Write(p)
}

func (comp *Compressor) GinGzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		const (
			gZip            = "gzip"
			contentEncoding = "Content-Encoding"
		)

		if strings.Contains(c.GetHeader("Accept-Encoding"), gZip) {
			writer := c.Writer
			newWriter := gzip.NewWriter(writer)
			defer newWriter.Close()
			c.Writer = &compressWriter{Writer: newWriter, ResponseWriter: writer}
			writer.Header().Set(contentEncoding, gZip)
		}

		if strings.Contains(c.GetHeader(contentEncoding), gZip) {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			defer reader.Close()
			c.Request.Body = io.NopCloser(reader)
		}
		c.Next()
	}
}
