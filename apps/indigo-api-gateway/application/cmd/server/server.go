package main

import (
	"context"
	nethttp "net/http"
	_ "net/http/pprof"

	_ "github.com/grafana/pyroscope-go/godeltaprof/http/pprof"

	"github.com/mileusna/crontab"
	"menlo.ai/indigo-api-gateway/app/domain/cron"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database"
	apphttp "menlo.ai/indigo-api-gateway/app/interfaces/http"
	"menlo.ai/indigo-api-gateway/app/utils/httpclients/serper"
	"menlo.ai/indigo-api-gateway/app/utils/logger"
	"menlo.ai/indigo-api-gateway/config/environment_variables"
)

type Application struct {
	HttpServer  *apphttp.HttpServer
	CronService *cron.CronService
}

func (application *Application) Start() {
	// Start cron service
	cronTab := crontab.New()
	background := context.Background()
	application.CronService.Start(background, cronTab)

	// Start HTTP server
	if err := application.HttpServer.Run(); err != nil {
		panic(err)
	}
}

func init() {
	logger.GetLogger()
	environment_variables.EnvironmentVariables.LoadFromEnv()
	serper.Init()
}

// @title Indigo Server
// @version 1.0
// @description This is the API gateway for Indigo Server.
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	background := context.Background()

	// Expose pprof endpoints for profiling (for Grafana Alloy/Pyroscope Go pull mode)
	go func() {
		// Default pprof mux is registered on DefaultServeMux by importing net/http/pprof
		// Listen on localhost:6060 (or change port as needed)
		if err := nethttp.ListenAndServe("0.0.0.0:6060", nil); err != nil {
			logger.GetLogger().Errorf("pprof server failed: %v", err)
		}
	}()

	application, err := CreateApplication()
	if err != nil {
		panic(err)
	}
	err = database.Migration()
	if err != nil {
		panic(err)
	}
	dataInitializer, err := CreateDataInitializer()
	if err != nil {
		panic(err)
	}
	err = dataInitializer.Install(background)
	if err != nil {
		panic(err)
	}
	application.Start()
}
