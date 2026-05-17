package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"dnd_schedule/internal/config"
	authenticateservice "dnd_schedule/internal/domain/service/authenticate-service"
	"dnd_schedule/internal/domain/service/demands-service"
	"dnd_schedule/internal/domain/service/masters-service"
	"dnd_schedule/internal/domain/service/slots-service"
	"dnd_schedule/internal/presentation/core/middleware"
	demandsHandlersPkg "dnd_schedule/internal/presentation/features/demands/demands-handlers"
	mastersHandlersPkg "dnd_schedule/internal/presentation/features/masters/masters-handlers"
	slotsHandlersPkg "dnd_schedule/internal/presentation/features/slots/slots-handlers"
	"dnd_schedule/internal/testing/datasource"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	"dnd_schedule/docs"
)

type DndMastersApp struct {
	server *http.Server
	cfg    config.ConfigProvider
}

func NewDndMastersApp(ctx context.Context, cfg config.ConfigProvider) (*DndMastersApp, error) {
	docs.SwaggerInfo.Title = "Swagger Example API"
	docs.SwaggerInfo.Description = "Terra dnd scheduling app"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "terra.ru"
	docs.SwaggerInfo.BasePath = "api/v1/dnd"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	app := &DndMastersApp{cfg: cfg}
	router := gin.Default()

	db := datasource.NewDatasource()
	mastersService := masters_service.NewMastersService(db)
	mastersHandler := mastersHandlersPkg.NewMastersHandler(cfg, mastersService)

	demandsService := demands_service.NewDemandsService(db)
	demandsHandler := demandsHandlersPkg.NewDemandsHandler(cfg, demandsService)

	slotsService := slots_service.NewSlotsService(db)
	slotsHandler := slotsHandlersPkg.NewSlotsHandler(cfg, slotsService)

	authService := authenticateservice.NewAuthService(db)
	middlewareProvider := middleware.NewMiddlewareProvider(cfg, authService)

	if err := app.registerHandlers(router, mastersHandler, slotsHandler, demandsHandler, middlewareProvider); err != nil {
		return nil, fmt.Errorf("register mastersHandlersPkg failed: %s", err)
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

func (app *DndMastersApp) registerHandlers(router *gin.Engine, mastersHandler mastersHandlersPkg.IMastersHandler, slotsHandler slotsHandlersPkg.ISlotsHandler, demandsHandler demandsHandlersPkg.IDemandsHandler, middlewareProvider *middleware.MiddlewareProvider) error {
	router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithDecompressFn(gzip.DefaultDecompressHandle)))
	apiPath := router.Group("/api/v1/dnd")

	// use ginSwagger middleware to serve the API docs
	apiPath.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	apiPath.Use(middlewareProvider.VerifyJWT)

	mastersPath := apiPath.Group("masters")

	mastersPath.GET(`/`, mastersHandler.GetMasters)
	mastersPath.POST(`/`, mastersHandler.AddMasters)
	mastersPath.DELETE(`/{id}`, mastersHandler.DeleteMasters)

	demandsPath := apiPath.Group("demands")

	demandsPath.GET(`/`, demandsHandler.GetDemands)
	demandsPath.POST(`/`, demandsHandler.AddDemands)

	slotsPath := apiPath.Group("slots")

	slotsPath.GET(`/`, slotsHandler.GetSlots)
	slotsPath.POST(`/`, slotsHandler.AddSlots)

	return nil
}
