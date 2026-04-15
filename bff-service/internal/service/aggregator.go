package service

import (
	"bff-service/internal/clients"
	"bff-service/internal/dto/returning"
	"bff-service/pkg/response"
	"context"
)

type AggregatorService struct {
	authClient    *clients.AuthClient
	companyClient *clients.CompanyClient
}

func NewAggregatorService(auth *clients.AuthClient, company *clients.CompanyClient) *AggregatorService {
	return &AggregatorService{
		authClient:    auth,
		companyClient: company,
	}
}

func (s *AggregatorService) GetCompanyMembers(ctx context.Context, company_uuid string) ([]returning.AggregatedMemberResponse, error) {

	members, err := s.companyClient.GetMembers(ctx, company_uuid)
	if err != nil {
		return []returning.AggregatedMemberResponse{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка получения списка участников компании",
			Data:    err,
		}
	}

	if len(members) == 0 {
		return []returning.AggregatedMemberResponse{}, nil
	}

	uuids := make([]string, len(members))
	for i, m := range members {
		uuids[i] = m.AccountUuid
	}

	profiles, err := s.authClient.GetProfilesBatch(ctx, uuids)
	if err != nil {
		return []returning.AggregatedMemberResponse{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка получения данных участников компании",
			Data:    err,
		}
	}

	profileMap := make(map[string]returning.AccountBatchResponse)
	for _, p := range profiles {
		profileMap[p.AccountUuid] = p
	}

	result := make([]returning.AggregatedMemberResponse, 0, len(members))
	for _, m := range members {
		p := profileMap[m.AccountUuid]
		result = append(result, returning.AggregatedMemberResponse{
			AccountUuid:  m.AccountUuid,
			Name:         p.Name,
			Nickname:     p.Nickname,
			AvatarFileID: p.AvatarFileID,
			RoleID:       m.RoleID,
			RoleName:     m.RoleName,
			RoleColor:    m.RoleColor,
			CreatedAt:    m.CreatedAt,
		})
	}

	return result, nil
}
