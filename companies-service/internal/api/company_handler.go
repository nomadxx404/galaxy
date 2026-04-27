package api

import (
	"companies-service/internal/dto/request"
	"companies-service/internal/repository/db"
	_ "companies-service/internal/repository/db"
	"companies-service/pkg/response"
	"context"

	"github.com/gin-gonic/gin"
)

type CompanyManager interface {
	CreateCompany(ctx context.Context, dbExecutor db.DBTX, req request.CreateCompanyRequest) (db.CreateCompanyRow, error)
	GetCompanies(ctx context.Context, dbExecutor db.DBTX) ([]db.GetCompaniesRow, error)
	UpdateCompany(ctx context.Context, dbExecutor db.DBTX, company_uuid string, req request.UpdateCompanyRequest) (db.UpdateCompanyRow, error)
	GetCompanyByUuid(ctx context.Context, dbExecutor db.DBTX, company_uuid string) (db.GetCompanyByUuidRow, error)
	DeleteCompany(ctx context.Context, dbExecutor db.DBTX, company_uuid string) error
}

type CompanyHandler struct {
	service CompanyManager
}

func NewCompanyHandler(s CompanyManager) *CompanyHandler {
	return &CompanyHandler{
		service: s,
	}
}

func (h *CompanyHandler) CreateCompany(c *gin.Context) {
	var req request.CreateCompanyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.SendFailure(c, 400, "Некорректный формат данных")
		return
	}

	res, err := h.service.CreateCompany(c.Request.Context(), nil, req)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 201, res, "Компания успешно создана")
}

func (h *CompanyHandler) GetCompanies(c *gin.Context) {

	res, err := h.service.GetCompanies(c.Request.Context(), nil)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, res, "Список получен")
}

func (h *CompanyHandler) UpdateCompany(c *gin.Context) {
	var req request.UpdateCompanyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	company_uuid := c.Param("company_uuid")
	if company_uuid == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	res, err := h.service.UpdateCompany(c.Request.Context(), nil, company_uuid, req)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, res, "Данные обновлены")
}

func (h *CompanyHandler) GetCompanyByUuid(c *gin.Context) {

	company_uuid := c.Param("company_uuid")
	if company_uuid == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	res, err := h.service.GetCompanyByUuid(c.Request.Context(), nil, company_uuid)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, res, "Успех")
}

func (h *CompanyHandler) DeleteCompany(c *gin.Context) {

	company_uuid := c.Param("company_uuid")
	if company_uuid == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	err := h.service.DeleteCompany(c.Request.Context(), nil, company_uuid)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, nil, "Компания удалена")
}
