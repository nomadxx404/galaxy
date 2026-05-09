package app

import (
	"companies-service/internal/api"
	"companies-service/internal/repository/db"
	"companies-service/internal/service"
	"companies-service/pkg/config"
	"companies-service/pkg/data"
	"companies-service/pkg/kafka/worker"
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Config            *config.Config
	Pool              *pgxpool.Pool
	Rdb               *redis.Client
	Store             *db.Queries
	txManager         *data.TransactionManager
	Router            *gin.Engine
	CompanyHandler    *api.CompanyHandler
	RoleHandler       *api.RoleHandler
	InvitationHandler *api.InvitationHandler
	MemberHandler     *api.MemberHandler
	PermissionHandler *api.PermissionHandler
	Producer          *worker.Producer
	Consumer          *worker.Consumer
}

func NewApp(ctx context.Context) (*App, func()) {
	cfg := config.GetConfig()

	pool, err := data.PostgresPool(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	rdb, err := data.RedisPool(cfg.Redis)
	if err != nil {
		log.Fatalf("failed to init redis: %v", err)
	}

	q := db.New(pool)
	txManager := data.NewTransactionManager(pool)

	roleService := service.NewRoleService(rdb, pool, txManager)
	permissionService := service.NewPermissionService(rdb, pool, txManager)
	membersService := service.NewMemberService(rdb, pool, txManager, permissionService)
	invitationService := service.NewInvitationService(rdb, pool, txManager, cfg, roleService, membersService)
	companyService := service.NewCompanyService(rdb, pool, txManager, membersService, roleService, permissionService)

	producer := worker.NewRelay(pool, q, cfg)
	consumer := worker.NewConsumer(companyService, cfg, pool, txManager)

	compHandler := api.NewCompanyHandler(companyService)
	roleHandler := api.NewRoleHandler(roleService)
	invitationHandler := api.NewInvitationHandler(invitationService)
	memberHandler := api.NewMemberHandler(membersService)
	permissionHandler := api.NewPermissionHandler(permissionService)

	router := gin.New()

	log.Println("Infrastructure, Services and Handlers initialized")

	closeFunc := func() {
		pool.Close()
		rdb.Close()
		log.Println("Connections closed")
	}

	return &App{
		Config:            cfg,
		Pool:              pool,
		Rdb:               rdb,
		Store:             q,
		txManager:         txManager,
		Router:            router,
		CompanyHandler:    compHandler,
		RoleHandler:       roleHandler,
		InvitationHandler: invitationHandler,
		MemberHandler:     memberHandler,
		PermissionHandler: permissionHandler,
		Producer:          producer,
		Consumer:          consumer,
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
