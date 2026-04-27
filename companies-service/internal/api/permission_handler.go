package api

import (
	"companies-service/internal/dto/request"
	"companies-service/internal/permissions"
	"companies-service/internal/repository/db"
	"companies-service/pkg/response"
	"context"

	"github.com/gin-gonic/gin"
)

type PermissionManager interface {
	UpdatePermissions(ctx context.Context, dbExecutor db.DBTX, company_uuid string, account_uuid string, req request.UpdatePermissionsRequest) error
	GetAccountPermission(ctx context.Context, dbExecutor db.DBTX, company_uuid, account_uuid string) ([]db.GetUserAllPermissionsRow, error)
}

type PermissionHandler struct {
	service PermissionManager
}

func NewPermissionHandler(s PermissionManager) *PermissionHandler {
	return &PermissionHandler{
		service: s,
	}
}

func (h *PermissionHandler) GetPermission(c *gin.Context) {
	response.SendSuccess(c, 200, permissions.AllPermissions, "Успешно")
}

func (h *PermissionHandler) UpdatePermissions(c *gin.Context) {

	company_uuid := c.Param("company_uuid")
	account_uuid := c.Param("account_uuid")

	var req request.UpdatePermissionsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.SendFailure(c, 400, "Неверный формат данных")
		return
	}

	err := h.service.UpdatePermissions(c.Request.Context(), nil, company_uuid, account_uuid, req)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, nil, "Права успешно обновлены")
}

func (h *PermissionHandler) GetAccountPermission(c *gin.Context) {
	company_uuid := c.Param("company_uuid")
	account_uuid := c.Param("account_uuid")

	res, err := h.service.GetAccountPermission(c.Request.Context(), nil, company_uuid, account_uuid)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, res, "Успешно")
}
