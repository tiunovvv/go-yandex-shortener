package handler

import (
	"github.com/gin-gonic/gin"
)

type errorResponce struct {
	Message string `json:"message"`
}

func newErrorResponce(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, errorResponce{message})
}
