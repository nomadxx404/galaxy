package service

import (
	"companies-service/internal/dto/request"
	"companies-service/internal/repository/db"
	rdb "companies-service/internal/repository/redis"
	"companies-service/pkg/response"
	usercontext "companies-service/pkg/user_context"
	pgxutil "companies-service/pkg/utils"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type InvitationService struct {
	store *db.Queries
	rdb   *redis.Client
	pool  *pgxpool.Pool
}

func NewInvitationService(
	store *db.Queries,
	rdb *redis.Client,
	pool *pgxpool.Pool) *InvitationService {

	return &InvitationService{
		store: store,
		rdb:   rdb,
		pool:  pool,
	}
}

func (s *InvitationService) CreateInvitation(
	ctx context.Context,
	company_uuid string,
	req request.CreateInvitationRequest) error {

	token, err := GenerateSecureToken(32)

	if err != nil {
		return &response.ApiError{
			Status:  500,
			Message: "Ошибка генерации токена",
			Data:    err}
	}

	key := rdb.GetInvitationKey(token)
	err = s.rdb.Set(ctx, key, company_uuid, 10*time.Minute).Err()

	if err != nil {
		return &response.ApiError{
			Status:  500,
			Message: "Ошибка сохранения токена",
			Data:    err,
		}
	}

	//TODO:
	//Kafka - create event send token email

	return nil
}

func (s *InvitationService) AcceptInvitation(
	ctx context.Context,
	token string) error {

	key := rdb.GetInvitationKey(token)
	company_uuid, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return &response.ApiError{
			Status:  410,
			Message: "Срок действия приглашения истек",
		}
	}

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

	var roleID int32
	role_id, err := s.store.GetRoleByName(ctx, db.GetRoleByNameParams{
		CompanyUuid: company_uuid,
		Name:        "Приглашенный",
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			newRole, createErr := qTx.CreateRole(ctx, db.CreateRoleParams{
				CompanyUuid: company_uuid,
				Name:        "Приглашенный",
				Color:       "123456",
				Description: pgxutil.TextValid("Стал участником по приглашению"),
			})
			if createErr != nil {
				return &response.ApiError{
					Status:  500,
					Message: "Не удалось создать роль по умолчанию",
					Data:    err,
				}
			}
			roleID = newRole.RoleID
		} else {
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка при поиске роли",
				Data:    err,
			}
		}
	} else {
		roleID = role_id
	}

	account_uuid, err := usercontext.GetAccountUuid(ctx)
	err = qTx.CreateMember(ctx, db.CreateMemberParams{
		CompanyUuid: company_uuid,
		AccountUuid: account_uuid,
		RoleID:      roleID,
		IsOwner:     false,
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
			Message: "Ошибка при вступлении в компанию",
			Data:    err,
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return &response.ApiError{
			Status:  500,
			Message: "Ошибка сохранения транзакции",
			Data:    err,
		}
	}

	s.rdb.Del(ctx, key)

	//TODO:
	//Kafka - create event send token email

	return nil
}

func GenerateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
