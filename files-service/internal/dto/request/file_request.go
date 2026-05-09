package request

import "mime/multipart"

type CreateFileRequest struct {
	EntityType string               `form:"entity_type" binding:"required"`
	EntityID   *string              `form:"entity_id" `
	File       multipart.FileHeader `form:"file" binding:"required"`
}

type GetFilesRequest struct {
	FileIDs []int64 `json:"file_ids" `
}
