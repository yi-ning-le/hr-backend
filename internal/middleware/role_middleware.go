package middleware

import (
	"net/http"

	"hr-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// RequireAdmin middleware checks if the current user is an admin
func RequireAdmin(queries *repository.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := parseUUID(userIDStr.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		isAdmin, err := queries.CheckIsAdmin(c.Request.Context(), userID)
		if err != nil || !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}

		c.Next()
	}
}

// RequireRecruiter middleware checks if the current user is a recruiter
func RequireRecruiter(queries *repository.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := parseUUID(userIDStr.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx := c.Request.Context()

		// Get employee by user ID
		employee, err := queries.GetEmployeeByUserID(ctx, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Employee record not found"})
			return
		}

		// Check if recruiter
		_, err = queries.CheckRecruiterRole(ctx, employee.ID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Recruiter access required"})
			return
		}

		// Store employee ID in context for later use
		c.Set("employeeID", uuidToString(employee.ID))
		c.Next()
	}
}

func parseUUID(s string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(s)
	return uuid, err
}

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return string(u.Bytes[:])
}
