//go:build wireinject

package main

import (
	"github.com/google/wire"
	"gorm.io/gorm"
	"menlo.ai/indigo-api-gateway/app/domain"
	"menlo.ai/indigo-api-gateway/app/infrastructure"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository"
	"menlo.ai/indigo-api-gateway/app/interfaces/http"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/routes"
)

func CreateApplication() (*Application, error) {
	wire.Build(
		database.NewDB,
		repository.RepositoryProvider,
		infrastructure.InfrastructureProvider,
		domain.ServiceProvider,
		routes.RouteProvider,
		http.NewHttpServer,
		wire.Struct(new(Application), "*"),
	)
	return nil, nil
}

func ProvideDatabase() *gorm.DB {
	return database.DB
}

func CreateDataInitializer() (*DataInitializer, error) {
	wire.Build(
		ProvideDatabase,
		repository.RepositoryProvider,
		infrastructure.InfrastructureProvider,
		domain.ServiceProvider,
		wire.Struct(new(DataInitializer), "*"),
	)
	return nil, nil
}
