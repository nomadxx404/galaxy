package clients

import (
	"bff-service/internal/dto/request"
	"bff-service/internal/dto/returning"
	"bff-service/pkg/response"
	"bytes"

	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type AuthClient struct {
	BaseURL string
	Client  *http.Client
}

func (c *AuthClient) GetProfilesBatch(ctx context.Context, uuids []string) ([]returning.AccountBatchResponse, error) {
	reqBody := request.AccountBatchRequest{
		AccountUuids: uuids,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, &response.ApiError{
			Status:  500,
			Message: "Ошибка подготовки данных для Auth сервиса",
			Data:    err,
		}
	}

	url := fmt.Sprintf("%s/api/profile/batch", c.BaseURL)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, &response.ApiError{
			Status:  500,
			Message: "Сервис auth недоступен",
			Data:    err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &response.ApiError{
			Status:  resp.StatusCode,
			Message: "Auth сервис вернул ошибку",
		}
	}

	var wrapper struct {
		Success bool                             `json:"success"`
		Data    []returning.AccountBatchResponse `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, &response.ApiError{
			Status:  500,
			Message: "Ошибка обработки данных профилей (десериализация)",
			Data:    err,
		}
	}

	return wrapper.Data, nil

}
