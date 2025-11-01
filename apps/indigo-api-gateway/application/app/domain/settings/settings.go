package settings

import (
	"context"
	"errors"
	"time"
)

const (
	SettingKeySMTP           = "smtp"
	SettingKeyWorkspaceQuota = "workspace_quota"
)

type SystemSetting struct {
	ID             uint
	OrganizationID uint
	Key            string
	Payload        map[string]interface{}
	LastUpdatedBy  *uint
	UpdatedByEmail *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type SystemSettingRepository interface {
	FindByKey(ctx context.Context, organizationID uint, key string) (*SystemSetting, error)
	Upsert(ctx context.Context, setting *SystemSetting) error
}

type SMTPSettings struct {
	Enabled     bool   `json:"enabled"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password,omitempty"`
	FromEmail   string `json:"from_email"`
	HasPassword bool   `json:"has_password"`
}

type WorkspaceQuotaOverride struct {
	UserPublicID string `json:"user_public_id"`
	Limit        int    `json:"limit"`
}

type WorkspaceQuotaConfig struct {
	DefaultLimit int                      `json:"default_limit"`
	Overrides    []WorkspaceQuotaOverride `json:"overrides"`
}

type AuditLog struct {
	ID             uint                   `json:"id"`
	OrganizationID uint                   `json:"organization_id"`
	UserID         *uint                  `json:"user_id,omitempty"`
	UserEmail      *string                `json:"user_email,omitempty"`
	Event          string                 `json:"event"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
}

type AuditLogRepository interface {
	Create(ctx context.Context, entry *AuditLog) error
	FindByFilter(ctx context.Context, filter AuditLogFilter) ([]*AuditLog, error)
	Count(ctx context.Context, filter AuditLogFilter) (int64, error)
}

type AuditLogFilter struct {
	OrganizationID uint
	AfterID        *uint
	Limit          int
}

var (
	ErrSettingNotFound = errors.New("setting not found")
)
