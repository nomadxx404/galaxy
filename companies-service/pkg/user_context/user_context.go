package usercontext

import (
	"context"
)

type ctxKey int

const accountUuidKey ctxKey = iota

func WithAccountUuid(ctx context.Context, uuid string) context.Context {
	return context.WithValue(ctx, accountUuidKey, uuid)
}

func GetAccountUuid(ctx context.Context) string {
	account_uuid, _ := ctx.Value(accountUuidKey).(string)
	return account_uuid
}
