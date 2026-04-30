package service

import (
	"companies-service/internal/dto/entity"
	"companies-service/internal/dto/request"
	"companies-service/internal/repository/db"
	rdb "companies-service/internal/repository/redis"
	"companies-service/pkg/data"
	"companies-service/pkg/kafka"
	"companies-service/pkg/response"
	usercontext "companies-service/pkg/user_context"
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type MemberPermissionManager interface {
	DeleteMemberPermissions(ctx context.Context, dbExecutor db.DBTX, company_uuid string, target_account_uuid string, account_uuid string) error
}

type memberService struct {
	rdb               *redis.Client
	pool              *pgxpool.Pool
	txManager         *data.TransactionManager
	permissionService MemberPermissionManager
}

func NewMemberService(
	rdb *redis.Client,
	pool *pgxpool.Pool,
	txManager *data.TransactionManager,
	permissionService MemberPermissionManager) *memberService {

	return &memberService{
		rdb:               rdb,
		pool:              pool,
		txManager:         txManager,
		permissionService: permissionService,
	}
}

func (s *memberService) CreateMember(
	ctx context.Context,
	dbExecutor db.DBTX,
	ent entity.CreateMemberEntity) error {

	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		err := q.CreateMember(ctx, db.CreateMemberParams{
			CompanyUuid: ent.CompanyUuid,
			AccountUuid: ent.AccountUuid,
			RoleID:      ent.RoleID,
			IsOwner:     ent.IsOwner,
		})

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return &response.ApiError{
					Status:  409,
					Message: "Вы уже являетесь участником этой компании",
				}
			}
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка добавления участника",
				Data:    err,
			}
		}

		payload := map[string]any{
			"company_uuid": ent.CompanyUuid,
			"account_uuid": ent.AccountUuid,
			"role_id":      ent.RoleID,
			"is_owner":     ent.IsOwner,
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.MemberCreated, payload); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (s *memberService) IsCompanyOwner(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string,
	account_uuid string) (bool, error) {

	if dbExecutor == nil {
		dbExecutor = s.pool
	}
	q := db.New(dbExecutor)

	isOwner, err := q.IsCompanyOwner(ctx, db.IsCompanyOwnerParams{
		CompanyUuid: company_uuid,
		AccountUuid: account_uuid,
	})

	if err != nil {
		return false, &response.ApiError{
			Status:  500,
			Message: "Ошибка проверки прав доступа",
			Data:    err,
		}
	}

	return isOwner, nil
}

func (s *memberService) UpdateRoleMember(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string,
	account_uuid string,
	req request.UpdateRoleMemberRequest) error {

	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		rowsAffected, err := q.UpdateRoleMember(ctx, db.UpdateRoleMemberParams{
			CompanyUuid: company_uuid,
			AccountUuid: account_uuid,
			RoleID:      req.RoleId,
		})

		if err != nil {
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка обновления роли",
				Data:    err,
			}
		}

		if rowsAffected == 0 {
			return &response.ApiError{
				Status:  404,
				Message: "Пользователь не найден в данной компании",
			}
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return &response.ApiError{
				Status:  400,
				Message: "Указанная роль не существует",
			}
		}

		payload := map[string]any{
			"company_uuid": company_uuid,
			"account_uuid": account_uuid,
			"role_id":      req.RoleId,
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.MemberRoleUpdated, payload); err != nil {
			return err
		}

		return nil
	})

	key := rdb.GetMembersKey(company_uuid)
	s.rdb.Del(ctx, key)

	return err
}

func (s *memberService) SetOwner(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string,
	account_uuid string) error {

	account_uuid_session := usercontext.GetAccountUuid(ctx)

	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		isOwner, err := s.IsCompanyOwner(ctx, tx, company_uuid, account_uuid_session)
		if err != nil {
			return err
		}

		if !isOwner {
			return &response.ApiError{
				Status:  403,
				Message: "Сделать пользователя владельцем компании может только активный владелец",
			}
		}

		rowsAffected, err := q.SetOwner(ctx, db.SetOwnerParams{
			CompanyUuid: company_uuid,
			AccountUuid: account_uuid,
		})

		if err != nil {
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка обновления роли пользователя",
				Data:    err,
			}
		}

		if rowsAffected == 0 {
			return &response.ApiError{
				Status:  404,
				Message: "Пользователь не найден в данной компании",
			}
		}

		payload := map[string]any{
			"company_uuid": company_uuid,
			"account_uuid": account_uuid,
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.SetOwner, payload); err != nil {
			return err
		}

		return nil
	})

	key := rdb.GetMembersKey(company_uuid)
	s.rdb.Del(ctx, key)

	return err
}

