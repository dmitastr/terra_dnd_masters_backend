package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"dnd_schedule/internal/config"
	authenticateservice "dnd_schedule/internal/domain/service/authenticate-service"
	demands_service "dnd_schedule/internal/domain/service/demands-service"
	masters_service "dnd_schedule/internal/domain/service/masters-service"
	slots_service "dnd_schedule/internal/domain/service/slots-service"
	"dnd_schedule/internal/presentation/core/middleware"
	authenticate_handlers "dnd_schedule/internal/presentation/features/authenticate/authenticate-handlers"
	demandsHandlersPkg "dnd_schedule/internal/presentation/features/demands-handlers"
	mastersHandlersPkg "dnd_schedule/internal/presentation/features/masters-handlers"
	slotsHandlersPkg "dnd_schedule/internal/presentation/features/slots-handlers"
	"dnd_schedule/internal/repository/migrations"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"dnd_schedule/docs"
)

type DndMastersApp struct {
	server *http.Server
	cfg    config.ConfigProvider
}

func NewDndMastersApp(ctx context.Context, cfg config.ConfigProvider, log *logrus.Logger) (*DndMastersApp, error) {
	docs.SwaggerInfo.Title = "Swagger Example API"
	docs.SwaggerInfo.Description = "Terra dnd scheduling app"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "terra.ru"
	docs.SwaggerInfo.BasePath = "api/v1/dnd"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	app := &DndMastersApp{cfg: cfg}
	router := gin.Default()

	// runner := migrations.NewRunner()
	runner := migrations.NewSQLiteRunner()
	dsProvider, err := runner.Run(ctx, cfg.GetDBConfig(), log)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// db := datasource.NewDatasource()

	mastersService := masters_service.NewMastersService(dsProvider, log)
	mastersHandler := mastersHandlersPkg.NewMastersHandler(cfg, mastersService, log)

	demandsService := demands_service.NewDemandsService(dsProvider, log)
	demandsHandler := demandsHandlersPkg.NewDemandsHandler(cfg, demandsService, log)

	slotsService := slots_service.NewSlotsService(dsProvider, log)
	slotsHandler := slotsHandlersPkg.NewSlotsHandler(cfg, slotsService, log)

	authService := authenticateservice.NewAuthService(dsProvider)
	middlewareProvider := middleware.NewMiddlewareProvider(cfg, authService)
	authHandler := authenticate_handlers.NewAuthHandler(cfg, authService)

	if err := app.registerHandlers(router, mastersHandler, slotsHandler, demandsHandler, middlewareProvider, authHandler); err != nil {
		return nil, fmt.Errorf("register handlers failed: %s", err)
	}

	srv := &http.Server{
		Addr:              cfg.GetAddress(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		Handler:           router,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	app.server = srv
	return app, nil
}

func (app *DndMastersApp) Start() error {
	if err := app.server.ListenAndServe(); err != nil {
		return fmt.Errorf("start server failed: %s", err)
	}
	return nil
}

func (app *DndMastersApp) Stop(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := app.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server failed: %s", err)
	}
	return nil
}

func (app *DndMastersApp) registerHandlers(router *gin.Engine, mastersHandler mastersHandlersPkg.IMastersHandler, slotsHandler slotsHandlersPkg.ISlotsHandler, demandsHandler demandsHandlersPkg.IDemandsHandler, middlewareProvider *middleware.MiddlewareProvider, authHandlers authenticate_handlers.IAuthHandler) error {
	router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithDecompressFn(gzip.DefaultDecompressHandle)))
	apiPath := router.Group("/api/v1/dnd")

	// use ginSwagger middleware to serve the API docs
	apiPath.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authPath := apiPath.Group("auth")
	authPath.POST(`/register`, authHandlers.AddUser)
	authPath.GET(`/token`, authHandlers.GetToken)

	mastersPath := apiPath.Group("masters")
	mastersPath.Use(middlewareProvider.Middleware())

	mastersPath.GET(`/`, mastersHandler.GetMasters)
	mastersPath.POST(`/`, mastersHandler.AddMasters)
	mastersPath.DELETE(`/:id`, mastersHandler.DeleteMasters)

	demandsPath := apiPath.Group("demands")
	demandsPath.Use(middlewareProvider.Middleware())

	demandsPath.GET(`/`, demandsHandler.GetDemands)
	demandsPath.GET(`/:vk_id`, demandsHandler.GetDemandForUser)
	demandsPath.POST(`/`, demandsHandler.AddDemands)

	slotsPath := apiPath.Group("slots")
	slotsPath.Use(middlewareProvider.Middleware())

	slotsPath.GET(`/`, slotsHandler.GetSlots)
	slotsPath.POST(`/`, slotsHandler.AddSlots)

	return nil
}
