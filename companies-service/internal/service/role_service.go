package service

import (
	"companies-service/internal/dto/request"
	"companies-service/internal/repository/db"
	rdb "companies-service/internal/repository/redis"
	"companies-service/pkg/response"
	pgxutil "companies-service/pkg/utils"
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type RoleService struct {
	store *db.Queries
	rdb   *redis.Client
	pool  *pgxpool.Pool
}

func NewRoleService(
	store *db.Queries,
	rdb *redis.Client,
	pool *pgxpool.Pool) *RoleService {

	return &RoleService{
		store: store,
		rdb:   rdb,
		pool:  pool,
	}
}

func (s *RoleService) CreateRole(
	ctx context.Context,
	company_uuid string,
	req request.CreateRoleRequest) (db.CreateRoleRow, error) {

	isExistsRole, err := s.store.IsExistsRole(ctx, db.IsExistsRoleParams{
		CompanyUuid: company_uuid,
		Name:        req.Name,
	})

	if err != nil {
		return db.CreateRoleRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка проверки существования роли",
			Data:    err,
		}
	}

	if isExistsRole != false {
		return db.CreateRoleRow{}, &response.ApiError{
			Status:  409,
			Message: "Такая роль в компании уже существует",
		}
	}

	new_role, err := s.store.CreateRole(ctx, db.CreateRoleParams{
		Name:        req.Name,
		Description: pgxutil.TextNotValid(req.Description),
		Color:       req.Color,
		CompanyUuid: company_uuid,
	})

	if err != nil {
		return db.CreateRoleRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка создания роли",
			Data:    err,
		}
	}

	return new_role, nil
}

func (s *RoleService) GetRoles(
	ctx context.Context,
	company_uuid string) ([]db.GetRolesRow, error) {

	getRoles, err := s.store.GetRoles(ctx, company_uuid)

	if err != nil {
		return []db.GetRolesRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка пролучения списка ролей",
			Data:    err,
		}
	}

	return getRoles, nil
}

func (s *RoleService) GetRoleByUuid(
	ctx context.Context,
	company_uuid string,
	role_id int32) (db.GetRoleByUuidRow, error) {

	getRole, err := s.store.GetRoleByUuid(ctx, db.GetRoleByUuidParams{
		CompanyUuid: company_uuid,
		RoleID:      role_id,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetRoleByUuidRow{}, &response.ApiError{
				Status:  404,
				Message: "Роль не найдена",
			}
		}
		return db.GetRoleByUuidRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка пролучения роли",
			Data:    err,
		}
	}

	return getRole, nil
}

func (s *RoleService) UpdateRole(
	ctx context.Context,
	company_uuid string,
	role_id int32,
	req request.UpdateRoleRequest) (db.UpdateRoleRow, error) {

	updateRoles, err := s.store.UpdateRole(ctx, db.UpdateRoleParams{
		CompanyUuid: company_uuid,
		RoleID:      role_id,
		Name:        pgxutil.TextNotValid(req.Name),
		Description: pgxutil.TextNotValid(req.Description),
		Color:       pgxutil.TextNotValid(req.Color),
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UpdateRoleRow{}, &response.ApiError{
				Status:  404,
				Message: "Роль не найдена",
			}
		}
		return db.UpdateRoleRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка обновления роли",
			Data:    err,
		}
	}

	key := rdb.GetMembersKey(company_uuid)
	s.rdb.Del(ctx, key)

	//TODO:
	// Kafka - send event

	return updateRoles, nil
}

func (s *RoleService) DeleteRole(
	ctx context.Context,
	company_uuid string,
	role_id int32) error {

	err := s.store.DeleteRole(ctx, db.DeleteRoleParams{
		CompanyUuid: company_uuid,
		RoleID:      role_id,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &response.ApiError{
				Status:  404,
				Message: "Роль не найдена",
			}
		}
		return &response.ApiError{
			Status:  500,
			Message: "Ошибка удаления роли",
			Data:    err,
		}
	}

	//TODO:
	// Kafka - send event

	return err
}
