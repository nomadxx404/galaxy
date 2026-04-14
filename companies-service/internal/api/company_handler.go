package api

import (
	"companies-service/internal/dto/request"
	_ "companies-service/internal/repository/db"
	"companies-service/internal/service"
	"companies-service/pkg/response"

	"github.com/gin-gonic/gin"
)

type CompanyHandler struct {
	service *service.CompanyService
}

func NewCompanyHandler(s *service.CompanyService) *CompanyHandler {
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

	res, err := h.service.CreateCompany(c.Request.Context(), req)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 201, res, "Компания успешно создана")
}

func (h *CompanyHandler) GetCompanies(c *gin.Context) {

	res, err := h.service.GetCompanies(c.Request.Context())

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
	// if company_uuid == "" {
	// 	response.SendFailure(c, 400, "ID компании не указан")
	// 	return
	// }

	res, err := h.service.UpdateCompany(c.Request.Context(), company_uuid, req)

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

	res, err := h.service.GetCompanyByUuid(c.Request.Context(), company_uuid)

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

	err := h.service.DeleteCompany(c.Request.Context(), company_uuid)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, nil, "Компания удалена")
}
