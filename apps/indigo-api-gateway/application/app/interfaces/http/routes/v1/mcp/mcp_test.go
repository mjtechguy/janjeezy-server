package mcp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMCPMethodGuardAppendsActivity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &MCPAPI{}
	guard := MCPMethodGuard(map[string]bool{
		"ping":       true,
		"tools/call": true,
	}, api)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req, _ := http.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(`{"method":"ping"}`))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	guard(ctx)

	if len(api.activityLog) != 1 {
		t.Fatalf("expected 1 activity entry, got %d", len(api.activityLog))
	}
	if api.activityLog[0].Method != "ping" {
		t.Fatalf("expected method ping, got %s", api.activityLog[0].Method)
	}
	if api.activityLog[0].UserID != nil {
		t.Fatalf("expected no user for synthetic request")
	}

	recorder2 := httptest.NewRecorder()
	ctx2, _ := gin.CreateTestContext(recorder2)
	req2, _ := http.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(`{"method":"tools/call","params":{"name":"google_search"}}`))
	req2.Header.Set("Content-Type", "application/json")
	ctx2.Request = req2

	guard(ctx2)

	if len(api.activityLog) != 2 {
		t.Fatalf("expected 2 activity entries, got %d", len(api.activityLog))
	}
	if api.activityLog[1].Tool == nil || *api.activityLog[1].Tool != "google_search" {
		if api.activityLog[1].Tool == nil {
			t.Fatalf("expected tool google_search, got nil")
		}
		t.Fatalf("expected tool google_search, got %s", *api.activityLog[1].Tool)
	}
}

func TestMCPAPIGetActivity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &MCPAPI{}
	api.appendActivity("user-123", "ping", "")
	api.appendActivity("", "tools/list", "")

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	api.GetActivity(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}
