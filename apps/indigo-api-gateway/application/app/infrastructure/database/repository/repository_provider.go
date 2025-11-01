package repository

import (
	"github.com/google/wire"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/apikeyrepo"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/conversationrepo"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/inviterepo"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/itemrepo"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/modelrepo"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/organizationrepo"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/projectrepo"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/responserepo"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/settingsrepo"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/transaction"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/userrepo"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/workspacerepo"
)

var RepositoryProvider = wire.NewSet(
	inviterepo.NewInviteGormRepository,
	organizationrepo.NewOrganizationGormRepository,
	projectrepo.NewProjectGormRepository,
	apikeyrepo.NewApiKeyGormRepository,
	userrepo.NewUserGormRepository,
	conversationrepo.NewConversationGormRepository,
	itemrepo.NewItemGormRepository,
	modelrepo.NewProviderGormRepository,
	modelrepo.NewProviderModelGormRepository,
	modelrepo.NewModelCatalogGormRepository,
	responserepo.NewResponseGormRepository,
	workspacerepo.NewWorkspaceGormRepository,
	settingsrepo.NewSettingRepository,
	settingsrepo.NewAuditRepository,
	transaction.NewDatabase,
)
