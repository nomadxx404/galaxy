package api

import (
	"companies-service/internal/dto/request"
	"companies-service/internal/repository/db"
	"companies-service/pkg/middleware"
	"companies-service/pkg/response"
	"context"

	"github.com/gin-gonic/gin"
)

type InvitationManager interface {
	CreateInvitation(ctx context.Context, dbExecutor db.DBTX, company_uuid string, req request.CreateInvitationRequest) error
	AcceptInvitation(ctx context.Context, dbExecutor db.DBTX, token string) error
}

type InvitationHandler struct {
	service InvitationManager
}

func NewInvitationHandler(s InvitationManager) *InvitationHandler {
	return &InvitationHandler{
		service: s,
	}
}

func (h *InvitationHandler) CreateInvitation(c *gin.Context) {
	company_uuid := c.Param("company_uuid")
	if company_uuid == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	req, ok := middleware.BindJSON[request.CreateInvitationRequest](c)
	if !ok {
		return
	}

	err := h.service.CreateInvitation(c.Request.Context(), nil, company_uuid, *req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 201, nil, "Приглашение создано")
}

func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	tokenUrl := c.Param("tokenUrl")
	if tokenUrl == "" {
		response.SendFailure(c, 400, "Токен не указан")
		return
	}

	err := h.service.AcceptInvitation(c.Request.Context(), nil, tokenUrl)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, nil, "Вы стали членом команды")
}
