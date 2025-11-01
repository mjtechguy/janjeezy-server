package organization

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"menlo.ai/indigo-api-gateway/app/domain/auth"
	"menlo.ai/indigo-api-gateway/app/domain/invite"
	"menlo.ai/indigo-api-gateway/app/domain/model"
	"menlo.ai/indigo-api-gateway/app/domain/organization"
	"menlo.ai/indigo-api-gateway/app/domain/project"
	"menlo.ai/indigo-api-gateway/app/domain/query"
	"menlo.ai/indigo-api-gateway/app/domain/settings"
	"menlo.ai/indigo-api-gateway/app/domain/user"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/responses"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/responses/openai"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/organization/invites"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/organization/projects"
	"menlo.ai/indigo-api-gateway/app/utils/ptr"
	environment_variables "menlo.ai/indigo-api-gateway/config/environment_variables"
)

type OrganizationRoute struct {
	adminApiKeyAPI     *AdminApiKeyAPI
	projectsRoute      *projects.ProjectsRoute
	inviteRoute        *invites.InvitesRoute
	modelProviderRoute *ModelProviderRoute
	authService        *auth.AuthService
	organizationSvc    *organization.OrganizationService
	projectService     *project.ProjectService
	inviteService      *invite.InviteService
	providerRegistry   *model.ProviderRegistryService
	userService        *user.UserService
	settingsService    *settings.Service
	auditService       *settings.AuditService
}

func NewOrganizationRoute(
	adminApiKeyAPI *AdminApiKeyAPI,
	projectsRoute *projects.ProjectsRoute,
	inviteRoute *invites.InvitesRoute,
	modelProviderRoute *ModelProviderRoute,
	authService *auth.AuthService,
	organizationSvc *organization.OrganizationService,
	projectService *project.ProjectService,
	inviteService *invite.InviteService,
	providerRegistry *model.ProviderRegistryService,
	userService *user.UserService,
	settingsService *settings.Service,
	auditService *settings.AuditService,
) *OrganizationRoute {
	return &OrganizationRoute{
		adminApiKeyAPI:     adminApiKeyAPI,
		projectsRoute:      projectsRoute,
		inviteRoute:        inviteRoute,
		modelProviderRoute: modelProviderRoute,
		authService:        authService,
		organizationSvc:    organizationSvc,
		projectService:     projectService,
		inviteService:      inviteService,
		providerRegistry:   providerRegistry,
		userService:        userService,
		settingsService:    settingsService,
		auditService:       auditService,
	}
}

type overviewProjects struct {
	Total    int `json:"total"`
	Active   int `json:"active"`
	Archived int `json:"archived"`
}

type overviewMembers struct {
	Total int `json:"total"`
}

type overviewInvites struct {
	Pending int `json:"pending"`
}

type overviewProviders struct {
	Active   int `json:"active"`
	Inactive int `json:"inactive"`
}

type OrganizationOverviewResponse struct {
	Object    string            `json:"object" example:"organization.overview"`
	Projects  overviewProjects  `json:"projects"`
	Members   overviewMembers   `json:"members"`
	Invites   overviewInvites   `json:"invites"`
	Providers overviewProviders `json:"providers"`
}

type smtpSettingsResponse struct {
	Object      string `json:"object"`
	Enabled     bool   `json:"enabled"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	FromEmail   string `json:"from_email"`
	HasPassword bool   `json:"has_password"`
}

type updateSMTPSettingsRequest struct {
	Enabled   bool    `json:"enabled"`
	Host      string  `json:"host"`
	Port      int     `json:"port"`
	Username  string  `json:"username"`
	Password  *string `json:"password,omitempty"`
	FromEmail string  `json:"from_email"`
}

type workspaceQuotaOverrideResponse struct {
	UserPublicID string `json:"user_public_id"`
	Limit        int    `json:"limit"`
}

type workspaceQuotaResponse struct {
	Object       string                           `json:"object"`
	DefaultLimit int                              `json:"default_limit"`
	Overrides    []workspaceQuotaOverrideResponse `json:"overrides"`
}

type updateWorkspaceQuotaRequest struct {
	DefaultLimit int                              `json:"default_limit"`
	Overrides    []workspaceQuotaOverrideResponse `json:"overrides"`
}

type auditLogResponse struct {
	Object    string                 `json:"object"`
	ID        uint                   `json:"id"`
	Event     string                 `json:"event"`
	UserEmail *string                `json:"user_email,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

type OrganizationMemberUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt int64  `json:"created_at"`
}

