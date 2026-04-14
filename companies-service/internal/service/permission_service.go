package service

import (
	"companies-service/internal/dto/request"
	"companies-service/internal/permissions"
	"companies-service/internal/repository/db"
	rdb "companies-service/internal/repository/redis"
	"companies-service/pkg/response"
	usercontext "companies-service/pkg/user_context"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PermissionService struct {
	store *db.Queries
	rdb   *redis.Client
	pool  *pgxpool.Pool
}

func NewPermissionService(
	store *db.Queries,
	rdb *redis.Client,
	pool *pgxpool.Pool) *PermissionService {

	return &PermissionService{
		store: store,
		rdb:   rdb,
		pool:  pool,
	}
}

func (s *PermissionService) UpdatePermissions(
	ctx context.Context,
	company_uuid string,
	account_uuid string,
	req request.UpdatePermissionsRequest) error {

	tx, err := s.pool.Begin(ctx)

	if err != nil {
		return &response.ApiError{
			Status:  500,
			Message: "Ошибка создания транзакции",
			Data:    err,
		}
	}

	defer tx.Rollback(ctx)

	qTx := s.store.WithTx(tx)

	account_uuid_change_by, err := usercontext.GetAccountUuid(ctx)

	for _, p := range req.Permissions {

		if !IsValidDomain(p.Domain) {
			return &response.ApiError{
				Status:  400,
				Message: "Недопустимый домен прав",
				Data:    err,
			}
		}

		err := qTx.UpdatePermissions(ctx, db.UpdatePermissionsParams{
			CompanyUuid: company_uuid,
			AccountUuid: account_uuid,
			Domain:      p.Domain,
			Mask:        p.Mask,
			ChangeBy:    account_uuid_change_by,
		})
		if err != nil {
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка выдачи прав",
				Data:    err,
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return &response.ApiError{
			Status:  500,
			Message: "Ошибка сохранения транзакции",
			Data:    err,
		}
	}

	keysToDelete := make([]string, 0, len(req.Permissions)+1)
	keysToDelete = append(keysToDelete, rdb.GetPermissionKey(company_uuid, account_uuid))

	for _, p := range req.Permissions {
		keysToDelete = append(keysToDelete, rdb.GetUserMaskKey(company_uuid, account_uuid, p.Domain))
	}

	s.rdb.Del(ctx, keysToDelete...)

	return nil
}

func (s *PermissionService) GetAccountPermission(ctx context.Context, company_uuid, account_uuid string) ([]db.GetUserAllPermissionsRow, error) {

	key := rdb.GetPermissionKey(company_uuid, account_uuid)
	permission_cache, err := s.rdb.Get(ctx, key).Result()

	if err == nil {
		var cachedRow []db.GetUserAllPermissionsRow
		if err := json.Unmarshal([]byte(permission_cache), &cachedRow); err == nil {
			return cachedRow, nil
		}
	} else if err != redis.Nil {
		log.Printf("[REDIS-ERROR] %v", err)
	}

	permission, err := s.store.GetUserAllPermissions(ctx, db.GetUserAllPermissionsParams{
		CompanyUuid: company_uuid,
		AccountUuid: account_uuid,
	})

	if err != nil {
		return []db.GetUserAllPermissionsRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка получения прав доступа",
			Data:    err,
		}
	}

	if permission == nil {
		permission = []db.GetUserAllPermissionsRow{}
	}

	if jsonBytes, err := json.Marshal(permission); err == nil {
		s.rdb.Set(ctx, key, jsonBytes, 10*time.Minute)
	}

	return permission, nil
}

func IsValidDomain(domain string) bool {
	for _, d := range permissions.AllDomainPermission {
		if d == domain {
			return true
		}
	}
	return false
}
