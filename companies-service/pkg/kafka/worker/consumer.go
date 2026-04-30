package worker

import (
	"context"
	"encoding/json"
	"log"

	"companies-service/internal/repository/db"
	"companies-service/pkg/config"
	"companies-service/pkg/data"
	kafkaEvents "companies-service/pkg/kafka"
	usercontext "companies-service/pkg/user_context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type ConsumnerCompanyManager interface {
	HandleGlobalAccountDeletion(ctx context.Context, dbExecutor db.DBTX, account_uuid string) error
}

type EventMessage struct {
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

type Consumer struct {
	readers        []*kafka.Reader
	companyService ConsumnerCompanyManager
	handlers       map[string]EventHandler
	cfg            *config.Config
	pool           *pgxpool.Pool
	txManager      *data.TransactionManager
}

type KafkaEvent struct {
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
	Metadata  struct {
		RequestID string `json:"request_id"`
	} `json:"metadata"`
}

type EventHandler func(ctx context.Context, event KafkaEvent) error

func NewConsumer(
	companyService ConsumnerCompanyManager,
	cfg *config.Config,
	pool *pgxpool.Pool,
	txManager *data.TransactionManager) *Consumer {

	authReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.Kafka.BROKERS},
		Topic:   cfg.Kafka.EVENTS_TOPIC_AUTH,
		GroupID: "companies-group",
	})

	companyReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.Kafka.BROKERS},
		Topic:   cfg.Kafka.EVENTS_TOPIC,
		GroupID: "companies-group",
	})

	c := &Consumer{
		cfg:            cfg,
		pool:           pool,
		txManager:      txManager,
		companyService: companyService,
		readers:        []*kafka.Reader{authReader, companyReader},
	}

	c.handlers = map[string]EventHandler{
		kafkaEvents.AccountDeleted: c.eventAccountDeleted,
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

func (c *Consumer) eventAccountDeleted(ctx context.Context, event KafkaEvent) error {

	var payloadData struct {
		AccountUuid string `json:"account_uuid"`
	}

	if err := json.Unmarshal(event.Payload, &payloadData); err != nil {
		log.Printf("[ERROR] Failed to unmarshal payload: %v", err)
		return err
	}

	if err := json.Unmarshal(event.Payload, &payloadData); err != nil {
		log.Printf("[ERROR] Failed to unmarshal payload: %v", err)
		return err
	}

	ctx = usercontext.WithAccountUuid(ctx, payloadData.AccountUuid)
	ctx = usercontext.WithRequestId(ctx, event.Metadata.RequestID)
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		log.Printf("[ERROR] Failed to begin transaction: %v", err)
		return err
	}
	defer tx.Rollback(ctx)

	return c.txManager.WithinTransaction(ctx, nil, func(tx db.DBTX) error {
		return c.companyService.HandleGlobalAccountDeletion(ctx, tx, payloadData.AccountUuid)
	})
}
