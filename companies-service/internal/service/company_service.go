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
	pgxutil "companies-service/pkg/utils"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type CompanyMemberManager interface {
	CreateMember(ctx context.Context, dbExecutor db.DBTX, ent entity.CreateMemberEntity) error
	DeleteAllMembers(ctx context.Context, dbExecutor db.DBTX, company_uuid string) error
	IsCompanyOwner(ctx context.Context, dbExecutor db.DBTX, company_uuid string, account_uuid string) (bool, error)
	GetCompanyMembershipsByAccountUUID(ctx context.Context, dbExecutor db.DBTX, account_uuid string) ([]db.GetCompanyMembershipsByAccountUUIDRow, error)
	DeleteMember(ctx context.Context, dbExecutor db.DBTX, company_uuid string, target_account_uuid string) error
}

type CompanyRoleManager interface {
	CreateRole(ctx context.Context, dbExecutor db.DBTX, company_uuid string, req request.CreateRoleRequest) (db.CreateRoleRow, error)
	DeleteAllRoles(ctx context.Context, dbExecutor db.DBTX, company_uuid string) error
}

type CompanyPermissionManager interface {
	CreatePermissions(ctx context.Context, dbExecutor db.DBTX, company_uuid string, account_uuid string) error
	DeleteAllPermissions(ctx context.Context, dbExecutor db.DBTX, company_uuid string, account_uuid string) error
}

type сompanyService struct {
	rdb               *redis.Client
	pool              *pgxpool.Pool
	txManager         *data.TransactionManager
	memberService     CompanyMemberManager
	roleService       CompanyRoleManager
	permissionService CompanyPermissionManager
}

func NewCompanyService(
	rdb *redis.Client,
	pool *pgxpool.Pool,
	txManager *data.TransactionManager,
	memberService CompanyMemberManager,
	roleService CompanyRoleManager,
	permissionService CompanyPermissionManager) *сompanyService {

	return &сompanyService{
		rdb:               rdb,
		pool:              pool,
		txManager:         txManager,
		memberService:     memberService,
		roleService:       roleService,
		permissionService: permissionService,
	}
}

func (s *сompanyService) CreateCompany(
	ctx context.Context,
	dbExecutor db.DBTX,
	req request.CreateCompanyRequest) (db.CreateCompanyRow, error) {

	var result db.CreateCompanyRow
	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		var company_uuid string = uuid.Must(uuid.NewV7()).String()

		createdCompany, err := q.CreateCompany(ctx, db.CreateCompanyParams{
			CompanyUuid:   company_uuid,
			Name:          req.Name,
			Description:   pgxutil.TextNotValid(req.Description),
			LogoFileID:    pgxutil.ToInt(req.LogoFileID),
			BillingStatus: "free",
		})

		if err != nil {
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка создания компании",
				Data:    err,
			}
		}

		createdRole, err := s.roleService.CreateRole(ctx, tx, company_uuid, request.CreateRoleRequest{
			Name:        "Владелец",
			Description: pgxutil.Pointer("Царь всея руси"),
			Color:       "111111",
		})
		if err != nil {
			return err
		}

		account_uuid := usercontext.GetAccountUuid(ctx)

		err = s.memberService.CreateMember(ctx, tx, entity.CreateMemberEntity{
			CompanyUuid: company_uuid,
			AccountUuid: account_uuid,
			RoleID:      createdRole.RoleID,
			IsOwner:     true,
		})
		if err != nil {
			return err
		}

		if err = s.permissionService.CreatePermissions(ctx, tx, company_uuid, account_uuid); err != nil {
			return err
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.CompanyCreated, createdCompany); err != nil {
			return err
		}

		result = createdCompany
		return nil
	})

	return result, err
}

func (s *сompanyService) GetCompanies(
	ctx context.Context,
	dbExecutor db.DBTX) ([]db.GetCompaniesRow, error) {

	if dbExecutor == nil {
		dbExecutor = s.pool
	}
	q := db.New(dbExecutor)

	account_uuid := usercontext.GetAccountUuid(ctx)

	companies, err := q.GetCompanies(ctx, account_uuid)

	if err != nil {
		return nil, &response.ApiError{
			Status:  500,
			Message: "Ошибка получения списка компаний",
			Data:    err,
		}
	}

	return companies, nil
}

