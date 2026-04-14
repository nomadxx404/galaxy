package middleware

import (
	"companies-service/pkg/response"
	usercontext "companies-service/pkg/user_context"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountUUID := c.GetHeader("X-Account-Uuid")

		if accountUUID == "" {
			response.SendFailure(c, 401, "X-Account-Uuid header is missing")
			c.Abort()
			return
		}

		newCtx := usercontext.WithAccountUuid(c.Request.Context(), accountUUID)
		c.Request = c.Request.WithContext(newCtx)

		c.Next()
	}
}
