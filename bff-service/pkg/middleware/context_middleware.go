package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

type contextKey string

const AccountUUIDKey contextKey = "x-account-uuid"

func ContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountUUID := c.GetHeader("X-Account-Uuid")

		if accountUUID != "" {
			ctx := context.WithValue(c.Request.Context(), AccountUUIDKey, accountUUID)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}
