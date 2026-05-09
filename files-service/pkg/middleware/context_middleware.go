package middleware

import (
	"files-service/pkg/response"
	usercontext "files-service/pkg/user_context"

	"github.com/gin-gonic/gin"
)

func ContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountUUID := c.GetHeader("X-Account-Uuid")
		requestID := c.GetHeader("X-Request-ID")

		if accountUUID == "" {
			response.SendFailure(c, 401, "X-Account-Uuid отсутствует")
			c.Abort()
			return
		}

		ctx := c.Request.Context()

		ctx = usercontext.WithAccountUuid(ctx, accountUUID)

		if requestID != "" {
			ctx = usercontext.WithRequestId(ctx, requestID)
		}

		ctx = usercontext.WithMetadata(ctx, c.Request.Method, c.FullPath())

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
