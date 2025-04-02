package utils

import (
	"github.com/gin-gonic/gin"
)

// SendError sends a JSON error response with a status code and message
func SendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": message})
}

// SendSuccess sends a JSON success response (can be customized)
func SendSuccess(c *gin.Context, statusCode int, data interface{}) {
	if data != nil {
		c.JSON(statusCode, data)
	} else {
		c.Status(statusCode)
	}
}