type OrganizationMemberResponse struct {
	Object   string                 `json:"object"`
	Role     string                 `json:"role"`
	JoinedAt int64                  `json:"joined_at"`
	User     OrganizationMemberUser `json:"user"`
}

type UpdateOrganizationMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=owner reader"`
}

type ProviderVendorResponse struct {
	Key            string  `json:"key"`
	Name           string  `json:"name"`
	Scope          string  `json:"scope"`
	DefaultBaseURL *string `json:"default_base_url,omitempty"`
	CredentialHint *string `json:"credential_hint,omitempty"`
}

type providerVendorUIConfig struct {
	Scope          string
	DefaultBaseURL string
	CredentialHint string
}

var providerVendorCatalog = map[model.ProviderKind]providerVendorUIConfig{
	model.ProviderJan: {
		Scope:          "jan",
		DefaultBaseURL: "https://api.jan.ai/v1",
	},
	model.ProviderOpenAI: {
		DefaultBaseURL: "https://api.openai.com/v1",
		CredentialHint: "sk-...",
	},
	model.ProviderOpenRouter: {
		DefaultBaseURL: "https://openrouter.ai/api/v1",
		CredentialHint: "sk-or-...",
	},
	model.ProviderAnthropic: {
		DefaultBaseURL: "https://api.anthropic.com/v1",
		CredentialHint: "sk-ant-...",
	},
	model.ProviderGemini: {
		DefaultBaseURL: "https://generativelanguage.googleapis.com/v1beta",
		CredentialHint: "AIza...",
	},
	model.ProviderMistral: {
		DefaultBaseURL: "https://api.mistral.ai/v1",
		CredentialHint: "mistral-...",
	},
	model.ProviderGroq: {
		DefaultBaseURL: "https://api.groq.com/openai/v1",
		CredentialHint: "gsk_...",
	},
	model.ProviderCohere: {
		DefaultBaseURL: "https://api.cohere.com/v1",
		CredentialHint: "co-...",
	},
	model.ProviderPerplexity: {
		DefaultBaseURL: "https://api.perplexity.ai",
		CredentialHint: "ppx-...",
	},
	model.ProviderTogetherAI: {
		DefaultBaseURL: "https://api.together.xyz/v1",
		CredentialHint: "tg-...",
	},
	model.ProviderDeepInfra: {
		DefaultBaseURL: "https://api.deepinfra.com/v1/openai",
		CredentialHint: "di-...",
	},
	model.ProviderOllama: {
		DefaultBaseURL: "http://localhost:11434",
	},
	model.ProviderReplicate: {
		DefaultBaseURL: "https://api.replicate.com/v1",
		CredentialHint: "r8_...",
	},
	model.ProviderAzureOpenAI: {
		Scope: "organization",
	},
	model.ProviderAWSBedrock: {
		Scope: "organization",
	},
	model.ProviderHuggingFace: {
		DefaultBaseURL: "https://api-inference.huggingface.co/models",
		CredentialHint: "hf_...",
	},
	model.ProviderVercelAI: {
		DefaultBaseURL: "https://api.vercel.ai/v1",
		CredentialHint: "vercel_...",
	},
	model.ProviderCustom: {
		Scope: "organization",
	},
}