func (s *memberService) DeleteMember(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string,
	target_account_uuid string) error {

	initiator_account_uuid := usercontext.GetAccountUuid(ctx)
	isSelfRemoval := initiator_account_uuid == target_account_uuid

	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		members, err := q.GetMembersStatuses(ctx, db.GetMembersStatusesParams{
			CompanyUuid:   company_uuid,
			AccountUuid:   initiator_account_uuid,
			AccountUuid_2: target_account_uuid,
		})
		if err != nil {
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка проверки прав",
				Data:    err,
			}
		}

		var initiatorIsOwner, targetIsOwner bool
		var foundInitiator, foundTarget bool

		for _, m := range members {
			if m.AccountUuid == initiator_account_uuid {
				initiatorIsOwner = m.IsOwner
				foundInitiator = true
			}
			if m.AccountUuid == target_account_uuid {
				targetIsOwner = m.IsOwner
				foundTarget = true
			}
		}

		if !foundInitiator {
			return &response.ApiError{
				Status:  403,
				Message: "Вы не являетесь участником компании"}
		}
		if !foundTarget {
			return &response.ApiError{
				Status:  404,
				Message: "Пользователь не найден в компании"}
		}

		if !isSelfRemoval {
			if targetIsOwner && !initiatorIsOwner {
				return &response.ApiError{
					Status:  403,
					Message: "Удалить владельца может только другой владелец",
				}
			}
		} else {
			if targetIsOwner {
				ownerCount, err := q.CountOwners(ctx, company_uuid)
				if err != nil {
					return err
				}

				if ownerCount <= 1 {
					return &response.ApiError{
						Status:  400,
						Message: "Вы последний владелец. Передайте права или удалите компанию",
					}
				}
			}
		}

		if err = s.permissionService.DeleteMemberPermissions(ctx, tx, company_uuid, target_account_uuid, initiator_account_uuid); err != nil {
			return err
		}

		err = q.DeleteMember(ctx, db.DeleteMemberParams{
			CompanyUuid: company_uuid,
			AccountUuid: target_account_uuid,
		})

		if err != nil {
			return &response.ApiError{Status: 500, Message: "Ошибка при удалении участника", Data: err}
		}

		return kafka.EmitOutbox(ctx, tx, kafka.MemberDeleted, map[string]any{
			"company_uuid": company_uuid,
			"account_uuid": target_account_uuid,
			"is_self":      isSelfRemoval,
		})
	})

	if err == nil {
		s.rdb.Del(ctx, rdb.GetMembersKey(company_uuid))
	}

	return err
}

func (s *memberService) GetMembers(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string) ([]db.GetMembersRow, error) {

	key := rdb.GetMembersKey(company_uuid)
	members_cache, err := s.rdb.Get(ctx, key).Result()

	if err == nil {
		var cachedRow []db.GetMembersRow
		if err := json.Unmarshal([]byte(members_cache), &cachedRow); err == nil {
			return cachedRow, nil
		}
	} else if err != redis.Nil {
		log.Printf("[REDIS-ERROR] %v", err)
	}

	if dbExecutor == nil {
		dbExecutor = s.pool
	}
	q := db.New(dbExecutor)

	members, err := q.GetMembers(ctx, company_uuid)

	if err != nil {
		return []db.GetMembersRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка при получении участников компании",
			Data:    err,
		}
	}

	if members == nil {
		return []db.GetMembersRow{}, nil
	}

	if jsonBytes, err := json.Marshal(members); err == nil {
		s.rdb.Set(ctx, key, jsonBytes, 10*time.Minute)
	}

	return members, nil
}

func (s *memberService) DeleteAllMembers(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string) error {

	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		account_uuids, err := q.DeleteAllMembers(ctx, company_uuid)

		if err != nil {
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка при удалении участников",
				Data:    err,
			}
		}

		payload := map[string]any{
			"company_uuid":  company_uuid,
			"account_uuids": account_uuids,
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.MemberDeleted, payload); err != nil {
			return err
		}

		return nil
	})

	key := rdb.GetMembersKey(company_uuid)
	s.rdb.Del(ctx, key)

	return err
}

func (s *memberService) GetCompanyMembershipsByAccountUUID(
	ctx context.Context,
	dbExecutor db.DBTX,
	account_uuid string) ([]db.GetCompanyMembershipsByAccountUUIDRow, error) {

	if dbExecutor == nil {
		dbExecutor = s.pool
	}
	q := db.New(dbExecutor)

	res, err := q.GetCompanyMembershipsByAccountUUID(ctx, account_uuid)

	if err != nil {
		return []db.GetCompanyMembershipsByAccountUUIDRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка получения информации о пользователе",
			Data:    err,
		}
	}

	return res, nil
}
