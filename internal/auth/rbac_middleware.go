package auth

import (
	"e-memo-job-reservation-api/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequirePermission(permissionName string, permissionRepo *repository.PositionPermissionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		positionID, exists := c.Get("user_position_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User position not found in token"})
			return
		}

		hasPermission, err := permissionRepo.CheckPermission(positionID.(int), permissionName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not verify permissions"})
			return
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You don't have permission to perform this action"})
			return
		}

		c.Next()
	}
}
