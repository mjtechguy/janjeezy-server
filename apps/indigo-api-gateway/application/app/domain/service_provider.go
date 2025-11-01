package domain

import (
	"github.com/google/wire"
	"menlo.ai/indigo-api-gateway/app/domain/apikey"
	"menlo.ai/indigo-api-gateway/app/domain/auth"
	"menlo.ai/indigo-api-gateway/app/domain/conversation"
	"menlo.ai/indigo-api-gateway/app/domain/cron"
	"menlo.ai/indigo-api-gateway/app/domain/invite"
	"menlo.ai/indigo-api-gateway/app/domain/mcp/serpermcp"
	domainmodel "menlo.ai/indigo-api-gateway/app/domain/model"
	"menlo.ai/indigo-api-gateway/app/domain/organization"
	"menlo.ai/indigo-api-gateway/app/domain/project"
	"menlo.ai/indigo-api-gateway/app/domain/response"
	"menlo.ai/indigo-api-gateway/app/domain/settings"
	"menlo.ai/indigo-api-gateway/app/domain/user"
	"menlo.ai/indigo-api-gateway/app/domain/workspace"
)

var ServiceProvider = wire.NewSet(
	auth.NewAuthService,
	invite.NewInviteService,
	organization.NewService,
	project.NewService,
	apikey.NewService,
	user.NewService,
	conversation.NewService,
	workspace.NewWorkspaceService,
	domainmodel.NewProviderModelService,
	domainmodel.NewModelCatalogService,
	domainmodel.NewProviderRegistryService,
	response.NewResponseService,
	response.NewResponseModelService,
	response.NewStreamModelService,
	response.NewNonStreamModelService,
	serpermcp.NewSerperService,
	cron.NewCronService,
	settings.NewService,
	settings.NewAuditService,
)
