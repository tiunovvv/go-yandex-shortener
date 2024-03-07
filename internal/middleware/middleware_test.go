package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGinGzipMiddleware(t *testing.T) {
	r := gin.New()

	logger, _ := zap.NewProduction()
	log := logger.Sugar()

	r.Use(GinGzip(log))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "This is a test response")
	})

	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Add("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !strings.Contains(w.Header().Get("Content-Encoding"), "gzip") {
		t.Error("Response is not compressed with gzip")
	}

	reader, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}

	decompressedBody, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read decompressed body: %v", err)
	}

	err = reader.Close()
	assert.NoError(t, err)

	expectedBody := "This is a test response"
	if string(decompressedBody) != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, decompressedBody)
	}
}

func TestGinLoggerMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	r := gin.New()
	r.Use(GinLogger(logger.Sugar()))

	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Test response")
	})

	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGinTimeOut(t *testing.T) {
	tests := []struct {
		name         string
		timeout      time.Duration
		expectedCode int
	}{
		{
			name:         "Timeout of 500 milliseconds",
			timeout:      250 * time.Millisecond,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Timeout of 750 milliseconds",
			timeout:      750 * time.Millisecond,
			expectedCode: http.StatusGatewayTimeout,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.Use(GinTimeOut(500*time.Millisecond, "Request Timeout"))

			router.GET("/test", func(c *gin.Context) {
				time.Sleep(test.timeout)
				c.String(http.StatusOK, "OK")
			})

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("User-Agent", "test")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			assert.Equal(t, test.expectedCode, resp.Code)
		})
	}
}
