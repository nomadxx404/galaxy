package usercontext

import (
	"context"
)

type ctxKey int

const (
	accountUuidKey ctxKey = iota
	requestIdKey
	methodKey
	pathKey
)

func WithAccountUuid(ctx context.Context, uuid string) context.Context {
	return context.WithValue(ctx, accountUuidKey, uuid)
}

func GetAccountUuid(ctx context.Context) string {
	account_uuid, _ := ctx.Value(accountUuidKey).(string)
	return account_uuid
}

func WithRequestId(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, requestIdKey, rid)
}

func GetRequestId(ctx context.Context) string {
	requestId, _ := ctx.Value(requestIdKey).(string)
	return requestId
}

func WithMetadata(ctx context.Context, method, path string) context.Context {
	ctx = context.WithValue(ctx, methodKey, method)
	return context.WithValue(ctx, pathKey, path)
}

func GetMethod(ctx context.Context) string {
	val, _ := ctx.Value(methodKey).(string)
	return val
}

func GetPath(ctx context.Context) string {
	val, _ := ctx.Value(pathKey).(string)
	return val
}
