package routes

import (
	"github.com/google/wire"
	v1 "menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/auth"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/auth/google"
	chat "menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/chat"
	conv_chat "menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/conv"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/conversations"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/mcp"
	mcp_impl "menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/mcp/mcp_impl"
	modelroute "menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/model"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/organization"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/organization/invites"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/organization/projects"
	api_keys "menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/organization/projects/api_keys"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/responses"
)

var RouteProvider = wire.NewSet(
	google.NewGoogleAuthAPI,
	auth.NewAuthRoute,
	projects.NewProjectsRoute,
	organization.NewAdminApiKeyAPI,
	organization.NewModelProviderRoute,
	organization.NewOrganizationRoute,
	mcp_impl.NewSerperMCP,
	chat.NewChatRoute,
	chat.NewCompletionAPI,
	conv_chat.NewConvChatRoute,
	conv_chat.NewConvCompletionAPI,
	conv_chat.NewConvMCPAPI,
	conv_chat.NewCompletionNonStreamHandler,
	conv_chat.NewCompletionStreamHandler,
	conv_chat.NewWorkspaceRoute,
	mcp.NewMCPAPI,
	modelroute.NewModelAPI,
	modelroute.NewProvidersAPI,
	responses.NewResponseRoute,
	v1.NewV1Route,
	conversations.NewConversationAPI,
	invites.NewInvitesRoute,
	api_keys.NewProjectApiKeyRoute,
)
