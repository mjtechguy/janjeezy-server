package modelroute

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"menlo.ai/indigo-api-gateway/app/domain/auth"
	domainmodel "menlo.ai/indigo-api-gateway/app/domain/model"
	"menlo.ai/indigo-api-gateway/app/domain/project"
	"menlo.ai/indigo-api-gateway/app/utils/ptr"
)

type ProvidersAPI struct {
	authService      *auth.AuthService
	projectService   *project.ProjectService
	providerRegistry *domainmodel.ProviderRegistryService
}

func NewProvidersAPI(authService *auth.AuthService, projectService *project.ProjectService, providerRegistry *domainmodel.ProviderRegistryService) *ProvidersAPI {
	return &ProvidersAPI{
		authService:      authService,
		projectService:   projectService,
		providerRegistry: providerRegistry,
	}
}

func (api *ProvidersAPI) RegisterRouter(router *gin.RouterGroup) {
	group := router.Group("/models/providers",
		api.authService.AppUserAuthMiddleware(),
		api.authService.RegisteredUserMiddleware(),
	)
	group.GET("", api.listProviders)
}

type providerSummary struct {
	ID          string            `json:"id"`
	Slug        string            `json:"slug"`
	Name        string            `json:"name"`
	Vendor      string            `json:"vendor"`
	BaseURL     string            `json:"base_url"`
	Active      bool              `json:"active"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Scope       string            `json:"scope"`
	ProjectID   *string           `json:"project_id,omitempty"`
	LastSync    *int64            `json:"last_synced_at,omitempty"`
	SyncLatency *int64            `json:"sync_latency_ms,omitempty"`
	APIKeyHint  *string           `json:"api_key_hint,omitempty"`
	Models      int               `json:"models_count"`
}

type providersListResponse struct {
	Object string            `json:"object"`
	Data   []providerSummary `json:"data"`
}

func (api *ProvidersAPI) listProviders(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	_, projectPublicIDs, providers, ok := ResolveAccessibleProviders(reqCtx, api.authService, api.projectService, api.providerRegistry)
	if !ok {
		return
	}

	resp := providersListResponse{
		Object: "list",
		Data:   make([]providerSummary, 0, len(providers)),
	}

	scopeOrder := map[string]int{
		"jan":          0,
		"organization": 1,
		"project":      2,
	}

	providerIDs := make([]uint, 0, len(providers))
	for _, provider := range providers {
		if provider != nil {
			providerIDs = append(providerIDs, provider.ID)
		}
	}

	counts := make(map[uint]int)
	if len(providerIDs) > 0 {
		if providerModels, err := api.providerRegistry.ListProviderModels(ctx, providerIDs); err == nil {
			for _, model := range providerModels {
				counts[model.ProviderID]++
			}
		}
	}

	for _, provider := range providers {
		scope := "organization"
		var projectID *string
		if provider.ProjectID != nil {
			if publicID, exists := projectPublicIDs[*provider.ProjectID]; exists {
				projectID = ptr.ToString(publicID)
			}
			scope = "project"
		} else if provider.OrganizationID == nil {
			scope = "jan"
		}

		var lastSync *int64
		if provider.LastSyncedAt != nil {
			timestamp := provider.LastSyncedAt.Unix()
			lastSync = &timestamp
		}

		resp.Data = append(resp.Data, providerSummary{
			ID:         provider.PublicID,
			Slug:       provider.Slug,
			Name:       provider.DisplayName,
			Vendor:     strings.ToLower(string(provider.Kind)),
			BaseURL:    provider.BaseURL,
			Active:     provider.Active,
			Metadata:   provider.Metadata,
			Scope:      scope,
			ProjectID:  projectID,
			LastSync:   lastSync,
			APIKeyHint: provider.APIKeyHint,
			Models:     counts[provider.ID],
		})
	}

	getScopeOrder := func(scope string) int {
		if order, exists := scopeOrder[scope]; exists {
			return order
		}
		return len(scopeOrder)
	}

	sort.Slice(resp.Data, func(i, j int) bool {
		if resp.Data[i].Scope == resp.Data[j].Scope {
			return resp.Data[i].Name < resp.Data[j].Name
		}
		return getScopeOrder(resp.Data[i].Scope) < getScopeOrder(resp.Data[j].Scope)
	})

	reqCtx.JSON(http.StatusOK, resp)
}
