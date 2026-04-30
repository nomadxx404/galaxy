package service

import (
	"companies-service/internal/dto/entity"
	"companies-service/internal/dto/request"
	"companies-service/internal/repository/db"
	rdb "companies-service/internal/repository/redis"
	"companies-service/pkg/data"
	"companies-service/pkg/kafka"
	"companies-service/pkg/response"
	pgxutil "companies-service/pkg/utils"
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type roleService struct {
	rdb       *redis.Client
	pool      *pgxpool.Pool
	txManager *data.TransactionManager
}

func NewRoleService(
	rdb *redis.Client,
	pool *pgxpool.Pool,
	txManager *data.TransactionManager) *roleService {

	return &roleService{
		rdb:       rdb,
		pool:      pool,
		txManager: txManager,
	}
}

func (s *roleService) CreateRole(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string,
	req request.CreateRoleRequest) (db.CreateRoleRow, error) {

	var result db.CreateRoleRow
	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		isExistsRole, err := q.IsExistsRole(ctx, db.IsExistsRoleParams{
			CompanyUuid: company_uuid,
			Name:        req.Name,
		})

		if err != nil {
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка проверки существования роли",
				Data:    err,
			}
		}

		if isExistsRole != false {
			return &response.ApiError{
				Status:  409,
				Message: "Такая роль в компании уже существует",
			}
		}

		createdRole, err := q.CreateRole(ctx, db.CreateRoleParams{
			Name:        req.Name,
			Description: pgxutil.TextNotValid(req.Description),
			Color:       req.Color,
			CompanyUuid: company_uuid,
		})

		if err != nil {
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка создания роли",
				Data:    err,
			}
		}

		payload := map[string]any{
			"company_uuid": company_uuid,
			"new_role":     createdRole,
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.RoleCreated, payload); err != nil {
			return err
		}
		result = createdRole
		return nil
	})

	return result, err
}

func (s *roleService) GetRoles(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string) ([]db.GetRolesRow, error) {

	if dbExecutor == nil {
		dbExecutor = s.pool
	}
	q := db.New(dbExecutor)

	getRoles, err := q.GetRoles(ctx, company_uuid)

	if err != nil {
		return []db.GetRolesRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка пролучения списка ролей",
			Data:    err,
		}
	}

	return getRoles, nil
}

func (s *roleService) GetRoleByUuid(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string,
	role_id int32) (db.GetRoleByUuidRow, error) {

	if dbExecutor == nil {
		dbExecutor = s.pool
	}
	q := db.New(dbExecutor)

	getRole, err := q.GetRoleByUuid(ctx, db.GetRoleByUuidParams{
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

func (s *roleService) UpdateRole(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string,
	role_id int32,
	req request.UpdateRoleRequest) (db.UpdateRoleRow, error) {

	var result db.UpdateRoleRow
	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		updatedRole, err := q.UpdateRole(ctx, db.UpdateRoleParams{
			CompanyUuid: company_uuid,
			RoleID:      role_id,
			Name:        pgxutil.TextNotValid(req.Name),
			Description: pgxutil.TextNotValid(req.Description),
			Color:       pgxutil.TextNotValid(req.Color),
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
				Message: "Ошибка обновления роли",
				Data:    err,
			}
		}

		payload := map[string]any{
			"company_uuid": company_uuid,
			"updateRoles":  updatedRole,
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.RoleUpdated, payload); err != nil {
			return err
		}
		result = updatedRole
		return nil
	})

	key := rdb.GetMembersKey(company_uuid)
	s.rdb.Del(ctx, key)

	return result, err
}

func (s *roleService) DeleteRole(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string,
	role_id int32) error {

	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		err := q.DeleteRole(ctx, db.DeleteRoleParams{
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

		payload := map[string]any{
			"company_uuid": company_uuid,
			"role_id":      role_id,
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.RoleDeleted, payload); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (s *roleService) DeleteAllRoles(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string) error {

	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		role_ids, err := q.DeleteAllRoles(ctx, company_uuid)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return &response.ApiError{
					Status:  404,
					Message: "Роли не найдена",
				}
			}
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка удаления ролей",
				Data:    err,
			}
		}

		payload := map[string]any{
			"company_uuid": company_uuid,
			"role_uuids":   role_ids,
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.RoleDeleted, payload); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (s *roleService) GetRoleIdByName(
	ctx context.Context,
	dbExecutor db.DBTX,
	ent entity.GetRoleIdByName) (int32, error) {

	if dbExecutor == nil {
		dbExecutor = s.pool
	}
	q := db.New(dbExecutor)

	role_id, err := q.GetRoleIdByName(ctx, db.GetRoleIdByNameParams{
		CompanyUuid: ent.CompanyUuid,
		Name:        ent.Name,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, &response.ApiError{
				Status:  404,
				Message: "Роль не найдена",
			}
		}
		return 0, &response.ApiError{
			Status:  500,
			Message: "Ошибка пролучения роли",
			Data:    err,
		}
	}

	return role_id, nil
}
