package app

import (
	"companies-service/internal/api"
	"companies-service/internal/repository/db"
	"companies-service/internal/service"
	"companies-service/pkg/config"
	"companies-service/pkg/data"
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
	Router            *gin.Engine
	CompanyHandler    *api.CompanyHandler
	RoleHandler       *api.RoleHandler
	InvitationHandler *api.InvitationHandler
	MemberHandler     *api.MemberHandler
	PermissionHandler *api.PermissionHandler
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

	compService := service.NewCompanyService(q, rdb, pool)
	roleService := service.NewRoleService(q, rdb, pool)
	invitationService := service.NewInvitationService(q, rdb, pool)
	membersService := service.NewMemberService(q, rdb, pool)
	permissionService := service.NewPermissionService(q, rdb, pool)

	compHandler := api.NewCompanyHandler(compService)
	roleHandler := api.NewRoleHandler(roleService)
	invitationHandler := api.NewInvitationHandler(invitationService)
	memberHandler := api.NewMemberHandler(membersService)
	permissionHandler := api.NewPermissionHandler(permissionService)

	router := gin.Default()

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
		Router:            router,
		CompanyHandler:    compHandler,
		RoleHandler:       roleHandler,
		InvitationHandler: invitationHandler,
		MemberHandler:     memberHandler,
		PermissionHandler: permissionHandler,
	}, closeFunc
}

func (a *App) Run() error {
	a.RegisterRoutes()

	serverAddr := ":8080"
	return a.Router.Run(serverAddr)
}
