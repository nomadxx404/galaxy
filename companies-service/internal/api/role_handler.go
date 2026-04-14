package api

import (
	"companies-service/internal/dto/request"
	_ "companies-service/internal/repository/db"
	"companies-service/internal/service"
	"companies-service/pkg/middleware"
	"companies-service/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	service *service.RoleService
}

func NewRoleHandler(s *service.RoleService) *RoleHandler {
	return &RoleHandler{
		service: s,
	}
}

func (h *RoleHandler) CreateRole(c *gin.Context) {

	company_uuid := c.Param("company_uuid")
	if company_uuid == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	req, ok := middleware.BindJSON[request.CreateRoleRequest](c)
	if !ok {
		return
	}

	res, err := h.service.CreateRole(c.Request.Context(), company_uuid, *req)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 201, res, "Роль успешно создана")
}

func (h *RoleHandler) GetRoles(c *gin.Context) {

	company_uuid := c.Param("company_uuid")
	if company_uuid == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	res, err := h.service.GetRoles(c.Request.Context(), company_uuid)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, res, "Успешно")
}

func (h *RoleHandler) GetRoleByUuid(c *gin.Context) {

	company_uuid := c.Param("company_uuid")
	if company_uuid == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	role_id, err := middleware.GetParamInt32(c, "role_id")
	if err != nil {
		response.SendFailure(c, 400, "ID роли не указан")
		return
	}

	res, err := h.service.GetRoleByUuid(c.Request.Context(), company_uuid, role_id)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, res, "Успешно")
}

func (h *RoleHandler) UpdateRole(c *gin.Context) {

	company_uuid := c.Param("company_uuid")
	if company_uuid == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	role_id, err := middleware.GetParamInt32(c, "role_id")
	if err != nil {
		response.SendFailure(c, 400, "ID роли не указан")
		return
	}

	req, ok := middleware.BindJSON[request.UpdateRoleRequest](c)
	if !ok {
		return
	}

	res, err := h.service.UpdateRole(c.Request.Context(), company_uuid, role_id, *req)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, res, "Данные обновлены")
}

func (h *RoleHandler) DeleteRole(c *gin.Context) {

	company_uuid := c.Param("company_uuid")
	role_id_header := c.Param("role_id")
	if company_uuid == "" || role_id_header == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	role_id, err := strconv.Atoi(role_id_header)

	err = h.service.DeleteRole(c.Request.Context(), company_uuid, int32(role_id))

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, nil, "Роль удалена")
}
