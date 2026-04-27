package service

import (
	"companies-service/internal/dto/request"
	"companies-service/internal/permissions"
	"companies-service/internal/repository/db"
	rdb "companies-service/internal/repository/redis"
	"companies-service/pkg/data"
	"companies-service/pkg/kafka"
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
	rdb       *redis.Client
	pool      *pgxpool.Pool
	txManager *data.TransactionManager
}

func NewPermissionService(
	rdb *redis.Client,
	pool *pgxpool.Pool,
	txManager *data.TransactionManager) *PermissionService {

	return &PermissionService{
		rdb:       rdb,
		pool:      pool,
		txManager: txManager,
	}
}

func (s *PermissionService) CreatePermissions(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string,
	account_uuid string) error {

	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		for _, domain := range permissions.AllDomainPermission {
			err := q.GiveAccess(ctx, db.GiveAccessParams{
				CompanyUuid: company_uuid,
				AccountUuid: account_uuid,
				Domain:      domain,
				Mask:        -1,
				ChangeBy:    account_uuid,
			})
			if err != nil {
				return &response.ApiError{
					Status:  500,
					Message: "Ошибка выдачи прав доступа",
					Data:    err,
				}
			}
		}

		payload := map[string]any{
			"company_uuid": company_uuid,
			"account_uuid": account_uuid,
			"change_by":    account_uuid,
		}

		if err := kafka.EmitOutbox(ctx, tx, kafka.PermissionCreated, payload); err != nil {
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка создания записи в outbox",
				Data:    err,
			}
		}
		return nil
	})

	return err
}

func (s *PermissionService) UpdatePermissions(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string,
	account_uuid string,
	req request.UpdatePermissionsRequest) error {

	account_uuid_change_by := usercontext.GetAccountUuid(ctx)

	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		for _, p := range req.Permissions {

			if !IsValidDomain(p.Domain) {
				return &response.ApiError{
					Status:  400,
					Message: "Недопустимый домен прав",
				}
			}

			err := q.UpdatePermissions(ctx, db.UpdatePermissionsParams{
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

		payload := map[string]any{
			"company_uuid": company_uuid,
			"account_uuid": account_uuid,
			"permissions":  req.Permissions,
		}

		if err := kafka.EmitOutbox(ctx, tx, kafka.PermissionUpdated, payload); err != nil {
			return err
		}
		return nil
	})

	keysToDelete := make([]string, 0, len(req.Permissions)+1)
	keysToDelete = append(keysToDelete, rdb.GetPermissionKey(company_uuid, account_uuid))

	for _, p := range req.Permissions {
		keysToDelete = append(keysToDelete, rdb.GetUserMaskKey(company_uuid, account_uuid, p.Domain))
	}

	s.rdb.Del(ctx, keysToDelete...)

	return err
}

func (s *PermissionService) GetAccountPermission(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid,
	account_uuid string) ([]db.GetUserAllPermissionsRow, error) {

	if dbExecutor == nil {
		dbExecutor = s.pool
	}
	q := db.New(dbExecutor)

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

	permission, err := q.GetUserAllPermissions(ctx, db.GetUserAllPermissionsParams{
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

func (s *PermissionService) DeleteAllPermissions(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string,
	account_uuid string) error {

	var affectedAccounts []string
	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		account_uuids, err := q.DeleteAllPermissionsCompany(ctx, db.DeleteAllPermissionsCompanyParams{
			CompanyUuid: company_uuid,
			ChangeBy:    account_uuid,
		})

		if err != nil {
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка деактивации прав доступа",
				Data:    err,
			}
		}

		affectedAccounts = account_uuids

		payload := map[string]any{
			"company_uuid":  company_uuid,
			"account_uuids": account_uuids,
			"change_by":     account_uuid,
		}

		if err := kafka.EmitOutbox(ctx, tx, kafka.PermissionDeleted, payload); err != nil {
			return err
		}

		return nil
	})

	if err == nil && len(affectedAccounts) > 0 {
		totalKeys := len(affectedAccounts) * (len(permissions.AllDomainPermission) + 1)
		keysToDelete := make([]string, 0, totalKeys)

		for _, accUUID := range affectedAccounts {
			keysToDelete = append(keysToDelete, rdb.GetPermissionKey(company_uuid, accUUID))

			for _, domain := range permissions.AllDomainPermission {
				keysToDelete = append(keysToDelete, rdb.GetUserMaskKey(company_uuid, accUUID, domain))
			}
		}

		if len(keysToDelete) > 0 {
			s.rdb.Del(ctx, keysToDelete...)
		}
	}

	return err
}

func (s *PermissionService) DeleteMemberPermissions(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string,
	target_account_uuid string,
	account_uuid string) error {

	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		err := q.DeleteAllPermissionsMember(ctx, db.DeleteAllPermissionsMemberParams{
			CompanyUuid: company_uuid,
			AccountUuid: target_account_uuid,
			ChangeBy:    account_uuid,
		})

		if err != nil {
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка при удалении прав пользователя",
				Data:    err,
			}
		}

		payload := map[string]any{
			"company_uuid":        company_uuid,
			"target_account_uuid": target_account_uuid,
			"change_by":           account_uuid,
		}

		if err := kafka.EmitOutbox(ctx, tx, kafka.PermissionDeleted, payload); err != nil {
			return err
		}

		return nil
	})

	if err == nil {
		keysToDelete := make([]string, 0, len(permissions.AllDomainPermission)+1)
		keysToDelete = append(keysToDelete, rdb.GetPermissionKey(company_uuid, target_account_uuid))

		for _, domain := range permissions.AllDomainPermission {
			keysToDelete = append(keysToDelete, rdb.GetUserMaskKey(company_uuid, target_account_uuid, domain))
		}

		s.rdb.Del(ctx, keysToDelete...)
	}

	return err
}

func IsValidDomain(domain string) bool {
	for _, d := range permissions.AllDomainPermission {
		if d == domain {
			return true
		}
	}
	return false
}
