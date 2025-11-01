package organization

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"menlo.ai/indigo-api-gateway/app/domain/auth"
	domainmodel "menlo.ai/indigo-api-gateway/app/domain/model"
	"menlo.ai/indigo-api-gateway/app/domain/project"
	"menlo.ai/indigo-api-gateway/app/domain/settings"
	"menlo.ai/indigo-api-gateway/app/infrastructure/inference"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/responses"
	"menlo.ai/indigo-api-gateway/app/utils/ptr"
)

type ModelProviderRoute struct {
	authService       *auth.AuthService
	providerRegistry  *domainmodel.ProviderRegistryService
	inferenceProvider *inference.InferenceProvider
	projectService    *project.ProjectService
	auditService      *settings.AuditService
}

func NewModelProviderRoute(
	authService *auth.AuthService,
	providerRegistry *domainmodel.ProviderRegistryService,
	inferenceProvider *inference.InferenceProvider,
	projectService *project.ProjectService,
	auditService *settings.AuditService,
) *ModelProviderRoute {
	return &ModelProviderRoute{
		authService:       authService,
		providerRegistry:  providerRegistry,
		inferenceProvider: inferenceProvider,
		projectService:    projectService,
		auditService:      auditService,
	}
}

func (route *ModelProviderRoute) RegisterRouter(router *gin.RouterGroup) {
	group := router.Group("/models/providers",
		route.authService.AdminUserAuthMiddleware(),
		route.authService.RegisteredUserMiddleware(),
		route.authService.OrganizationMemberRoleMiddleware(auth.OrganizationMemberRuleOwnerOnly),
	)
	group.POST("", route.registerProvider)
	group.PATCH("/:provider_public_id", route.updateProvider)
	group.POST("/:provider_public_id/sync", route.syncProvider)
}

type registerProviderRequest struct {
	Name     string            `json:"name" binding:"required"`
	Vendor   string            `json:"vendor" binding:"required"`
	BaseURL  string            `json:"base_url" binding:"required"`
	APIKey   string            `json:"api_key"`
	Metadata map[string]string `json:"metadata"`
	Active   *bool             `json:"active"`
	Project  *string           `json:"project_public_id"`
}

type registerProviderResponse struct {
	ID          string                         `json:"id"`
	Slug        string                         `json:"slug"`
	Name        string                         `json:"name"`
	Vendor      string                         `json:"vendor"`
	BaseURL     string                         `json:"base_url"`
	Active      bool                           `json:"active"`
	Metadata    map[string]string              `json:"metadata,omitempty"`
	Scope       string                         `json:"scope"`
	Project     *string                        `json:"project_public_id,omitempty"`
	LastSync    *int64                         `json:"last_synced_at,omitempty"`
	SyncLatency *int64                         `json:"sync_latency_ms,omitempty"`
	APIKeyHint  *string                        `json:"api_key_hint,omitempty"`
	Models      []registerProviderModelSummary `json:"models"`
	ModelsCount int                            `json:"models_count"`
}

type registerProviderModelSummary struct {
	ID            string  `json:"id"`
	ModelKey      string  `json:"model_key"`
	DisplayName   string  `json:"display_name"`
	CatalogID     *string `json:"catalog_id,omitempty"`
	CatalogStatus *string `json:"catalog_status,omitempty"`
	Active        bool    `json:"active"`
	UpdatedAt     int64   `json:"updated_at"`
}

type updateProviderRequest struct {
	Name     *string            `json:"name"`
	BaseURL  *string            `json:"base_url"`
	APIKey   *string            `json:"api_key"`
	Metadata *map[string]string `json:"metadata"`
	Active   *bool              `json:"active"`
}

