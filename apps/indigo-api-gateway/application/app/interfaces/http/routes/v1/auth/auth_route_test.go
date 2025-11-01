package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	domainauth "menlo.ai/indigo-api-gateway/app/domain/auth"
	"menlo.ai/indigo-api-gateway/app/domain/query"
	"menlo.ai/indigo-api-gateway/app/domain/user"
	"menlo.ai/indigo-api-gateway/config/environment_variables"
)

type memoryUserRepo struct {
	mu       sync.Mutex
	nextID   uint
	byID     map[uint]*user.User
	byEmail  map[string]uint
	byPublic map[string]uint
}

func newMemoryUserRepo() *memoryUserRepo {
	return &memoryUserRepo{
		byID:     make(map[uint]*user.User),
		byEmail:  make(map[string]uint),
		byPublic: make(map[string]uint),
	}
}

func cloneUser(u *user.User) *user.User {
	if u == nil {
		return nil
	}
	cp := *u
	return &cp
}

func (m *memoryUserRepo) Create(ctx context.Context, u *user.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	now := time.Now().UTC()
	copy := cloneUser(u)
	copy.ID = m.nextID
	if copy.CreatedAt.IsZero() {
		copy.CreatedAt = now
	}
	m.byID[copy.ID] = copy
	m.byPublic[copy.PublicID] = copy.ID
	m.byEmail[strings.ToLower(copy.Email)] = copy.ID
	*u = *cloneUser(copy)
	return nil
}

func (m *memoryUserRepo) Update(ctx context.Context, u *user.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	copy := cloneUser(u)
	m.byID[copy.ID] = copy
	m.byPublic[copy.PublicID] = copy.ID
	m.byEmail[strings.ToLower(copy.Email)] = copy.ID
	return nil
}

func (m *memoryUserRepo) FindFirst(ctx context.Context, filter user.UserFilter) (*user.User, error) {
	results, err := m.FindByFilter(ctx, filter, nil)
	if err != nil || len(results) == 0 {
		return nil, err
	}
	return results[0], nil
}

func (m *memoryUserRepo) FindByFilter(ctx context.Context, filter user.UserFilter, _ *query.Pagination) ([]*user.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var matches []*user.User
	for _, entry := range m.byID {
		if filter.Email != nil && !strings.EqualFold(entry.Email, *filter.Email) {
			continue
		}
		if filter.PublicID != nil && entry.PublicID != *filter.PublicID {
			continue
		}
		matches = append(matches, cloneUser(entry))
	}
	return matches, nil
}

func (m *memoryUserRepo) FindByID(ctx context.Context, id uint) (*user.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if entry, ok := m.byID[id]; ok {
		return cloneUser(entry), nil
	}
	return nil, nil
}

func TestAuthRouteLocalLogin(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo := newMemoryUserRepo()
	userService := user.NewService(repo, nil)
	authService := domainauth.NewAuthService(userService, nil, nil, nil, nil)
	environment_variables.EnvironmentVariables.JWT_SECRET = []byte("test-secret")

	ctx := context.Background()
	admin, err := userService.RegisterUser(ctx, &user.User{
		Name:    "Admin",
		Email:   "owner@example.com",
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("register user: %v", err)
	}
	if err := authService.SetUserPassword(ctx, admin, "super-secret"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	authRoute := NewAuthRoute(nil, userService, authService)

	t.Run("success", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(recorder)

		payload := map[string]string{
			"email":    "owner@example.com",
			"password": "super-secret",
		}
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/local/login", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		ginCtx.Request = req

		authRoute.LocalLogin(ginCtx)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}

		var response AccessTokenResponse
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.AccessToken == "" {
			t.Fatal("expected access token in response")
		}

		var refreshCookie *http.Cookie
		for _, cookie := range recorder.Result().Cookies() {
			if cookie.Name == domainauth.RefreshTokenKey {
				refreshCookie = cookie
				break
			}
		}
		if refreshCookie == nil {
			t.Fatal("expected refresh token cookie")
		}
		if refreshCookie.Value == "" {
			t.Fatal("expected refresh token cookie value")
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(recorder)

		payload := map[string]string{
			"email":    "owner@example.com",
			"password": "wrong-password",
		}
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/local/login", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		ginCtx.Request = req

		authRoute.LocalLogin(ginCtx)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", recorder.Code)
		}
		var response struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.Error == "" {
			t.Fatal("expected error message for invalid credentials")
		}
	})
}
