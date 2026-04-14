package usercontext

import (
	"companies-service/pkg/response"
	"context"
)

type ctxKey int

const accountUuidKey ctxKey = iota

func WithAccountUuid(ctx context.Context, uuid string) context.Context {
	return context.WithValue(ctx, accountUuidKey, uuid)
}

func GetAccountUuid(ctx context.Context) (string, error) {
	accountUUID, ok := ctx.Value(accountUuidKey).(string)
	if !ok || accountUUID == "" {
		return "", &response.ApiError{
			Status:  401,
			Message: "Сессия не найдена или не авторизована",
		}
	}
	return accountUUID, nil
}
