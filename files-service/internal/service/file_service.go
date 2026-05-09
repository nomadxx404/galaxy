package service

import (
	"context"
	"database/sql"
	"errors"
	"files-service/internal/dto/request"
	"files-service/internal/dto/returning"
	"files-service/internal/repository/db"
	"files-service/pkg/config"
	"files-service/pkg/data"
	"files-service/pkg/kafka"
	"files-service/pkg/minio"
	"files-service/pkg/response"
	usercontext "files-service/pkg/user_context"
	pgxutil "files-service/pkg/utils"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type fileService struct {
	pool      *pgxpool.Pool
	txManager *data.TransactionManager
	storage   *minio.Storage
	cfg       *config.Config
}

func NewFileService(
	pool *pgxpool.Pool,
	txManager *data.TransactionManager,
	storage *minio.Storage,
	cfg *config.Config) *fileService {

	return &fileService{
		pool:      pool,
		txManager: txManager,
		storage:   storage,
		cfg:       cfg,
	}
}

func (s *fileService) CreateFile(
	ctx context.Context,
	dbExecutor db.DBTX,
	req request.CreateFileRequest) (db.CreateFileRow, error) {

	src, err := req.File.Open()
	if err != nil {
		return db.CreateFileRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка открытия файла",
			Data:    err,
		}
	}
	defer src.Close()

	storageKey := fmt.Sprintf("%s/%s", req.EntityType, uuid.Must(uuid.NewV7()).String())
	err = s.storage.Upload(ctx, storageKey, src, req.File.Size, req.File.Header.Get("Content-Type"))
	if err != nil {
		return db.CreateFileRow{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка сохранения файла",
			Data:    err,
		}
	}

	var result db.CreateFileRow
	err = s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)
		fileInfo, err := q.CreateFile(ctx, db.CreateFileParams{
			EntityType:  req.EntityType,
			EntityID:    pgxutil.TextNotValid(req.EntityID),
			Name:        req.File.Filename,
			Size:        req.File.Size,
			ContentType: req.File.Header.Get("Content-Type"),
			StorageKey:  storageKey,
			CreatedBy:   usercontext.GetAccountUuid(ctx),
		})

		if err != nil {
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка сохранения информации о файле",
				Data:    err,
			}
		}

		payload := map[string]any{
			"fileRequest": req,
			"fileInfo":    fileInfo,
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.FileCreated, payload); err != nil {
			return err
		}

		result = fileInfo
		return nil
	})

	return result, err
}

func (s *fileService) GetFiles(
	ctx context.Context,
	dbExecutor db.DBTX,
	entity_id string,
	req request.GetFilesRequest) ([]returning.FileInfoReturning, error) {

	if dbExecutor == nil {
		dbExecutor = s.pool
	}
	q := db.New(dbExecutor)

	filesInfo, err := q.GetFiles(ctx, db.GetFilesParams{
		Column1:  req.FileIDs,
		EntityID: pgxutil.TextValid(entity_id),
	})

	if err != nil {
		return nil, &response.ApiError{
			Status:  500,
			Message: "Ошибка получения файлов",
			Data:    err,
		}
	}

	res := make([]returning.FileInfoReturning, len(filesInfo))
	for i, row := range filesInfo {
		res[i] = returning.FileInfoReturning{
			FileID:      row.FileID,
			EntityType:  row.EntityType,
			EntityID:    row.EntityID.String,
			Name:        row.Name,
			Size:        row.Size,
			ContentType: row.ContentType,
			URL:         fmt.Sprintf("%s/api/files/download/%d", s.cfg.BaseUrl.FILES_SERVICE_URL, row.FileID),
			Status:      string(row.Status),
			CreatedAt:   row.CreatedAt,
		}
	}

	return res, nil
}

func (s *fileService) DeleteFiles(
	ctx context.Context,
	dbExecutor db.DBTX,
	file_id int64,
	entity_id string) error {

	var keysToDelete string
	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		storageKey, err := q.DeleteFile(ctx, db.DeleteFileParams{
			FileID:   file_id,
			EntityID: pgxutil.TextValid(entity_id),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return &response.ApiError{
					Status:  404,
					Message: "Файл не найден",
				}
			}
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка удаления файла",
				Data:    err,
			}
		}

		payload := map[string]any{
			"file_id":   file_id,
			"entity_id": entity_id,
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.FileDeleted, payload); err != nil {
			return err
		}

		keysToDelete = storageKey
		return nil
	})

	if err == nil {
		err = s.storage.RemoveObject(ctx, keysToDelete)
		if err != nil {
			return fmt.Errorf("[MINIO] failed to delete object: %w", err)
		}
	}

	return err
}

func (s *fileService) DownloadFile(
	ctx context.Context,
	file_id int64,
) (*returning.DownloadFileDataReturning, error) {
	q := db.New(s.pool)

	fileRow, err := q.GetFileById(ctx, file_id)
	if err != nil {
		return &returning.DownloadFileDataReturning{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка получения данных файла",
			Data:    err,
		}
	}

	object, err := s.storage.GetObject(ctx, fileRow.StorageKey)
	if err != nil {
		return &returning.DownloadFileDataReturning{}, &response.ApiError{
			Status:  500,
			Message: "Ошибка скачивания файла",
			Data:    err,
		}
	}

	return &returning.DownloadFileDataReturning{
		Name:        fileRow.Name,
		ContentType: fileRow.ContentType,
		Size:        fileRow.Size,
		File:        object,
	}, nil
}

func (s *fileService) HandleGlobalEntityDeletion(
	ctx context.Context,
	dbExecutor db.DBTX,
	entity_id string,
) error {

	var keysToDelete []string
	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		storageKeys, err := q.DeleteAllFilesEntity(ctx, pgxutil.TextValid(entity_id))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return &response.ApiError{
					Status:  404,
					Message: "Файл не найден",
				}
			}
			return &response.ApiError{
				Status:  500,
				Message: "Ошибка удаления файла",
				Data:    err,
			}
		}

		payload := map[string]any{
			"entity_id":    entity_id,
			"storage_keys": storageKeys,
		}

		if err = kafka.EmitOutbox(ctx, tx, kafka.FileDeleted, payload); err != nil {
			return err
		}

		keysToDelete = storageKeys
		return nil
	})

	if err != nil {
		return &response.ApiError{
			Status:  500,
			Message: "Ошибка удаления записей из БД",
			Data:    err,
		}
	}

	if len(keysToDelete) > 0 {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			if err := s.storage.RemoveObjects(bgCtx, keysToDelete); err != nil {
				log.Printf("[ERROR] Batch delete failed: %v", err)
			}
		}()
	}

	return err
}

func (s *fileService) MarkFileAsProcessed(
	ctx context.Context,
	dbExecutor db.DBTX,
	file_id *int64,
	entity_id string) error {

	err := s.txManager.WithinTransaction(ctx, dbExecutor, func(tx db.DBTX) error {
		q := db.New(tx)

		log.Printf("[DEBAG] Parametrs: %v, %v", file_id, entity_id)

		_, err := q.MarkFileAsProcessed(ctx, db.MarkFileAsProcessedParams{
			FileID:   *file_id,
			EntityID: pgxutil.TextValid(entity_id),
		})

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return err
			}
			return err
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Update status file: %v", err)
		return err
	}
	return nil
}
