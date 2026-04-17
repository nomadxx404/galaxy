package middleware

import (
	"companies-service/internal/repository/db"
	rdb "companies-service/internal/repository/redis"
	"companies-service/pkg/response"
	usercontext "companies-service/pkg/user_context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func HasAccess(userMask int64, requiredPermission int64) bool {
	if userMask == -1 {
		return true
	}
	return (userMask & requiredPermission) != 0
}

func PermissionMiddleware(
	store *db.Queries,
	rdbClient *redis.Client,
	domain string,
	requiredBit int64) gin.HandlerFunc {

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		account_uuid := usercontext.GetAccountUuid(ctx)

		company_uuid := c.Param("company_uuid")
		if company_uuid == "" {
			response.SendFailure(c, 400, "ID компании не указан")
			return
		}

		cacheKey := rdb.GetUserMaskKey(company_uuid, account_uuid, domain)
		var mask int64

		cachedMask, err := rdbClient.Get(ctx, cacheKey).Result()

		if err == nil {
			mask, _ = strconv.ParseInt(cachedMask, 10, 64)
		} else {
			dbMask, err := store.GetUserPermissionMask(ctx, db.GetUserPermissionMaskParams{
				CompanyUuid: company_uuid,
				AccountUuid: account_uuid,
				Domain:      domain,
			})

			if err != nil {
				mask = 0
			} else {
				mask = dbMask.Mask
			}

			rdbClient.Set(ctx, cacheKey, strconv.FormatInt(mask, 10), 5*time.Minute)
		}

		if !HasAccess(mask, requiredBit) {
			response.SendFailure(c, 403, "Недостаточно прав для выполнения операции")
			c.Abort()
			return
		}

		c.Next()
	}
}
