package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	dndapp "dnd_schedule/internal/app"
	"dnd_schedule/internal/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// @title           Terra DND masters-service API
// @version         1.0
// @description
// @termsOfService  https://terragames.ru/

// @contact.name   API Support
// @contact.url    https://t.me/dmastr

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1/dnd

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	log := logrus.New()
	logrus.SetLevel(logrus.DebugLevel)
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer func() {
		stop()
	}()

	app, err := dndapp.NewDndMastersApp(ctx, cfg, log)
	if err != nil {
		panic(err)
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := app.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("app run failed: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		return app.Stop(ctx)
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
