package api

import (
	"companies-service/internal/dto/request"
	"companies-service/internal/permissions"
	"companies-service/internal/service"
	"companies-service/pkg/response"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	service *service.PermissionService
}

func NewPermissionHandler(s *service.PermissionService) *PermissionHandler {
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

	err := h.service.UpdatePermissions(c.Request.Context(), company_uuid, account_uuid, req)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, nil, "Права успешно обновлены")
}

func (h *PermissionHandler) GetAccountPermission(c *gin.Context) {
	company_uuid := c.Param("company_uuid")
	account_uuid := c.Param("account_uuid")

	res, err := h.service.GetAccountPermission(c.Request.Context(), company_uuid, account_uuid)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, res, "Успешно")
}
