package api

import (
	"context"
	"files-service/internal/dto/request"
	"files-service/internal/dto/returning"
	"files-service/internal/repository/db"
	_ "files-service/internal/repository/db"
	"files-service/pkg/response"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FileManager interface {
	CreateFile(ctx context.Context, dbExecutor db.DBTX, req request.CreateFileRequest) (db.CreateFileRow, error)
	GetFiles(ctx context.Context, dbExecutor db.DBTX, entity_id string, req request.GetFilesRequest) ([]returning.FileInfoReturning, error)
	DeleteFiles(ctx context.Context, dbExecutor db.DBTX, file_id int64, entity_id string) error
	DownloadFile(ctx context.Context, file_id int64) (*returning.DownloadFileDataReturning, error)
}

type FileHandler struct {
	service FileManager
}

func NewFileHandler(s FileManager) *FileHandler {
	return &FileHandler{
		service: s,
	}
}

func (h *FileHandler) CreateFile(c *gin.Context) {
	var req request.CreateFileRequest

	if err := c.ShouldBind(&req); err != nil {
		response.SendFailure(c, 400, "Некорректный формат данных")
		return
	}

	res, err := h.service.CreateFile(c.Request.Context(), nil, req)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 201, res, "Файл успешно сохранен")
}

func (h *FileHandler) GetFiles(c *gin.Context) {
	var req request.GetFilesRequest

	entity_id := c.Param("entity_id")
	if entity_id == "" {
		response.SendFailure(c, 400, "ID сущности не указан")
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.SendFailure(c, 400, "Некорректный формат данных")
		return
	}

	res, err := h.service.GetFiles(c.Request.Context(), nil, entity_id, req)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, res, "Успешно")
}

func (h *FileHandler) DeleteFiles(c *gin.Context) {

	file_id_param := c.Param("file_id")
	file_id, err := strconv.ParseInt(file_id_param, 10, 64)
	if err != nil {
		response.SendFailure(c, 400, "ID файла должен быть числом")
		return
	}

	entity_id := c.Param("entity_id")
	if entity_id != "" {
		response.SendFailure(c, 400, "ID сущности не указан")
		return
	}

	err = h.service.DeleteFiles(c.Request.Context(), nil, file_id, entity_id)

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SendSuccess(c, 200, nil, "Успешно")
}

func (h *FileHandler) DownloadFile(c *gin.Context) {

	file_id_param := c.Param("file_id")
	file_id, err := strconv.ParseInt(file_id_param, 10, 64)
	if err != nil {
		response.SendFailure(c, 400, "ID файла должен быть числом")
		return
	}

	res, err := h.service.DownloadFile(c.Request.Context(), file_id)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	defer res.File.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", res.Name))
	c.Header("Content-Type", res.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", res.Size))

	_, err = io.Copy(c.Writer, res.File)
	if err != nil {
		log.Printf("[ERROR] Failed to stream file: %v", err)
	}
}
