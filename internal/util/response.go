package util

import (
	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, gin.H{
		"status": gin.H{
			"code":    statusCode,
			"message": "Success",
		},
		"data": data,
	})
}

func PaginatedResponse(c *gin.Context, statusCode int, data interface{}, pagination interface{}) {
	c.JSON(statusCode, gin.H{
		"status": gin.H{
			"code":    statusCode,
			"message": "Success",
		},
		"data":       data,
		"pagination": pagination,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string, details interface{}) {
	c.JSON(statusCode, gin.H{
		"status": gin.H{
			"code":    statusCode,
			"message": message,
		},
		"errors": details,
	})
}