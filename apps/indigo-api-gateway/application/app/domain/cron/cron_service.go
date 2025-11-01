package cron

import (
	"context"

	"github.com/mileusna/crontab"
	"menlo.ai/indigo-api-gateway/config/environment_variables"
)

type CronService struct {
}

func NewCronService() *CronService {
	return &CronService{}
}

func (cs *CronService) Start(ctx context.Context, ctab *crontab.Crontab) {

	ctab.AddJob("* * * * *", func() {
		environment_variables.EnvironmentVariables.LoadFromEnv()
	})
}
