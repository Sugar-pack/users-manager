package main

import (
	"context"
	"log"

	"github.com/Sugar-pack/users-manager/internal/app"
	"github.com/Sugar-pack/users-manager/internal/config"
	"github.com/Sugar-pack/users-manager/internal/db"
	"github.com/Sugar-pack/users-manager/internal/migrations"
	"github.com/Sugar-pack/users-manager/pkg/logging"
)

func main() {
	ctx := context.Background()
	appConfig, err := config.GetAppConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	logger := logging.GetLogger()
	ctx = logging.WithContext(ctx, logger)
	err = migrations.Apply(ctx, appConfig.Db)

	if err != nil {
		log.Fatal(err)
		return
	}

	dbConn, err := db.Connect(ctx, appConfig.Db)
	if err != nil {
		log.Fatal(err)
		return
	}

	application := app.CreateApp(logger, dbConn, appConfig.Monitoring)
	application.Start(logger, appConfig.API)
}