func (organizationRoute *OrganizationRoute) RegisterRouter(router gin.IRouter) {
	organizationRouter := router.Group("/organization")
	organizationRoute.adminApiKeyAPI.RegisterRouter(organizationRouter)
	organizationRoute.projectsRoute.RegisterRouter(organizationRouter)
	organizationRoute.inviteRoute.RegisterRouter(organizationRouter)
	organizationRoute.modelProviderRoute.RegisterRouter(organizationRouter)

	permissionAll := organizationRoute.authService.OrganizationMemberRoleMiddleware(auth.OrganizationMemberRuleAll)
	permissionOwnerOnly := organizationRoute.authService.OrganizationMemberRoleMiddleware(auth.OrganizationMemberRuleOwnerOnly)
	organizationRouter.GET("/overview",
		organizationRoute.authService.AdminUserAuthMiddleware(),
		organizationRoute.authService.RegisteredUserMiddleware(),
		permissionAll,
		organizationRoute.GetOverview,
	)
	organizationRouter.GET("/members",
		organizationRoute.authService.AdminUserAuthMiddleware(),
		organizationRoute.authService.RegisteredUserMiddleware(),
		permissionAll,
		organizationRoute.ListMembers,
	)
	organizationRouter.PATCH("/members/:user_public_id",
		organizationRoute.authService.AdminUserAuthMiddleware(),
		organizationRoute.authService.RegisteredUserMiddleware(),
		permissionOwnerOnly,
		organizationRoute.UpdateMemberRole,
	)
	organizationRouter.GET("/providers/vendors",
		organizationRoute.authService.AdminUserAuthMiddleware(),
		organizationRoute.authService.RegisteredUserMiddleware(),
		permissionAll,
		organizationRoute.GetProviderVendors,
	)

	settingsRouter := organizationRouter.Group("/settings",
		organizationRoute.authService.AdminUserAuthMiddleware(),
		organizationRoute.authService.RegisteredUserMiddleware(),
		permissionOwnerOnly,
	)
	settingsRouter.GET("/smtp", organizationRoute.GetSMTPSettings)
	settingsRouter.PUT("/smtp", organizationRoute.UpdateSMTPSettings)
	settingsRouter.GET("/workspace-quotas", organizationRoute.GetWorkspaceQuota)
	settingsRouter.PUT("/workspace-quotas", organizationRoute.UpdateWorkspaceQuota)

	auditRouter := organizationRouter.Group("/audit-logs",
		organizationRoute.authService.AdminUserAuthMiddleware(),
		organizationRoute.authService.RegisteredUserMiddleware(),
		permissionOwnerOnly,
	)
	auditRouter.GET("", organizationRoute.ListAuditLogs)
}

func (organizationRoute *OrganizationRoute) GetOverview(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	orgEntity, ok := auth.GetAdminOrganizationFromContext(reqCtx)
	if !ok {
		reqCtx.AbortWithStatusJSON(http.StatusUnauthorized, responses.ErrorResponse{
			Code: "ad716e24-620d-11ef-99db-33b5f3c73692",
		})
		return
	}

	projects, err := organizationRoute.projectService.Find(ctx, project.ProjectFilter{
		OrganizationID: &orgEntity.ID,
	}, nil)
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "ad7170aa-620d-11ef-aa44-57a4a5df6fcb",
			Error: err.Error(),
		})
		return
	}

	projectIDs := make([]uint, 0, len(projects))
	projectsSummary := overviewProjects{}
	for _, p := range projects {
		if p == nil {
			continue
		}
		projectIDs = append(projectIDs, p.ID)
		projectsSummary.Total++
		if strings.EqualFold(p.Status, string(project.ProjectStatusArchived)) || p.ArchivedAt != nil {
			projectsSummary.Archived++
		} else {
			projectsSummary.Active++
		}
	}

	membersCount, err := organizationRoute.organizationSvc.CountMembers(ctx, organization.OrganizationMemberFilter{
		OrganizationID: &orgEntity.ID,
	})
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "ad7172be-620d-11ef-bc33-1faff31806eb",
			Error: err.Error(),
		})
		return
	}

	inviteCount, err := organizationRoute.inviteService.CountInvites(ctx, invite.InvitesFilter{
		OrganizationID: &orgEntity.ID,
	})
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "ad7173cc-620d-11ef-b7d7-7f6d4d5e3c4d",
			Error: err.Error(),
		})
		return
	}

	providers, err := organizationRoute.providerRegistry.ListAccessibleProviders(ctx, orgEntity.ID, projectIDs)
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "ad7174d6-620d-11ef-9b55-7f4a6de68b6f",
			Error: err.Error(),
		})
		return
	}

	providerSummary := overviewProviders{}
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		if provider.Active {
			providerSummary.Active++
		} else {
			providerSummary.Inactive++
		}
	}

	response := OrganizationOverviewResponse{
		Object:    "organization.overview",
		Projects:  projectsSummary,
		Members:   overviewMembers{Total: int(membersCount)},
		Invites:   overviewInvites{Pending: int(inviteCount)},
		Providers: providerSummary,
	}

	reqCtx.JSON(http.StatusOK, response)
}

