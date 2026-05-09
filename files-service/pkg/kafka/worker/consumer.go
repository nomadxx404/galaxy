package worker

import (
	"context"
	"encoding/json"
	"log"

	"files-service/internal/repository/db"
	"files-service/pkg/config"
	"files-service/pkg/data"
	kafkaEvents "files-service/pkg/kafka"
	usercontext "files-service/pkg/user_context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type ConsumnerFileManager interface {
	HandleGlobalEntityDeletion(ctx context.Context, dbExecutor db.DBTX, entity_id string) error
	MarkFileAsProcessed(ctx context.Context, dbExecutor db.DBTX, file_id *int64, entity_id string) error
}

type EventMessage struct {
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

type Consumer struct {
	readers     []*kafka.Reader
	fileService ConsumnerFileManager
	handlers    map[string]EventHandler
	cfg         *config.Config
	pool        *pgxpool.Pool
	txManager   *data.TransactionManager
}

type KafkaEvent struct {
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
	Metadata  struct {
		RequestID   string `json:"request_id"`
		AccountUuid string `json:"account_uuid"`
	} `json:"metadata"`
}

type EventHandler func(ctx context.Context, event KafkaEvent) error

func NewConsumer(
	fileService ConsumnerFileManager,
	cfg *config.Config,
	pool *pgxpool.Pool,
	txManager *data.TransactionManager) *Consumer {

	authReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.Kafka.BROKERS},
		Topic:   cfg.Kafka.EVENTS_TOPIC_AUTH,
		GroupID: "files-group",
	})

	companyReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.Kafka.BROKERS},
		Topic:   cfg.Kafka.EVENTS_TOPIC_COMPANY,
		GroupID: "files-group",
	})

	c := &Consumer{
		cfg:         cfg,
		pool:        pool,
		txManager:   txManager,
		fileService: fileService,
		readers:     []*kafka.Reader{authReader, companyReader},
	}

	c.handlers = map[string]EventHandler{
		kafkaEvents.AccountUpdated: c.eventAccountProcessing,
		kafkaEvents.AccountDeleted: c.eventAccountDeleted,

		kafkaEvents.CompanyCreated: c.eventCompanyProcessing,
		kafkaEvents.CompanyUpdated: c.eventCompanyProcessing,
		kafkaEvents.CompanyDeleted: c.eventCompanyDeleted,
	}

	return c
}

func (c *Consumer) Run(ctx context.Context) {
	for _, r := range c.readers {
		go c.listen(ctx, r)
	}
}

func (c *Consumer) listen(ctx context.Context, r *kafka.Reader) {
	log.Printf("Starting consumer for topic: %s", r.Config().Topic)

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("Error reading message from %s: %v", r.Config().Topic, err)
			continue
		}

		var msg KafkaEvent
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		if handler, ok := c.handlers[msg.EventType]; ok {
			if err := handler(ctx, msg); err != nil {
				log.Printf("Error handling event %s: %v", msg.EventType, err)
			}
		}
	}
}

func GenericHandler[T any](
	ctx context.Context,
	event KafkaEvent,
	c *Consumer,
	action func(ctx context.Context, tx db.DBTX, payload T) error,
) error {
	var payload T

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		var rawString string
		if strErr := json.Unmarshal(event.Payload, &rawString); strErr == nil {
			log.Printf("[WARN] Received raw string, attempting to cast to payload type")
		} else {
			log.Printf("[ERROR] Failed to unmarshal payload: %v", err)
			return err
		}
	}

	ctx = usercontext.WithRequestId(ctx, event.Metadata.RequestID)
	ctx = usercontext.WithAccountUuid(ctx, event.Metadata.AccountUuid)

	return c.txManager.WithinTransaction(ctx, nil, func(tx db.DBTX) error {
		return action(ctx, tx, payload)
	})
}

func (c *Consumer) eventAccountDeleted(ctx context.Context, event KafkaEvent) error {
	type payload struct {
		AccountUuid string `json:"account_uuid"`
	}

	return GenericHandler(ctx, event, c, func(ctx context.Context, tx db.DBTX, p payload) error {
		return c.fileService.HandleGlobalEntityDeletion(ctx, tx, p.AccountUuid)
	})
}

func (c *Consumer) eventCompanyDeleted(ctx context.Context, event KafkaEvent) error {
	type payload struct {
		CompanyUuid string `json:"company_uuid"`
	}

	return GenericHandler(ctx, event, c, func(ctx context.Context, tx db.DBTX, p payload) error {
		return c.fileService.HandleGlobalEntityDeletion(ctx, tx, p.CompanyUuid)
	})
}

func (c *Consumer) eventAccountProcessing(ctx context.Context, event KafkaEvent) error {
	type payload struct {
		AccountUuid string `json:"account_uuid"`
		FileId      *int64 `json:"avatar_file_id"`
	}

	return GenericHandler(ctx, event, c, func(ctx context.Context, tx db.DBTX, p payload) error {
		if p.FileId == nil {
			return nil
		}
		return c.fileService.MarkFileAsProcessed(ctx, tx, p.FileId, p.AccountUuid)
	})
}

func (c *Consumer) eventCompanyProcessing(ctx context.Context, event KafkaEvent) error {
	type payload struct {
		CompanyUuid string `json:"company_uuid"`
		FileId      *int64 `json:"logo_file_id"`
	}

	return GenericHandler(ctx, event, c, func(ctx context.Context, tx db.DBTX, p payload) error {
		if p.FileId == nil {
			return nil
		}
		return c.fileService.MarkFileAsProcessed(ctx, tx, p.FileId, p.CompanyUuid)
	})
}
