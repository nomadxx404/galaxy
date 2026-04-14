package api

import (
	"companies-service/internal/dto/request"
	"companies-service/internal/service"
	"companies-service/pkg/middleware"
	"companies-service/pkg/response"

	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	service *service.MemberService
}

func NewMemberHandler(s *service.MemberService) *MemberHandler {
	return &MemberHandler{
		service: s,
	}
}

func (h *MemberHandler) UpdateRoleMember(c *gin.Context) {
	company_uuid := c.Param("company_uuid")
	if company_uuid == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	account_uuid := c.Param("account_uuid")
	if account_uuid == "" {
		response.SendFailure(c, 400, "ID аккаунта не указан")
		return
	}

	req, ok := middleware.BindJSON[request.UpdateRoleMemberRequest](c)
	if !ok {
		return
	}

	err := h.service.UpdateRoleMember(c.Request.Context(), company_uuid, account_uuid, *req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, nil, "Роль обновлена")
}

func (h *MemberHandler) SetOwner(c *gin.Context) {
	company_uuid := c.Param("company_uuid")
	if company_uuid == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	account_uuid := c.Param("account_uuid")
	if account_uuid == "" {
		response.SendFailure(c, 400, "ID аккаунта не указан")
		return
	}

	err := h.service.SetOwner(c.Request.Context(), company_uuid, account_uuid)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, nil, "Роль обновлена")
}

func (h *MemberHandler) DeleteMember(c *gin.Context) {
	company_uuid := c.Param("company_uuid")
	if company_uuid == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	account_uuid := c.Param("account_uuid")
	if account_uuid == "" {
		response.SendFailure(c, 400, "ID аккаунта не указан")
		return
	}

	err := h.service.DeleteMember(c.Request.Context(), company_uuid, account_uuid)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, nil, "Участник удален успешно")
}

func (h *MemberHandler) GetMembers(c *gin.Context) {
	company_uuid := c.Param("company_uuid")
	if company_uuid == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	members, err := h.service.GetMembers(c.Request.Context(), company_uuid)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, members, "Успешно")
}
