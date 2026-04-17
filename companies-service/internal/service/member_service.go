package service

import (
	"companies-service/internal/dto/request"
	"companies-service/internal/repository/db"
	rdb "companies-service/internal/repository/redis"
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

type MemberService struct {
	store *db.Queries
	rdb   *redis.Client
	pool  *pgxpool.Pool
}

func NewMemberService(
	store *db.Queries,
	rdb *redis.Client,
	pool *pgxpool.Pool) *MemberService {

	return &MemberService{
		store: store,
		rdb:   rdb,
		pool:  pool,
	}
}

func (s *MemberService) UpdateRoleMember(
	ctx context.Context,
	company_uuid string,
	account_uuid string,
	req request.UpdateRoleMemberRequest) error {

	rowsAffected, err := s.store.UpdateRoleMember(ctx, db.UpdateRoleMemberParams{
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

	key := rdb.GetMembersKey(company_uuid)
	s.rdb.Del(ctx, key)

	//TODO:
	// Kafka - send event

	return nil
}

func (s *MemberService) SetOwner(
	ctx context.Context,
	company_uuid string,
	account_uuid string) error {

	account_uuid_session := usercontext.GetAccountUuid(ctx)

	isOwner, err := s.store.IsCompanyOwner(ctx, db.IsCompanyOwnerParams{
		AccountUuid: account_uuid_session,
		CompanyUuid: company_uuid,
	})

	if err != nil {
		return &response.ApiError{
			Status:  500,
			Message: "Ошибка проверки прав доступа",
			Data:    err,
		}
	}

	if !isOwner {
		return &response.ApiError{
			Status:  403,
			Message: "Сделать пользователя владельцем компании может только активный владелец",
		}
	}

	rowsAffected, err := s.store.SetOwner(ctx, db.SetOwnerParams{
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

	key := rdb.GetMembersKey(company_uuid)
	s.rdb.Del(ctx, key)

	//TODO:
	// Kafka - send event

	return nil
}

func (s *MemberService) DeleteMember(
	ctx context.Context,
	company_uuid string,
	target_account_uuid string) error {

	initiator_account_uuid := usercontext.GetAccountUuid(ctx)

	if initiator_account_uuid == target_account_uuid {
		return &response.ApiError{
			Status:  403,
			Message: "Вы не можете покинуть компанию самостоятельно",
		}
	}

	members, err := s.store.GetMembersStatuses(ctx, db.GetMembersStatusesParams{
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
			Message: "Вы не являетесь участником этой компании",
		}
	}
	if !foundTarget {
		return &response.ApiError{
			Status:  404,
			Message: "Удаляемый пользователь не найден в компании",
		}
	}

	if targetIsOwner && !initiatorIsOwner {
		return &response.ApiError{
			Status:  403,
			Message: "Удалить владельца может только другой владелец",
		}
	}

	err = s.store.DeleteMember(ctx, db.DeleteMemberParams{
		CompanyUuid: company_uuid,
		AccountUuid: target_account_uuid,
	})

	if err != nil {
		return &response.ApiError{
			Status:  500,
			Message: "Ошибка при удалении участника",
			Data:    err,
		}
	}

	key := rdb.GetMembersKey(company_uuid)
	s.rdb.Del(ctx, key)

	//TODO:
	// Kafka - send event

	return nil
}

func (s *MemberService) GetMembers(
	ctx context.Context,
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

	members, err := s.store.GetMembers(ctx, company_uuid)

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
