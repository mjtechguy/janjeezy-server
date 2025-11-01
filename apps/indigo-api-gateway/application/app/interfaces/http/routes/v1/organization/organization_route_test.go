package organization

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	domainauth "menlo.ai/indigo-api-gateway/app/domain/auth"
	"menlo.ai/indigo-api-gateway/app/domain/organization"
	"menlo.ai/indigo-api-gateway/app/domain/query"
	"menlo.ai/indigo-api-gateway/app/domain/user"
	"menlo.ai/indigo-api-gateway/app/interfaces/http/responses/openai"
)

type memoryOrganizationRepo struct {
	mu            sync.Mutex
	nextOrgID     uint
	nextMemberID  uint
	orgs          map[uint]*organization.Organization
	orgByPublicID map[string]uint
	members       []*organization.OrganizationMember
}

func newMemoryOrganizationRepo() *memoryOrganizationRepo {
	return &memoryOrganizationRepo{
		orgs:          make(map[uint]*organization.Organization),
		orgByPublicID: make(map[string]uint),
	}
}

func cloneOrganizationEntity(src *organization.Organization) *organization.Organization {
	if src == nil {
		return nil
	}
	cp := *src
	return &cp
}

func cloneMemberEntity(src *organization.OrganizationMember) *organization.OrganizationMember {
	if src == nil {
		return nil
	}
	cp := *src
	return &cp
}

func (m *memoryOrganizationRepo) Create(ctx context.Context, o *organization.Organization) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextOrgID++
	copy := cloneOrganizationEntity(o)
	if copy.ID == 0 {
		copy.ID = m.nextOrgID
	}
	if copy.PublicID == "" {
		copy.PublicID = strings.ReplaceAll(strings.ToLower(copy.Name), " ", "-")
	}
	if copy.CreatedAt.IsZero() {
		copy.CreatedAt = time.Now().UTC()
	}
	m.orgs[copy.ID] = copy
	m.orgByPublicID[copy.PublicID] = copy.ID
	*o = *cloneOrganizationEntity(copy)
	return nil
}

func (m *memoryOrganizationRepo) Update(ctx context.Context, o *organization.Organization) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	copy := cloneOrganizationEntity(o)
	m.orgs[copy.ID] = copy
	m.orgByPublicID[copy.PublicID] = copy.ID
	return nil
}

func (m *memoryOrganizationRepo) DeleteByID(ctx context.Context, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if org, ok := m.orgs[id]; ok {
		delete(m.orgByPublicID, org.PublicID)
	}
	delete(m.orgs, id)
	return nil
}

func (m *memoryOrganizationRepo) FindByID(ctx context.Context, id uint) (*organization.Organization, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneOrganizationEntity(m.orgs[id]), nil
}

func (m *memoryOrganizationRepo) FindByPublicID(ctx context.Context, publicID string) (*organization.Organization, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if id, ok := m.orgByPublicID[publicID]; ok {
		return cloneOrganizationEntity(m.orgs[id]), nil
	}
	return nil, nil
}

func (m *memoryOrganizationRepo) FindByFilter(ctx context.Context, filter organization.OrganizationFilter, _ *query.Pagination) ([]*organization.Organization, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*organization.Organization
	for _, org := range m.orgs {
		if filter.PublicID != nil && org.PublicID != *filter.PublicID {
			continue
		}
		if filter.Enabled != nil && org.Enabled != *filter.Enabled {
			continue
		}
		result = append(result, cloneOrganizationEntity(org))
	}
	return result, nil
}

func (m *memoryOrganizationRepo) Count(ctx context.Context, filter organization.OrganizationFilter) (int64, error) {
	items, _ := m.FindByFilter(ctx, filter, nil)
	return int64(len(items)), nil
}

func (m *memoryOrganizationRepo) AddMember(ctx context.Context, member *organization.OrganizationMember) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextMemberID++
	copy := cloneMemberEntity(member)
	if copy.ID == 0 {
		copy.ID = m.nextMemberID
	}
	if copy.CreatedAt.IsZero() {
		copy.CreatedAt = time.Now().UTC()
	}
	m.members = append(m.members, copy)
	*member = *cloneMemberEntity(copy)
	return nil
}

func (m *memoryOrganizationRepo) FindMemberByFilter(ctx context.Context, filter organization.OrganizationMemberFilter, p *query.Pagination) ([]*organization.OrganizationMember, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var filtered []*organization.OrganizationMember
	for _, member := range m.members {
		if filter.OrganizationID != nil && member.OrganizationID != *filter.OrganizationID {
			continue
		}
		if filter.UserID != nil && member.UserID != *filter.UserID {
			continue
		}
		if filter.Role != nil && string(member.Role) != *filter.Role {
			continue
		}
		filtered = append(filtered, cloneMemberEntity(member))
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ID < filtered[j].ID
	})
	if p != nil {
		if p.After != nil {
			cursor := *p.After
			var withoutCursor []*organization.OrganizationMember
			for _, member := range filtered {
				if member.ID > cursor {
					withoutCursor = append(withoutCursor, member)
				}
			}
			filtered = withoutCursor
		}
		if p.Limit != nil && len(filtered) > *p.Limit {
			filtered = filtered[:*p.Limit]
		}
	}
	return filtered, nil
}

func (m *memoryOrganizationRepo) CountMembers(ctx context.Context, filter organization.OrganizationMemberFilter) (int64, error) {
	items, _ := m.FindMemberByFilter(ctx, filter, nil)
	return int64(len(items)), nil
}

func (m *memoryOrganizationRepo) UpdateMemberRole(ctx context.Context, organizationID uint, userID uint, role organization.OrganizationMemberRole) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, member := range m.members {
		if member.OrganizationID == organizationID && member.UserID == userID {
			member.Role = role
		}
	}
	return nil
}

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

