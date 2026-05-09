package worker

import (
	"context"
	"log"
	"time"

	"files-service/internal/repository/db"
	"files-service/pkg/config"
	pgxutil "files-service/pkg/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	store  *db.Queries
	pool   *pgxpool.Pool
	writer *kafka.Writer
	cfg    *config.Config
}

func NewRelay(pool *pgxpool.Pool, store *db.Queries, cfg *config.Config) *Producer {
	return &Producer{
		pool:  pool,
		store: store,
		cfg:   cfg,
		writer: &kafka.Writer{
			Addr:     kafka.TCP(cfg.Kafka.BROKERS),
			Topic:    cfg.Kafka.EVENTS_TOPIC_FILE,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (r *Producer) Run(ctx context.Context) {
	log.Println("[KAFKA] Outbox Producer worker started")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[KAFKA] Outbox Producer worker stopping...")
			r.writer.Close()
			return
		case <-ticker.C:
			r.process(ctx)
		}
	}
}

func (r *Producer) process(ctx context.Context) {
	events, err := r.store.GetUnprocessedEvents(ctx)
	if err != nil {
		log.Printf("[KAFKA] Error fetching outbox events: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	for _, event := range events {
		msg := kafka.Message{
			Key:   []byte(event.AccountUuid),
			Value: event.Payload,
		}

		err := r.writer.WriteMessages(ctx, msg)
		if err != nil {
			log.Printf("[KAFKA] Failed to send message %s to Kafka: %v", event.MessageUuid, err)
			_ = r.store.MarkEventFailed(ctx, db.MarkEventFailedParams{
				MessageUuid: event.MessageUuid,
				LastError:   pgxutil.TextValid(err.Error()),
			})
			continue
		}

		err = r.store.MarkEventProcessed(ctx, event.MessageUuid)
		if err != nil {
			log.Printf("[KAFKA] Failed to mark event %s as processed: %v", event.MessageUuid, err)
		}
	}
}
