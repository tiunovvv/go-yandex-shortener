package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

func SetCookie(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		const userIDKey = "user_id"
		session := sessions.Default(c)

		if userID, ok := session.Get(userIDKey).(string); ok {
			c.Set(userIDKey, userID)
			c.Next()
			return
		}

		userID, err := generateUniqueUserID()
		if err != nil {
			log.Sugar().Errorf("failed to generate userID: %w", err)
			c.AbortWithStatus(http.StatusBadRequest)
		}

		session.Set(userIDKey, userID)
		if err := session.Save(); err != nil {
			log.Sugar().Errorf("failed to set userID: %w", err)
			c.AbortWithStatus(http.StatusBadRequest)
		}

		c.Set(userIDKey, userID)
		c.Next()
	}
}

func generateUniqueUserID() (string, error) {
	uuid, err := uuid.NewV4()
	return uuid.String(), err
}