// GetProviderVendors godoc
// @Summary List provider vendors
// @Description Returns supported provider vendor catalog for registration forms.
// @Tags Administration API
// @Security BearerAuth
// @Success 200 {object} openai.ListResponse[ProviderVendorResponse]
// @Router /v1/organization/providers/vendors [get]
func (organizationRoute *OrganizationRoute) GetProviderVendors(reqCtx *gin.Context) {
	vendors := make([]ProviderVendorResponse, 0)
	for _, kind := range model.AllProviderKinds() {
		cfg, ok := providerVendorCatalog[kind]
		scope := "organization"
		if ok && cfg.Scope != "" {
			scope = cfg.Scope
		}
		vendor := ProviderVendorResponse{
			Key:   string(kind),
			Name:  model.ProviderKindDisplayName(kind),
			Scope: scope,
		}
		if ok {
			if cfg.DefaultBaseURL != "" {
				vendor.DefaultBaseURL = ptr.ToString(cfg.DefaultBaseURL)
			}
			if cfg.CredentialHint != "" {
				vendor.CredentialHint = ptr.ToString(cfg.CredentialHint)
			}
		}
		vendors = append(vendors, vendor)
	}
	reqCtx.JSON(http.StatusOK, openai.ListResponse[ProviderVendorResponse]{
		Object: "list",
		Data:   vendors,
	})
}

// GetSMTPSettings returns the SMTP configuration for the organization.
func (organizationRoute *OrganizationRoute) GetSMTPSettings(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	orgEntity, ok := auth.GetAdminOrganizationFromContext(reqCtx)
	if !ok {
		return
	}

	settings, err := organizationRoute.settingsService.GetSMTPSettings(ctx, orgEntity.ID)
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "smtp-settings-fetch-failed",
			Error: err.Error(),
		})
		return
	}

	resp := smtpSettingsResponse{
		Object:      "organization.smtp_settings",
		Enabled:     settings.Enabled,
		Host:        settings.Host,
		Port:        settings.Port,
		Username:    settings.Username,
		FromEmail:   settings.FromEmail,
		HasPassword: settings.HasPassword,
	}
	reqCtx.JSON(http.StatusOK, resp)
}

// UpdateSMTPSettings updates the SMTP settings and records an audit log entry.
func (organizationRoute *OrganizationRoute) UpdateSMTPSettings(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	orgEntity, ok := auth.GetAdminOrganizationFromContext(reqCtx)
	if !ok {
		return
	}
	userEntity, ok := auth.GetUserFromContext(reqCtx)
	if !ok {
		return
	}

	var payload updateSMTPSettingsRequest
	if err := reqCtx.ShouldBindJSON(&payload); err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:  "smtp-settings-invalid",
			Error: err.Error(),
		})
		return
	}

	result, err := organizationRoute.settingsService.UpdateSMTPSettings(ctx, orgEntity.ID, settings.UpdateSMTPSettingsInput{
		Enabled:    payload.Enabled,
		Host:       payload.Host,
		Port:       payload.Port,
		Username:   payload.Username,
		Password:   payload.Password,
		FromEmail:  payload.FromEmail,
		ActorID:    ptr.ToUint(userEntity.ID),
		ActorEmail: ptr.ToString(userEntity.Email),
	})
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:  "smtp-settings-update-failed",
			Error: err.Error(),
		})
		return
	}

	// Apply settings to the in-memory environment variables to reflect changes immediately.
	env := &environment_variables.EnvironmentVariables
	env.SMTP_HOST = result.Host
	env.SMTP_PORT = result.Port
	env.SMTP_USERNAME = result.Username
	if payload.Password != nil {
		env.SMTP_PASSWORD = strings.TrimSpace(*payload.Password)
	}
	env.SMTP_SENDER_EMAIL = result.FromEmail

	_ = organizationRoute.auditService.Record(ctx, settings.RecordAuditInput{
		OrganizationID: orgEntity.ID,
		UserID:         ptr.ToUint(userEntity.ID),
		UserEmail:      ptr.ToString(userEntity.Email),
		Event:          "smtp_settings.updated",
		Metadata: map[string]interface{}{
			"host":       result.Host,
			"port":       result.Port,
			"username":   result.Username,
			"from_email": result.FromEmail,
			"enabled":    result.Enabled,
		},
	})

	resp := smtpSettingsResponse{
		Object:      "organization.smtp_settings",
		Enabled:     result.Enabled,
		Host:        result.Host,
		Port:        result.Port,
		Username:    result.Username,
		FromEmail:   result.FromEmail,
		HasPassword: result.HasPassword,
	}
	reqCtx.JSON(http.StatusOK, resp)
}

