package response

import (
	usercontext "companies-service/pkg/user_context"
	"log"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func GlobalRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := usercontext.GetRequestId(c.Request.Context())

				log.Printf(
					"[PANIC] | ID: %v | Error: %v\nStack Trace:\n%s",
					requestID,
					err,
					debug.Stack(),
				)

				c.AbortWithStatusJSON(500, Response{
					Success:   false,
					Status:    500,
					Message:   "Внутренняя ошибка сервера. Мы уже работаем над этим.",
					RequestId: requestID,
				})
			}
		}()
		c.Next()
	}
}
