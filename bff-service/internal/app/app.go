package app

import (
	"bff-service/internal/api"
	"bff-service/internal/clients"
	"bff-service/internal/service"
	"bff-service/pkg/config"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *api.Handler
}

func NewApp(ctx context.Context) *App {
	cfg := config.GetConfig()
	router := gin.Default()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &clients.HeaderPropagator{
			Base: http.DefaultTransport,
		},
	}

	authCli := &clients.AuthClient{
		BaseURL: cfg.AuthServiceURL,
		Client:  httpClient,
	}
	compCli := &clients.CompanyClient{
		BaseURL: cfg.CompanyServiceURL,
		Client:  httpClient,
	}

	aggregator := service.NewAggregatorService(authCli, compCli)
	handler := api.NewHandler(aggregator)

	return &App{
		Config:  cfg,
		Router:  router,
		Handler: handler,
	}
}

func (a *App) Run() error {
	a.RegisterRoutes()

	serverAddr := ":8080"
	return a.Router.Run(serverAddr)
}
