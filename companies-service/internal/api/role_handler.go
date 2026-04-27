package api

import (
	"companies-service/internal/dto/request"
	"companies-service/internal/repository/db"
	_ "companies-service/internal/repository/db"
	"companies-service/pkg/middleware"
	"companies-service/pkg/response"
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoleManager interface {
	CreateRole(ctx context.Context, dbExecutor db.DBTX, company_uuid string, req request.CreateRoleRequest) (db.CreateRoleRow, error)
	GetRoles(ctx context.Context, dbExecutor db.DBTX, company_uuid string) ([]db.GetRolesRow, error)
	GetRoleByUuid(ctx context.Context, dbExecutor db.DBTX, company_uuid string, role_id int32) (db.GetRoleByUuidRow, error)
	UpdateRole(ctx context.Context, dbExecutor db.DBTX, company_uuid string, role_id int32, req request.UpdateRoleRequest) (db.UpdateRoleRow, error)
	DeleteRole(ctx context.Context, dbExecutor db.DBTX, company_uuid string, role_id int32) error
}

type RoleHandler struct {
	service RoleManager
}

func NewRoleHandler(s RoleManager) *RoleHandler {
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

	res, err := h.service.CreateRole(c.Request.Context(), nil, company_uuid, *req)

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

	res, err := h.service.GetRoles(c.Request.Context(), nil, company_uuid)

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

	res, err := h.service.GetRoleByUuid(c.Request.Context(), nil, company_uuid, role_id)

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

	res, err := h.service.UpdateRole(c.Request.Context(), nil, company_uuid, role_id, *req)

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

	err = h.service.DeleteRole(c.Request.Context(), nil, company_uuid, int32(role_id))

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, nil, "Роль удалена")
}
