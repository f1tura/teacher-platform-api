package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "role not found in context",
			})
			c.Abort()
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "invalid role type in context",
			})
			c.Abort()
			return
		}

		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		c.Abort()
	}
}