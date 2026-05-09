package returning

import (
	"io"
	"time"
)

type FileInfoReturning struct {
	FileID      int64     `json:"file_id"`
	EntityType  string    `json:"entity_type"`
	EntityID    string    `json:"entity_id"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	URL         string    `json:"url"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type DownloadFileDataReturning struct {
	Name        string
	ContentType string
	Size        int64
	File        io.ReadCloser
}
