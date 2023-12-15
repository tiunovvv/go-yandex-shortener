package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GinTimeOut(dt time.Duration, msg string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), dt)
		defer cancel()
		done := make(chan bool, 1)

		go func() {
			c.Next()
			done <- true
		}()

		select {
		case <-done:
		case <-ctx.Done():
			c.String(http.StatusGatewayTimeout, msg)
			c.Abort()
		}
	}
}
