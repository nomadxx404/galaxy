package request

type CreateCompanyRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	LogoFileID  *int32  `json:"logo_file_id"`
}

type UpdateCompanyRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	LogoFileID  *int32  `json:"logo_file_id"`
}