// GetWorkspaceQuota returns the workspace quota configuration.
func (organizationRoute *OrganizationRoute) GetWorkspaceQuota(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	orgEntity, ok := auth.GetAdminOrganizationFromContext(reqCtx)
	if !ok {
		return
	}

	config, err := organizationRoute.settingsService.GetWorkspaceQuota(ctx, orgEntity.ID)
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "workspace-quota-fetch-failed",
			Error: err.Error(),
		})
		return
	}

	resp := workspaceQuotaResponse{
		Object:       "organization.workspace_quota",
		DefaultLimit: config.DefaultLimit,
		Overrides:    make([]workspaceQuotaOverrideResponse, 0, len(config.Overrides)),
	}
	for _, override := range config.Overrides {
		resp.Overrides = append(resp.Overrides, workspaceQuotaOverrideResponse{
			UserPublicID: override.UserPublicID,
			Limit:        override.Limit,
		})
	}
	reqCtx.JSON(http.StatusOK, resp)
}

// UpdateWorkspaceQuota updates workspace quota configuration.
func (organizationRoute *OrganizationRoute) UpdateWorkspaceQuota(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	orgEntity, ok := auth.GetAdminOrganizationFromContext(reqCtx)
	if !ok {
		return
	}
	userEntity, ok := auth.GetUserFromContext(reqCtx)
	if !ok {
		return
	}

	var payload updateWorkspaceQuotaRequest
	if err := reqCtx.ShouldBindJSON(&payload); err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:  "workspace-quota-invalid",
			Error: err.Error(),
		})
		return
	}

	input := settings.UpdateWorkspaceQuotaInput{
		DefaultLimit: payload.DefaultLimit,
		Overrides:    make([]settings.WorkspaceQuotaOverride, 0, len(payload.Overrides)),
		ActorID:      ptr.ToUint(userEntity.ID),
		ActorEmail:   ptr.ToString(userEntity.Email),
	}
	for _, override := range payload.Overrides {
		input.Overrides = append(input.Overrides, settings.WorkspaceQuotaOverride{
			UserPublicID: override.UserPublicID,
			Limit:        override.Limit,
		})
	}

	config, err := organizationRoute.settingsService.UpdateWorkspaceQuota(ctx, orgEntity.ID, input)
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:  "workspace-quota-update-failed",
			Error: err.Error(),
		})
		return
	}

	_ = organizationRoute.auditService.Record(ctx, settings.RecordAuditInput{
		OrganizationID: orgEntity.ID,
		UserID:         ptr.ToUint(userEntity.ID),
		UserEmail:      ptr.ToString(userEntity.Email),
		Event:          "workspace_quota.updated",
		Metadata: map[string]interface{}{
			"default_limit": config.DefaultLimit,
			"overrides":     config.Overrides,
		},
	})

	resp := workspaceQuotaResponse{
		Object:       "organization.workspace_quota",
		DefaultLimit: config.DefaultLimit,
		Overrides:    make([]workspaceQuotaOverrideResponse, 0, len(config.Overrides)),
	}
	for _, override := range config.Overrides {
		resp.Overrides = append(resp.Overrides, workspaceQuotaOverrideResponse{
			UserPublicID: override.UserPublicID,
			Limit:        override.Limit,
		})
	}
	reqCtx.JSON(http.StatusOK, resp)
}

