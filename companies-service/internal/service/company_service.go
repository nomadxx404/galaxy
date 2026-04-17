package service

import (
	"companies-service/internal/dto/request"
	"companies-service/internal/permissions"
	"companies-service/internal/repository/db"
	rdb "companies-service/internal/repository/redis"
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

type CompanyService struct {
	store *db.Queries
	rdb   *redis.Client
	pool  *pgxpool.Pool
}

func NewCompanyService(
	store *db.Queries,
	rdb *redis.Client,
	pool *pgxpool.Pool) *CompanyService {

	return &CompanyService{
		store: store,
		rdb:   rdb,
		pool:  pool,
	}
}

func (s *CompanyService) CreateCompany(
	ctx context.Context,
	req request.CreateCompanyRequest) (db.CreateCompanyRow, error) {

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.CreateCompanyRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка создания транзакции",
			Data:    err,
		}
	}
	defer tx.Rollback(ctx)

	qTx := s.store.WithTx(tx)
	var company_uuid string = uuid.Must(uuid.NewV7()).String()

	company, err := qTx.CreateCompany(ctx, db.CreateCompanyParams{
		CompanyUuid:   company_uuid,
		Name:          req.Name,
		Description:   pgxutil.TextNotValid(req.Description),
		LogoFileID:    pgxutil.ToInt(req.LogoFileID),
		BillingStatus: "free",
	})
	if err != nil {
		return db.CreateCompanyRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка создания компании",
			Data:    err,
		}
	}

	role, err := qTx.CreateRole(ctx, db.CreateRoleParams{
		Name:        "Владелец",
		Description: pgxutil.TextValid("Царь всея руси"),
		Color:       "111111",
		CompanyUuid: company_uuid,
	})

	if err != nil {
		return db.CreateCompanyRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка создания роли",
			Data:    err,
		}
	}

	account_uuid := usercontext.GetAccountUuid(ctx)

	err = qTx.CreateMember(ctx, db.CreateMemberParams{
		CompanyUuid: company_uuid,
		AccountUuid: account_uuid,
		RoleID:      role.RoleID,
		IsOwner:     true,
	})
	if err != nil {
		return db.CreateCompanyRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка добавления участника",
			Data:    err,
		}
	}

	for _, domain := range permissions.AllDomainPermission {
		err = qTx.GiveAccess(ctx, db.GiveAccessParams{
			CompanyUuid: company_uuid,
			AccountUuid: account_uuid,
			Domain:      domain,
			Mask:        -1,
			ChangeBy:    account_uuid,
		})
		if err != nil {
			return db.CreateCompanyRow{}, &response.ApiError{
				Status:  500,
				Message: "Ошибка выдачи прав доступа",
				Data:    err,
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return db.CreateCompanyRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка сохранения транзакции",
			Data:    err,
		}
	}

	return company, nil
}

func (s *CompanyService) GetCompanies(ctx context.Context) ([]db.GetCompaniesRow, error) {
	account_uuid := usercontext.GetAccountUuid(ctx)

	companies, err := s.store.GetCompanies(ctx, account_uuid)

	if err != nil {
		return nil, &response.ApiError{
			Status:  500,
			Message: "Ошибка получения списка компаний",
			Data:    err,
		}
	}

	return companies, nil
}

func (s *CompanyService) UpdateCompany(
	ctx context.Context,
	company_uuid string,
	req request.UpdateCompanyRequest) (db.UpdateCompanyRow, error) {

	updated_company, err := s.store.UpdateCompany(ctx, db.UpdateCompanyParams{
		CompanyUuid: company_uuid,
		Name:        pgxutil.TextNotValid(req.Name),
		Description: pgxutil.TextNotValid(req.Description),
		LogoFileID:  pgxutil.ToInt(req.LogoFileID),
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UpdateCompanyRow{}, &response.ApiError{
				Status:  404,
				Message: "Компания не найдена",
			}
		}
		return db.UpdateCompanyRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка обновления данных компании",
			Data:    err,
		}
	}

	key := rdb.GetCompanyKey(company_uuid)
	s.rdb.Del(ctx, key)

	//TODO:
	// Kafka - send event

	return updated_company, nil
}

func (s *CompanyService) GetCompanyByUuid(
	ctx context.Context,
	company_uuid string) (db.GetCompanyByUuidRow, error) {

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

	company, err := s.store.GetCompanyByUuid(ctx, db.GetCompanyByUuidParams{
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

func (s *CompanyService) DeleteCompany(
	ctx context.Context,
	company_uuid string) error {

	account_uuid := usercontext.GetAccountUuid(ctx)

	isOwner, err := s.store.IsCompanyOwner(ctx, db.IsCompanyOwnerParams{
		AccountUuid: account_uuid,
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
			Message: "Удалить компанию может только активный владелец",
			Data:    err,
		}
	}

	err = s.store.DeleteCompany(ctx, company_uuid)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &response.ApiError{
				Status:  404,
				Message: "Компания не найдена",
			}
		}
		return &response.ApiError{
			Status:  500,
			Message: "Ошибка получения компании",
			Data:    err,
		}
	}

	key := rdb.GetCompanyKey(company_uuid)
	s.rdb.Del(ctx, key)

	//TODO:
	// Kafka - send event

	return err
}