func (s *сompanyService) UpdateCompany(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string,
	req request.UpdateCompanyRequest) (db.UpdateCompanyRow, error) {

	var result db.UpdateCompanyRow
	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		updateCompany, err := q.UpdateCompany(ctx, db.UpdateCompanyParams{
			CompanyUuid: company_uuid,
			Name:        pgxutil.TextNotValid(req.Name),
			Description: pgxutil.TextNotValid(req.Description),
			LogoFileID:  pgxutil.ToInt(req.LogoFileID),
		})

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return &response.ApiError{
					Status:  404,
					Message: "Компания не найдена",
				}
			}
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка обновления данных компании",
				Data:    err,
			}
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.CompanyUpdated, updateCompany); err != nil {
			return err
		}

		result = updateCompany
		return nil
	})

	key := rdb.GetCompanyKey(company_uuid)
	s.rdb.Del(ctx, key)

	return result, err
}

func (s *сompanyService) GetCompanyByUuid(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string) (db.GetCompanyByUuidRow, error) {

	if dbExecutor == nil {
		dbExecutor = s.pool
	}
	q := db.New(dbExecutor)

	account_uuid := usercontext.GetAccountUuid(ctx)

	key := rdb.GetCompanyKey(company_uuid)
	company_cache, err := s.rdb.Get(ctx, key).Result()

	if err == nil {
		var cachedRow db.GetCompanyByUuidRow
		if err := json.Unmarshal([]byte(company_cache), &cachedRow); err == nil {
			return cachedRow, nil
		}
	} else if err != redis.Nil {
		log.Printf("[REDIS-ERROR] %v", err)
	}

	company, err := q.GetCompanyByUuid(ctx, db.GetCompanyByUuidParams{
		AccountUuid: account_uuid,
		CompanyUuid: company_uuid,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetCompanyByUuidRow{}, &response.ApiError{
				Status:  404,
				Message: "Компания не найдена",
			}
		}
		return db.GetCompanyByUuidRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка получения компании",
			Data:    err,
		}
	}

	if jsonBytes, err := json.Marshal(company); err == nil {
		s.rdb.Set(ctx, key, jsonBytes, 1*time.Hour)
	}

	return company, nil
}

func (s *сompanyService) DeleteCompany(
	ctx context.Context,
	dbExecutor db.DBTX,
	company_uuid string) error {

	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		account_uuid := usercontext.GetAccountUuid(ctx)

		isOwner, err := s.memberService.IsCompanyOwner(ctx, tx, company_uuid, account_uuid)
		if err != nil {
			return err
		}

		if !isOwner {
			return &response.ApiError{
				Status:  403,
				Message: "Удалить компанию может только активный владелец",
				Data:    err,
			}
		}

		if err = s.permissionService.DeleteAllPermissions(ctx, tx, company_uuid, account_uuid); err != nil {
			return err
		}

		if err = s.roleService.DeleteAllRoles(ctx, tx, company_uuid); err != nil {
			return err
		}

		if err = s.memberService.DeleteAllMembers(ctx, tx, company_uuid); err != nil {
			return err
		}

		if err = q.DeleteCompany(ctx, company_uuid); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return &response.ApiError{
					Status:  404,
					Message: "Компания не найдена",
				}
			}
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка удаления компании",
				Data:    err,
			}
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.CompanyDeleted, company_uuid); err != nil {
			return err
		}

		return nil
	})

	if err == nil {
		key := rdb.GetCompanyKey(company_uuid)
		s.rdb.Del(ctx, key)
	}

	return err
}

func (s *сompanyService) HandleGlobalAccountDeletion(
	ctx context.Context,
	dbExecutor db.DBTX,
	account_uuid string) error {

	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {

		memberships, err := s.memberService.GetCompanyMembershipsByAccountUUID(ctx, tx, account_uuid)
		if err != nil {
			return err
		}

		if len(memberships) == 0 {
			return nil
		}

		for _, m := range memberships {
			if m.IsOwner {
				if err := s.DeleteCompany(ctx, tx, m.CompanyUuid); err != nil {
					return err
				}
			} else {
				if err := s.memberService.DeleteMember(ctx, tx, m.CompanyUuid, m.AccountUuid); err != nil {
					return err
				}
			}
		}
		return nil
	})

	return err
}
