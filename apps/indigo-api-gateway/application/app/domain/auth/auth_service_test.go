package auth_test

import (
	context "context"
	"errors"
	"strings"
	"sync"
	"testing"

	"menlo.ai/indigo-api-gateway/app/domain/auth"
	"menlo.ai/indigo-api-gateway/app/domain/query"
	"menlo.ai/indigo-api-gateway/app/domain/user"
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
	clone := *u
	return &clone
}

func (m *memoryUserRepo) Create(ctx context.Context, u *user.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	copy := cloneUser(u)
	copy.ID = m.nextID
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
	items, err := m.FindByFilter(ctx, filter, nil)
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return items[0], nil
}

func (m *memoryUserRepo) FindByFilter(ctx context.Context, filter user.UserFilter, p *query.Pagination) ([]*user.User, error) {
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

func TestAuthenticateLocalUser(t *testing.T) {
	repo := newMemoryUserRepo()
	userService := user.NewService(repo, nil)
	authService := auth.NewAuthService(userService, nil, nil, nil, nil)
	ctx := context.Background()

	u, err := userService.RegisterUser(ctx, &user.User{
		Name:    "Admin",
		Email:   "Owner@Example.com",
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("register user: %v", err)
	}

	if err := authService.SetUserPassword(ctx, u, "super-secret-pass"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	authenticated, err := authService.AuthenticateLocalUser(ctx, "OWNER@example.com", "super-secret-pass")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if authenticated.PublicID != u.PublicID {
		t.Fatalf("expected public id %s, got %s", u.PublicID, authenticated.PublicID)
	}

	if _, err := authService.AuthenticateLocalUser(ctx, "owner@example.com", "wrong-pass"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}

	u.Enabled = false
	if _, err := userService.UpdateUser(ctx, u); err != nil {
		t.Fatalf("update user: %v", err)
	}
	if _, err := authService.AuthenticateLocalUser(ctx, "owner@example.com", "super-secret-pass"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials for disabled user, got %v", err)
	}
}
