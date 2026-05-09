package app

import (
	"context"
	"files-service/internal/api"
	"files-service/internal/repository/db"
	"files-service/internal/service"
	"files-service/pkg/config"
	"files-service/pkg/data"
	"files-service/pkg/kafka/worker"
	"files-service/pkg/minio"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Config       *config.Config
	Pool         *pgxpool.Pool
	Store        *db.Queries
	txManager    *data.TransactionManager
	Router       *gin.Engine
	FileHandler  *api.FileHandler
	Producer     *worker.Producer
	Consumer     *worker.Consumer
	MinioStorage *minio.Storage
}

func NewApp(ctx context.Context) (*App, func()) {
	cfg := config.GetConfig()

	pool, err := data.PostgresPool(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	q := db.New(pool)
	txManager := data.NewTransactionManager(pool)

	mStorage, err := minio.NewStorage(cfg)
	if err != nil {
		log.Fatalf("failed to initialize minio storage: %v", err)
	}

	if err := mStorage.InitBucket(ctx); err != nil {
		log.Fatalf("failed to create/verify minio bucket: %v", err)
	}
	log.Printf("MinIO bucket '%s' is ready", cfg.Mimio.BucketName)

	fileService := service.NewFileService(pool, txManager, mStorage, cfg)

	producer := worker.NewRelay(pool, q, cfg)
	consumer := worker.NewConsumer(fileService, cfg, pool, txManager)

	fileHandler := api.NewFileHandler(fileService)

	router := gin.New()

	log.Println("Infrastructure, Services and Handlers initialized")

	closeFunc := func() {
		pool.Close()
		log.Println("Connections closed")
	}

	return &App{
		Config:       cfg,
		Pool:         pool,
		Store:        q,
		txManager:    txManager,
		Router:       router,
		FileHandler:  fileHandler,
		Producer:     producer,
		Consumer:     consumer,
		MinioStorage: mStorage,
	}, closeFunc
}

func (a *App) Run(ctx context.Context) error {
	a.RegisterRoutes()

	go a.Producer.Run(ctx)
	log.Println("Background workers started")

	go a.Consumer.Run(ctx)
	log.Println("Background consumer started")

	serverAddr := ":8080"
	log.Printf("Starting server on %s", serverAddr)
	return a.Router.Run(serverAddr)
}
