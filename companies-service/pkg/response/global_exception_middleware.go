package response

import "github.com/gin-gonic/gin"

func GlobalRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.AbortWithStatusJSON(500, Response{
					Success: false,
					Status:  500,
					Message: "Внутренняя ошибка сервера. Мы уже работаем над этим.",
				})
			}
		}()
		c.Next()
	}
}