// ListAuditLogs returns audit log entries for the organization.
func (organizationRoute *OrganizationRoute) ListAuditLogs(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	orgEntity, ok := auth.GetAdminOrganizationFromContext(reqCtx)
	if !ok {
		return
	}

	limitValue, err := strconv.Atoi(reqCtx.DefaultQuery("limit", "20"))
	if err != nil || limitValue <= 0 {
		limitValue = 20
	}
	if limitValue > 100 {
		limitValue = 100
	}

	var afterID *uint
	if cursor := strings.TrimSpace(reqCtx.Query("after")); cursor != "" {
		if idValue, err := strconv.ParseUint(cursor, 10, 64); err == nil {
			afterID = ptr.ToUint(uint(idValue))
		}
	}

	list, err := organizationRoute.auditService.List(ctx, settings.ListAuditLogsInput{
		OrganizationID: orgEntity.ID,
		AfterID:        afterID,
		Limit:          limitValue,
	})
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "audit-log-fetch-failed",
			Error: err.Error(),
		})
		return
	}

	logs := make([]auditLogResponse, 0, len(list.Logs))
	for _, entry := range list.Logs {
		logs = append(logs, auditLogResponse{
			Object:    "organization.audit_log",
			ID:        entry.ID,
			Event:     entry.Event,
			UserEmail: entry.UserEmail,
			Metadata:  entry.Metadata,
			CreatedAt: entry.CreatedAt,
		})
	}

	var firstID, lastID *string
	if len(list.Logs) > 0 {
		first := fmt.Sprintf("%d", list.Logs[0].ID)
		last := fmt.Sprintf("%d", list.Logs[len(list.Logs)-1].ID)
		firstID = &first
		lastID = &last
	}

	hasMore := list.Total > int64(len(list.Logs))
	reqCtx.JSON(http.StatusOK, openai.ListResponse[auditLogResponse]{
		Object:  "list",
		Data:    logs,
		Total:   list.Total,
		FirstID: firstID,
		LastID:  lastID,
		HasMore: hasMore,
	})
}

// ListMembers godoc
// @Summary List organization members
// @Description Retrieves paginated organization member directory including role information.
// @Tags Administration API
// @Security BearerAuth
// @Param limit query int false "The maximum number of items to return" default(20)
// @Param after query string false "Cursor referencing the last member ID from previous page"
// @Success 200 {object} openai.ListResponse[*OrganizationMemberResponse]
// @Router /v1/organization/members [get]
func (organizationRoute *OrganizationRoute) ListMembers(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	orgEntity, ok := auth.GetAdminOrganizationFromContext(reqCtx)
	if !ok {
		return
	}

	pagination, err := query.GetCursorPaginationFromQuery(reqCtx, func(cursor string) (*uint, error) {
		id, parseErr := strconv.ParseUint(cursor, 10, 64)
		if parseErr != nil {
			return nil, parseErr
		}
		memberID := uint(id)
		return &memberID, nil
	})
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:  "f8baf7d4-620f-11ef-95bf-1f5ff0cc0bd1",
			Error: err.Error(),
		})
		return
	}

	filter := organization.OrganizationMemberFilter{
		OrganizationID: &orgEntity.ID,
	}
	members, err := organizationRoute.organizationSvc.FindMembersByFilter(ctx, filter, pagination)
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "f8bafa4e-620f-11ef-8c00-4b956cf3f629",
			Error: err.Error(),
		})
		return
	}

	total, err := organizationRoute.organizationSvc.CountMembers(ctx, filter)
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "f8bafb7e-620f-11ef-92c0-37dc98b77a97",
			Error: err.Error(),
		})
		return
	}

	responsesList := make([]*OrganizationMemberResponse, 0, len(members))
	for _, member := range members {
		if member == nil {
			continue
		}
		userEntity, userErr := organizationRoute.userService.FindByID(ctx, member.UserID)
		if userErr != nil {
			reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
				Code:  "f8bafce8-620f-11ef-89b7-ebb8cbfb9afd",
				Error: userErr.Error(),
			})
			return
		}
		if userEntity == nil {
			continue
		}
		responsesList = append(responsesList, &OrganizationMemberResponse{
			Object:   "organization.member",
			Role:     string(member.Role),
			JoinedAt: member.CreatedAt.Unix(),
			User: OrganizationMemberUser{
				ID:        userEntity.PublicID,
				Name:      userEntity.Name,
				Email:     userEntity.Email,
				CreatedAt: userEntity.CreatedAt.Unix(),
			},
		})
	}

	var firstID *string
	var lastID *string
	if len(members) > 0 {
		first := strconv.FormatUint(uint64(members[0].ID), 10)
		firstID = &first
		last := strconv.FormatUint(uint64(members[len(members)-1].ID), 10)
		lastID = &last
	}

	response := openai.ListResponse[*OrganizationMemberResponse]{
		Object:  "list",
		Data:    responsesList,
		FirstID: firstID,
		LastID:  lastID,
		Total:   total,
	}

	reqCtx.JSON(http.StatusOK, response)
}

