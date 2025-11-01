package mcp

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"menlo.ai/indigo-api-gateway/app/domain/auth"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/responses/openai"
	mcpimpl "menlo.ai/indigo-api-gateway/app/interfaces/http/routes/v1/mcp/mcp_impl"
	"menlo.ai/indigo-api-gateway/app/utils/ptr"
)

func MCPMethodGuard(allowedMethods map[string]bool, api *MCPAPI) gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		var req struct {
			Method string                 `json:"method"`
			Params map[string]interface{} `json:"params"`
		}

		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			c.Abort()
			return
		}

		if !allowedMethods[req.Method] {
			c.Abort()
			return
		}

		if api != nil {
			var userID string
			if id, ok := auth.GetUserIDFromContext(c); ok {
				userID = id
			}
			tool := extractToolName(req.Method, req.Params)
			api.appendActivity(userID, req.Method, tool)
		}

		c.Next()
	}
}

func extractToolName(method string, params map[string]interface{}) string {
	if method != "tools/call" {
		return ""
	}
	if params == nil {
		return ""
	}
	if raw, ok := params["name"]; ok {
		if value, okCast := raw.(string); okCast {
			return value
		}
	}
	if raw, ok := params["tool"]; ok {
		if value, okCast := raw.(string); okCast {
			return value
		}
	}
	return ""
}

type MCPAPI struct {
	SerperMCP   *mcpimpl.SerperMCP
	MCPServer   *mcpserver.MCPServer
	authService *auth.AuthService
	activityLog []*MCPActivityResponse
	activityMu  sync.Mutex
}

type MCPActivityResponse struct {
	Object    string  `json:"object"`
	Method    string  `json:"method"`
	Tool      *string `json:"tool,omitempty"`
	UserID    *string `json:"user_id,omitempty"`
	CreatedAt int64   `json:"created_at"`
}

func NewMCPAPI(serperMCP *mcpimpl.SerperMCP, authService *auth.AuthService) *MCPAPI {
	mcpSrv := mcpserver.NewMCPServer("demo", "0.1.0",
		mcpserver.WithToolCapabilities(true),
		mcpserver.WithRecovery(),
	)
	return &MCPAPI{
		SerperMCP:   serperMCP,
		MCPServer:   mcpSrv,
		authService: authService,
	}
}

// MCPStream
// @Summary MCP streamable endpoint
// @Description Handles Model Context Protocol (MCP) requests over an HTTP stream. The response is sent as a continuous stream of data.
// @Tags Chat Completions API
// @Security BearerAuth
// @Accept json
// @Produce text/event-stream
// @Param request body any true "MCP request payload"
// @Success 200 {string} string "Streamed response (SSE or chunked transfer)"
// @Router /v1/mcp [post]
func (mcpAPI *MCPAPI) RegisterRouter(router *gin.RouterGroup) {
	mcpAPI.SerperMCP.RegisterTool(mcpAPI.MCPServer)

	mcpHttpHandler := mcpserver.NewStreamableHTTPServer(mcpAPI.MCPServer)
	router.Any(
		"/mcp",
		mcpAPI.authService.AppUserAuthMiddleware(),
		MCPMethodGuard(map[string]bool{
			// Initialization / handshake
			"initialize":                true,
			"notifications/initialized": true,
			"ping":                      true,

			// Tools
			"tools/list": true,
			"tools/call": true,

			// Prompts
			"prompts/list": true,
			"prompts/call": true,

			// Resources
			"resources/list":           true,
			"resources/templates/list": true,
			"resources/read":           true,

			// If you support subscription:
			"resources/subscribe": true,
		}, mcpAPI),
		gin.WrapH(mcpHttpHandler))

	router.GET(
		"/mcp/activity",
		mcpAPI.authService.AdminUserAuthMiddleware(),
		mcpAPI.GetActivity,
	)
}

// GetActivity godoc
// @Summary List MCP activity
// @Description Returns recent Model Context Protocol interactions captured by the gateway.
// @Tags Administration API
// @Security BearerAuth
// @Success 200 {object} openai.ListResponse[*MCPActivityResponse]
// @Router /v1/mcp/activity [get]
func (mcpAPI *MCPAPI) GetActivity(reqCtx *gin.Context) {
	mcpAPI.activityMu.Lock()
	data := make([]*MCPActivityResponse, len(mcpAPI.activityLog))
	copy(data, mcpAPI.activityLog)
	mcpAPI.activityMu.Unlock()
	reqCtx.JSON(200, openai.ListResponse[*MCPActivityResponse]{
		Object: "list",
		Data:   data,
		Total:  int64(len(data)),
	})
}

func (mcpAPI *MCPAPI) appendActivity(userID string, method string, tool string) {
	entry := &MCPActivityResponse{
		Object:    "mcp.activity",
		Method:    method,
		CreatedAt: time.Now().Unix(),
	}
	if tool != "" {
		entry.Tool = ptr.ToString(tool)
	}
	if userID != "" {
		entry.UserID = ptr.ToString(userID)
	}
	mcpAPI.activityMu.Lock()
	defer mcpAPI.activityMu.Unlock()
	mcpAPI.activityLog = append(mcpAPI.activityLog, entry)
	const maxEntries = 200
	if len(mcpAPI.activityLog) > maxEntries {
		mcpAPI.activityLog = mcpAPI.activityLog[len(mcpAPI.activityLog)-maxEntries:]
	}
}
