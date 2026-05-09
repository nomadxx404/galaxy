package minio

import (
	"context"
	"files-service/pkg/config"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client *minio.Client
	bucket string
}

func NewStorage(cfg *config.Config) (*Storage, error) {
	endpoint := fmt.Sprintf("%s:%s", cfg.Mimio.Host, cfg.Mimio.Port)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Mimio.RootUser, cfg.Mimio.RootPassword, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("[MINIO] Failed to create minio client: %w", err)
	}

	return &Storage{
		client: client,
		bucket: cfg.Mimio.BucketName,
	}, nil
}

func (s *Storage) InitBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}

	if !exists {
		log.Printf("[MINIO] Bucket %s does not exist, creating...", s.bucket)
		return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
	}

	return nil
}

func (s *Storage) Upload(ctx context.Context, storageKey string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, storageKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("[MINIO] Failed to upload object: %w", err)
	}

	return nil
}

func (s *Storage) RemoveObject(ctx context.Context, storageKey string) error {
	opts := minio.RemoveObjectOptions{
		ForceDelete: false,
	}

	err := s.client.RemoveObject(ctx, s.bucket, storageKey, opts)
	if err != nil {
		return fmt.Errorf("[MINIO] failed to remove object %s: %w", storageKey, err)
	}

	log.Printf("[MINIO] Object %s successfully removed from bucket %s", storageKey, s.bucket)
	return nil
}

func (s *Storage) RemoveObjects(ctx context.Context, storageKeys []string) error {
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for _, key := range storageKeys {
			objectsCh <- minio.ObjectInfo{
				Key: key,
			}
		}
	}()

	errorCh := s.client.RemoveObjects(ctx, s.bucket, objectsCh, minio.RemoveObjectsOptions{})

	var errs []error
	for e := range errorCh {
		if e.Err != nil {
			errs = append(errs, fmt.Errorf("failed to remove object %s: %w", e.ObjectName, e.Err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("[MINIO] batch removal completed with errors: %v", errs)
	}

	log.Printf("[MINIO] Successfully removed %d objects from bucket %s", len(storageKeys), s.bucket)
	return nil
}

func (s *Storage) GetObject(ctx context.Context, storageKey string) (*minio.Object, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, storageKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("[MINIO] Failed to get object %s: %w", storageKey, err)
	}

	_, err = obj.Stat()
	if err != nil {
		return nil, fmt.Errorf("[MINIO] Object %s not found or inaccessible: %w", storageKey, err)
	}

	return obj, nil
}