// UpdateMemberRole godoc
// @Summary Update organization member role
// @Description Changes the role for an organization member identified by user public ID.
// @Tags Administration API
// @Security BearerAuth
// @Param user_public_id path string true "Public ID of the user"
// @Param request body UpdateOrganizationMemberRoleRequest true "Role update payload"
// @Success 200 {object} OrganizationMemberResponse
// @Router /v1/organization/members/{user_public_id} [patch]
func (organizationRoute *OrganizationRoute) UpdateMemberRole(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	orgEntity, ok := auth.GetAdminOrganizationFromContext(reqCtx)
	if !ok {
		return
	}

	userPublicID := strings.TrimSpace(reqCtx.Param("user_public_id"))
	if userPublicID == "" {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:  "f8bafdfe-620f-11ef-840f-af92b1a9cbad",
			Error: "user id is required",
		})
		return
	}

	var request UpdateOrganizationMemberRoleRequest
	if err := reqCtx.ShouldBindJSON(&request); err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:  "f8baff10-620f-11ef-9a8d-2346f746e79d",
			Error: err.Error(),
		})
		return
	}

	userEntity, err := organizationRoute.userService.FindByPublicID(ctx, userPublicID)
	if err != nil || userEntity == nil {
		reqCtx.AbortWithStatusJSON(http.StatusNotFound, responses.ErrorResponse{
			Code:  "f8bb000e-620f-11ef-a78f-3fba6d4035d7",
			Error: "member not found",
		})
		return
	}

	role := organization.OrganizationMemberRole(request.Role)
	if role != organization.OrganizationMemberRoleOwner && role != organization.OrganizationMemberRoleReader {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:  "f8bb0112-620f-11ef-a44d-271aa7d5efb1",
			Error: "invalid role",
		})
		return
	}

	if err := organizationRoute.organizationSvc.UpdateMemberRole(ctx, orgEntity.ID, userEntity.ID, role); err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "f8bb020a-620f-11ef-b39d-0bcf960b088b",
			Error: err.Error(),
		})
		return
	}

	member, err := organizationRoute.organizationSvc.FindOneMemberByFilter(ctx, organization.OrganizationMemberFilter{
		OrganizationID: &orgEntity.ID,
		UserID:         &userEntity.ID,
	})
	if err != nil || member == nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  "f8bb0310-620f-11ef-9c91-7f94292b850e",
			Error: "member update failed",
		})
		return
	}

	resp := &OrganizationMemberResponse{
		Object:   "organization.member",
		Role:     string(member.Role),
		JoinedAt: member.CreatedAt.Unix(),
		User: OrganizationMemberUser{
			ID:        userEntity.PublicID,
			Name:      userEntity.Name,
			Email:     userEntity.Email,
			CreatedAt: userEntity.CreatedAt.Unix(),
		},
	}

	reqCtx.JSON(http.StatusOK, resp)
}
