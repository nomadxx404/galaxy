package api

import (
	"companies-service/internal/dto/request"
	"companies-service/internal/service"
	"companies-service/pkg/middleware"
	"companies-service/pkg/response"

	"github.com/gin-gonic/gin"
)

type InvitationHandler struct {
	service *service.InvitationService
}

func NewInvitationHandler(s *service.InvitationService) *InvitationHandler {
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

	err := h.service.CreateInvitation(c.Request.Context(), company_uuid, *req)
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

	err := h.service.AcceptInvitation(c.Request.Context(), tokenUrl)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, nil, "Вы стали членом команды")
}