func cloneTestUser(u *user.User) *user.User {
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
	copy := cloneTestUser(u)
	copy.ID = m.nextID
	if copy.CreatedAt.IsZero() {
		copy.CreatedAt = time.Now().UTC()
	}
	m.byID[copy.ID] = copy
	m.byPublic[copy.PublicID] = copy.ID
	m.byEmail[strings.ToLower(copy.Email)] = copy.ID
	*u = *cloneTestUser(copy)
	return nil
}

func (m *memoryUserRepo) Update(ctx context.Context, u *user.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	copy := cloneTestUser(u)
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
		matches = append(matches, cloneTestUser(entry))
	}
	return matches, nil
}

func (m *memoryUserRepo) FindByID(ctx context.Context, id uint) (*user.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneTestUser(m.byID[id]), nil
}

func seedOrganizationRoute(t *testing.T) (*OrganizationRoute, *organization.Organization, *user.User, *user.User, *memoryOrganizationRepo, *organization.OrganizationService, *user.UserService, context.Context) {
	t.Helper()
	ctx := context.Background()

	orgRepo := newMemoryOrganizationRepo()
	orgService := organization.NewService(orgRepo)
	userRepo := newMemoryUserRepo()
	userService := user.NewService(userRepo, nil)

	orgEntity := &organization.Organization{
		Name:    "Acme",
		Enabled: true,
	}
	if err := orgRepo.Create(ctx, orgEntity); err != nil {
		t.Fatalf("seed organization: %v", err)
	}

	owner, err := userService.RegisterUser(ctx, &user.User{
		Name:    "Owner User",
		Email:   "owner@example.com",
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("seed owner user: %v", err)
	}
	reader, err := userService.RegisterUser(ctx, &user.User{
		Name:    "Reader User",
		Email:   "reader@example.com",
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("seed reader user: %v", err)
	}

	if err := orgRepo.AddMember(ctx, &organization.OrganizationMember{
		OrganizationID: orgEntity.ID,
		UserID:         owner.ID,
		Role:           organization.OrganizationMemberRoleOwner,
		CreatedAt:      time.Now().UTC().Add(-2 * time.Hour),
	}); err != nil {
		t.Fatalf("add owner member: %v", err)
	}
	if err := orgRepo.AddMember(ctx, &organization.OrganizationMember{
		OrganizationID: orgEntity.ID,
		UserID:         reader.ID,
		Role:           organization.OrganizationMemberRoleReader,
		CreatedAt:      time.Now().UTC().Add(-time.Hour),
	}); err != nil {
		t.Fatalf("add reader member: %v", err)
	}

	route := &OrganizationRoute{
		organizationSvc: orgService,
		userService:     userService,
	}

	return route, orgEntity, owner, reader, orgRepo, orgService, userService, ctx
}

func TestOrganizationRoute_GetProviderVendors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	route := &OrganizationRoute{}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	route.GetProviderVendors(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Data []ProviderVendorResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Data) == 0 {
		t.Fatalf("expected at least one vendor entry")
	}
}

func TestOrganizationRoute_ListMembers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	route, orgEntity, owner, _, _, _, _, ctx := seedOrganizationRoute(t)

	recorder := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/v1/organization/members?limit=1", nil)
	req = req.WithContext(ctx)
	ginCtx.Request = req
	domainauth.SetAdminOrganizationToContext(ginCtx, orgEntity)

	route.ListMembers(ginCtx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response openai.ListResponse[*OrganizationMemberResponse]
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Data) != 1 {
		t.Fatalf("expected 1 member entry, got %d", len(response.Data))
	}
	if response.Total != 2 {
		t.Fatalf("expected total 2 members, got %d", response.Total)
	}
	if response.FirstID == nil || *response.FirstID == "" {
		t.Fatal("expected first cursor identifier")
	}
	if response.LastID == nil || *response.LastID == "" {
		t.Fatal("expected last cursor identifier")
	}
	member := response.Data[0]
	if member.User.Email != owner.Email {
		t.Fatalf("expected owner email %s, got %s", owner.Email, member.User.Email)
	}
	if member.Role != string(organization.OrganizationMemberRoleOwner) {
		t.Fatalf("expected role owner, got %s", member.Role)
	}
}

func TestOrganizationRoute_UpdateMemberRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	route, orgEntity, owner, _, _, orgService, _, ctx := seedOrganizationRoute(t)

	recorder := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(recorder)
	payload := map[string]string{"role": "reader"}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	req := httptest.NewRequest(http.MethodPatch, "/v1/organization/members/"+owner.PublicID, bytes.NewReader(body))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	ginCtx.Request = req
	ginCtx.Params = gin.Params{{Key: "user_public_id", Value: owner.PublicID}}
	domainauth.SetAdminOrganizationToContext(ginCtx, orgEntity)

	route.UpdateMemberRole(ginCtx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response OrganizationMemberResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Role != string(organization.OrganizationMemberRoleReader) {
		t.Fatalf("expected response role reader, got %s", response.Role)
	}
	if response.User.ID != owner.PublicID {
		t.Fatalf("expected response user id %s, got %s", owner.PublicID, response.User.ID)
	}

	member, err := orgService.FindOneMemberByFilter(ctx, organization.OrganizationMemberFilter{
		OrganizationID: &orgEntity.ID,
		UserID:         &owner.ID,
	})
	if err != nil {
		t.Fatalf("lookup member: %v", err)
	}
	if member == nil || member.Role != organization.OrganizationMemberRoleReader {
		t.Fatalf("expected member role reader, got %v", member)
	}
}
