package handler

import "github.com/gin-gonic/gin"

func (h *Handler) getUserID(c *gin.Context) string {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		return ""
	}

	userID, ok := userIDInterface.(string)

	if !ok {
		h.log.Errorf("failed to get userID from %v", userIDInterface)
		return ""
	}

	return userID
}
