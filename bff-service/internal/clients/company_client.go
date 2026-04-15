package clients

import (
	"bff-service/internal/dto/returning"
	"bff-service/pkg/response"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type CompanyClient struct {
	BaseURL string
	Client  *http.Client
}

func (c *CompanyClient) GetMembers(ctx context.Context, companyUUID string) ([]returning.GetMembersResponse, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/companies/%s/members", c.BaseURL, companyUUID), nil)

	resp, err := c.Client.Do(req)
	if err != nil {
		return []returning.GetMembersResponse{}, &response.ApiError{
			Status:  503,
			Message: "Ошибка при получении участников компании",
			Data:    err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []returning.GetMembersResponse{}, &response.ApiError{
			Status:  resp.StatusCode,
			Message: "Сервис company ответил ошибкой",
		}
	}

	var wrapper struct {
		Success bool                           `json:"success"`
		Data    []returning.GetMembersResponse `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, &response.ApiError{
			Status:  500,
			Message: "Ошибка десериализации ответа от сервиса company",
			Data:    err,
		}
	}

	return wrapper.Data, nil
}
