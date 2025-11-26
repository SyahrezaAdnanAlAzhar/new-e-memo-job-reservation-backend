package auth

import (
	"net/http"

	"e-memo-job-reservation-api/internal/repository"
	"e-memo-job-reservation-api/internal/util"

	"github.com/gin-gonic/gin"
)

type EditModeMiddleware struct {
	authRepo *repository.AuthRepository
}

func NewEditModeMiddleware(authRepo *repository.AuthRepository) *EditModeMiddleware {
	return &EditModeMiddleware{authRepo: authRepo}
}

func (m *EditModeMiddleware) CheckEditMode() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if exists && userType.(string) == "master" {
			c.Next()
			return
		}

		isEditing, err := m.authRepo.GetEditMode(c.Request.Context())
		if err != nil {
			util.ErrorResponse(c, http.StatusServiceUnavailable, "Failed to verify system status", err.Error())
			c.Abort()
			return
		}

		if isEditing {
			util.ErrorResponse(c, http.StatusServiceUnavailable, "Service Unavailable", "System is currently in edit mode. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}
