package service

import (
	"companies-service/internal/dto/entity"
	"companies-service/internal/dto/request"
	"companies-service/internal/repository/db"
	rdb "companies-service/internal/repository/redis"
	"companies-service/pkg/config"
	"companies-service/pkg/data"
	"companies-service/pkg/kafka"
	"companies-service/pkg/response"
	usercontext "companies-service/pkg/user_context"
	pgxutil "companies-service/pkg/utils"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type InvitationMemberManager interface {
	CreateMember(ctx context.Context, dbExecutor db.DBTX, ent entity.CreateMemberEntity) error
}

type InvitationRoleManager interface {
	CreateRole(ctx context.Context, dbExecutor db.DBTX, company_uuid string, req request.CreateRoleRequest) (db.CreateRoleRow, error)
	GetRoleIdByName(ctx context.Context, dbExecutor db.DBTX, ent entity.GetRoleIdByName) (int32, error)
}

type invitationService struct {
	rdb           *redis.Client
	pool          *pgxpool.Pool
	txManager     *data.TransactionManager
	cfg           *config.Config
	roleService   InvitationRoleManager
	memberService InvitationMemberManager
}

func NewInvitationService(
	rdb *redis.Client,
	pool *pgxpool.Pool,
	txManager *data.TransactionManager,
	cfg *config.Config,
	roleService InvitationRoleManager,
	memberService InvitationMemberManager) *invitationService {

	return &invitationService{
		rdb:           rdb,
		pool:          pool,
		txManager:     txManager,
		cfg:           cfg,
		roleService:   roleService,
		memberService: memberService,
	}
}

func (s *invitationService) CreateInvitation(
	ctx context.Context,
	dbExecutor db.DBTX,
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
	err = s.rdb.Set(ctx, key, company_uuid, time.Duration(s.cfg.Invitation.INVITATION_LINK_ACCESS_MINUTES)*time.Minute).Err()

	if err != nil {
		return &response.ApiError{
			Status:  500,
			Message: "Ошибка сохранения токена",
			Data:    err,
		}
	}

	payload := map[string]any{
		"company_uuid": company_uuid,
		"email":        req.Email,
		"token":        token,
	}

	err = s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		if err = kafka.EmitOutbox(ctx, tx, kafka.InvitationCreated, payload); err != nil {
			return err
		}
		return nil
	})

	return err
}

func (s *invitationService) AcceptInvitation(
	ctx context.Context,
	dbExecutor db.DBTX,
	token string) error {

	keyToken := rdb.GetInvitationKey(token)
	company_uuid, err := s.rdb.Get(ctx, keyToken).Result()
	if err == redis.Nil {
		return &response.ApiError{
			Status:  410,
			Message: "Срок действия приглашения истек",
		}
	}

	err = s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {

		var roleID int32
		role_id, err := s.roleService.GetRoleIdByName(ctx, tx, entity.GetRoleIdByName{
			CompanyUuid: company_uuid,
			Name:        "Приглашенный",
		})

		if err != nil {
			var apiErr *response.ApiError
			if errors.As(err, &apiErr) && apiErr.Status == 404 {
				createdRole, errRole := s.roleService.CreateRole(ctx, tx, company_uuid, request.CreateRoleRequest{
					Name:        "Приглашенный",
					Color:       "123456",
					Description: pgxutil.Pointer("Стал участником по приглашению"),
				})
				if errRole != nil {
					return err
				}
				roleID = createdRole.RoleID
			} else {
				return err
			}
		} else {
			roleID = role_id
		}

		account_uuid := usercontext.GetAccountUuid(ctx)

		err = s.memberService.CreateMember(ctx, tx, entity.CreateMemberEntity{
			CompanyUuid: company_uuid,
			AccountUuid: account_uuid,
			RoleID:      roleID,
			IsOwner:     false,
		})
		if err != nil {
			return err
		}

		payload := map[string]any{
			"company_uuid": company_uuid,
			"account_uuid": account_uuid,
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.MemberJoined, payload); err != nil {
			return err
		}

		return nil
	})

	if err == nil {
		s.rdb.Del(ctx, keyToken, rdb.GetMembersKey(company_uuid))
	}

	return err
}

func GenerateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