type providerDetailResponse struct {
	ID         string            `json:"id"`
	Slug       string            `json:"slug"`
	Name       string            `json:"name"`
	Vendor     string            `json:"vendor"`
	BaseURL    string            `json:"base_url"`
	Active     bool              `json:"active"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Scope      string            `json:"scope"`
	Project    *string           `json:"project_public_id,omitempty"`
	LastSync   *int64            `json:"last_synced_at,omitempty"`
	APIKeyHint *string           `json:"api_key_hint,omitempty"`
}

func (route *ModelProviderRoute) registerProvider(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	orgEntity, ok := auth.GetAdminOrganizationFromContext(reqCtx)
	if !ok {
		return
	}

	var request registerProviderRequest
	if err := reqCtx.ShouldBindJSON(&request); err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:          "4dfe7980-9d40-47fb-8cf1-1864dfd1e3eb",
			ErrorInstance: err,
		})
		return
	}

	active := true
	if request.Active != nil {
		active = *request.Active
	}

	var projectID uint
	var projectPublicID *string
	if request.Project != nil && strings.TrimSpace(*request.Project) != "" {
		projectPublic := strings.TrimSpace(*request.Project)
		projectEntity, err := route.projectService.FindOne(ctx, project.ProjectFilter{
			PublicID:       &projectPublic,
			OrganizationID: ptr.ToUint(orgEntity.ID),
		})
		if err != nil || projectEntity == nil {
			reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
				Code:  "9e3b60fc-6211-11ef-8f79-0f0350035c70",
				Error: "invalid project_public_id",
			})
			return
		}
		projectID = projectEntity.ID
		projectPublicID = ptr.ToString(projectEntity.PublicID)
	}

	result, err := route.providerRegistry.RegisterProvider(ctx, domainmodel.RegisterProviderInput{
		OrganizationID: orgEntity.ID,
		ProjectID:      projectID,
		Name:           request.Name,
		Vendor:         request.Vendor,
		BaseURL:        request.BaseURL,
		APIKey:         request.APIKey,
		Metadata:       request.Metadata,
		Active:         active,
	})
	if err != nil {
		status := http.StatusBadRequest
		reqCtx.AbortWithStatusJSON(status, responses.ErrorResponse{
			Code:  err.GetCode(),
			Error: err.GetMessage(),
		})
		return
	}

	syncStarted := time.Now()
	models, fetchErr := route.inferenceProvider.ListModels(ctx, result.Provider)
	if fetchErr != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadGateway, responses.ErrorResponse{
			Code:          "cbe9fb03-a434-4d57-8a59-7b1e6830f9e5",
			ErrorInstance: fetchErr,
		})
		return
	}

	syncResults, syncErr := route.providerRegistry.SyncProviderModels(ctx, result.Provider, models)
	if syncErr != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  syncErr.GetCode(),
			Error: syncErr.GetMessage(),
		})
		return
	}
	result.Models = syncResults
	durationMs := time.Since(syncStarted).Milliseconds()
	var syncLatency *int64
	if durationMs >= 0 {
		syncLatency = ptr.ToInt64(durationMs)
	}

	if projectPublicID == nil && result.Provider.ProjectID != nil {
		if projectEntity, err := route.projectService.FindProjectByID(ctx, *result.Provider.ProjectID); err == nil && projectEntity != nil {
			projectPublicID = ptr.ToString(projectEntity.PublicID)
		}
	}

	resp := toRegisterProviderResponse(result, projectPublicID, syncLatency)
	reqCtx.JSON(http.StatusOK, resp)
}

func (route *ModelProviderRoute) syncProvider(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	orgEntity, ok := auth.GetAdminOrganizationFromContext(reqCtx)
	if !ok {
		return
	}
	userEntity, ok := auth.GetUserFromContext(reqCtx)
	if !ok {
		return
	}

	providerPublicID := strings.TrimSpace(reqCtx.Param("provider_public_id"))
	if providerPublicID == "" {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:  "provider-sync-invalid",
			Error: "provider_public_id is required",
		})
		return
	}

	provider, commonErr := route.providerRegistry.GetProviderByPublicID(ctx, providerPublicID)
	if commonErr != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  commonErr.GetCode(),
			Error: commonErr.GetMessage(),
		})
		return
	}
	if provider == nil {
		reqCtx.AbortWithStatusJSON(http.StatusNotFound, responses.ErrorResponse{
			Code:  "provider-not-found",
			Error: "provider not found",
		})
		return
	}

	if provider.OrganizationID != nil && *provider.OrganizationID != orgEntity.ID {
		reqCtx.AbortWithStatusJSON(http.StatusForbidden, responses.ErrorResponse{
			Code:  "provider-access-denied",
			Error: "provider does not belong to organization",
		})
		return
	}

	syncStarted := time.Now()
	models, fetchErr := route.inferenceProvider.ListModels(ctx, provider)
	if fetchErr != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadGateway, responses.ErrorResponse{
			Code:          "provider-sync-fetch",
			ErrorInstance: fetchErr,
		})
		return
	}

	results, syncErr := route.providerRegistry.SyncProviderModels(ctx, provider, models)
	if syncErr != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responses.ErrorResponse{
			Code:  syncErr.GetCode(),
			Error: syncErr.GetMessage(),
		})
		return
	}

	durationMs := time.Since(syncStarted).Milliseconds()
	var syncLatency *int64
	if durationMs >= 0 {
		syncLatency = ptr.ToInt64(durationMs)
	}

	var projectPublicID *string
	if provider.ProjectID != nil {
		if projectEntity, err := route.projectService.FindProjectByID(ctx, *provider.ProjectID); err == nil && projectEntity != nil {
			projectPublicID = ptr.ToString(projectEntity.PublicID)
		}
	}

	resp := toRegisterProviderResponse(&domainmodel.ProviderRegistrationResult{
		Provider: provider,
		Models:   results,
	}, projectPublicID, syncLatency)

	_ = route.auditService.Record(ctx, settings.RecordAuditInput{
		OrganizationID: orgEntity.ID,
		UserID:         ptr.ToUint(userEntity.ID),
		UserEmail:      ptr.ToString(userEntity.Email),
		Event:          "provider.synced",
		Metadata: map[string]interface{}{
			"provider_id":   provider.PublicID,
			"models_synced": len(results),
			"latency_ms":    durationMs,
		},
	})

	reqCtx.JSON(http.StatusOK, resp)
}

func scopeForProvider(provider *domainmodel.Provider) string {
	if provider.ProjectID != nil {
		return "project"
	}
	if provider.OrganizationID != nil {
		return "organization"
	}
	return "jan"
}

func toRegisterProviderResponse(result *domainmodel.ProviderRegistrationResult, projectPublicID *string, syncLatency *int64) registerProviderResponse {
	provider := result.Provider
	resp := registerProviderResponse{
		ID:          provider.PublicID,
		Slug:        provider.Slug,
		Name:        provider.DisplayName,
		Vendor:      strings.ToLower(string(provider.Kind)),
		BaseURL:     provider.BaseURL,
		Active:      provider.Active,
		Metadata:    provider.Metadata,
		Scope:       scopeForProvider(provider),
		Project:     projectPublicID,
		SyncLatency: syncLatency,
		APIKeyHint:  provider.APIKeyHint,
	}
	if provider.LastSyncedAt != nil {
		timestamp := provider.LastSyncedAt.Unix()
		resp.LastSync = &timestamp
	}

	for _, model := range result.Models {
		item := registerProviderModelSummary{
			ID:          model.ProviderModel.PublicID,
			ModelKey:    model.ProviderModel.ModelKey,
			DisplayName: model.ProviderModel.DisplayName,
			Active:      model.ProviderModel.Active,
			UpdatedAt:   model.ProviderModel.UpdatedAt.Unix(),
		}
		if model.Catalog != nil {
			item.CatalogID = ptr.ToString(model.Catalog.PublicID)
			status := string(model.Catalog.Status)
			item.CatalogStatus = ptr.ToString(status)
		}
		resp.Models = append(resp.Models, item)
	}
	resp.ModelsCount = len(resp.Models)

	return resp
}

func (route *ModelProviderRoute) updateProvider(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	orgEntity, ok := auth.GetAdminOrganizationFromContext(reqCtx)
	if !ok {
		return
	}
	publicID := strings.TrimSpace(reqCtx.Param("provider_public_id"))
	if publicID == "" {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:  "28dd6e4a-b7df-4e75-bb70-2b7f2a44d8ec",
			Error: "provider id is required",
		})
		return
	}

	provider, err := route.providerRegistry.FindByPublicID(ctx, publicID)
	if err != nil {
		status := http.StatusBadRequest
		if err.GetCode() == "d16271bf-54f5-4b25-bbd2-2353f1d5265c" {
			status = http.StatusNotFound
		}
		reqCtx.AbortWithStatusJSON(status, responses.ErrorResponse{
			Code:  err.GetCode(),
			Error: err.GetMessage(),
		})
		return
	}
	if provider.OrganizationID == nil || *provider.OrganizationID != orgEntity.ID {
		reqCtx.AbortWithStatusJSON(http.StatusNotFound, responses.ErrorResponse{
			Code:  "a2b8c03f-4a15-4431-9a0f-0a5c8ef0e83d",
			Error: "provider not found",
		})
		return
	}
	if provider.ProjectID != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:  "4b4ff5ab-6a55-4aa7-842c-9a8d6fd8b061",
			Error: "only organization providers can be updated here",
		})
		return
	}

	var request updateProviderRequest
	if err := reqCtx.ShouldBindJSON(&request); err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:          "f9be18d0-5eac-46e2-8fd6-779b272918aa",
			ErrorInstance: err,
		})
		return
	}

	input := domainmodel.UpdateProviderInput{
		Name:     request.Name,
		BaseURL:  request.BaseURL,
		APIKey:   request.APIKey,
		Metadata: request.Metadata,
		Active:   request.Active,
	}

	updated, updateErr := route.providerRegistry.UpdateProvider(ctx, provider, input)
	if updateErr != nil {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responses.ErrorResponse{
			Code:  updateErr.GetCode(),
			Error: updateErr.GetMessage(),
		})
		return
	}

	var projectPublicID *string
	if updated.ProjectID != nil {
		if projectEntity, err := route.projectService.FindProjectByID(ctx, *updated.ProjectID); err == nil && projectEntity != nil {
			projectPublicID = ptr.ToString(projectEntity.PublicID)
		}
	}
	reqCtx.JSON(http.StatusOK, route.toProviderDetailResponse(updated, projectPublicID))
}

func (route *ModelProviderRoute) toProviderDetailResponse(provider *domainmodel.Provider, projectPublicID *string) providerDetailResponse {
	resp := providerDetailResponse{
		ID:         provider.PublicID,
		Slug:       provider.Slug,
		Name:       provider.DisplayName,
		Vendor:     strings.ToLower(string(provider.Kind)),
		BaseURL:    provider.BaseURL,
		Active:     provider.Active,
		Metadata:   provider.Metadata,
		Scope:      scopeForProvider(provider),
		Project:    projectPublicID,
		APIKeyHint: provider.APIKeyHint,
	}
	if provider.LastSyncedAt != nil {
		timestamp := provider.LastSyncedAt.Unix()
		resp.LastSync = &timestamp
	}
	return resp
}
