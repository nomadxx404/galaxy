package api

import (
	"bff-service/internal/service"
	"bff-service/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	aggregator *service.AggregatorService
}

func NewHandler(agg *service.AggregatorService) *Handler {
	return &Handler{
		aggregator: agg,
	}
}

func (h *Handler) GetCompanyMembers(c *gin.Context) {
	company_uuid := c.Param("company_uuid")

	if company_uuid == "" {
		response.SendFailure(c, 400, "ID компании не указан")
		return
	}

	members, err := h.aggregator.GetCompanyMembers(c.Request.Context(), company_uuid)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, members, "Успешно")
}
