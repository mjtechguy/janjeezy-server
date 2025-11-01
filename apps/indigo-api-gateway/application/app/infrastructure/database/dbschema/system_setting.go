package dbschema

import (
	"encoding/json"
	"time"

	"menlo.ai/indigo-api-gateway/app/domain/settings"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database"
)

func init() {
	database.RegisterSchemaForAutoMigrate(SystemSetting{}, AuditLog{})
}

type SystemSetting struct {
	BaseModel
	OrganizationID uint      `gorm:"not null;index:idx_setting_org_key,unique"`
	Key            string    `gorm:"size:128;not null;index:idx_setting_org_key,unique"`
	Payload        []byte    `gorm:"type:jsonb;not null"`
	LastUpdatedBy  *uint     `gorm:"index"`
	UpdatedByEmail *string   `gorm:"size:255"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

func NewSystemSettingFromDomain(setting *settings.SystemSetting) *SystemSetting {
	payload, _ := json.Marshal(setting.Payload)
	return &SystemSetting{
		BaseModel: BaseModel{
			ID: setting.ID,
		},
		OrganizationID: setting.OrganizationID,
		Key:            setting.Key,
		Payload:        payload,
		LastUpdatedBy:  setting.LastUpdatedBy,
		UpdatedByEmail: setting.UpdatedByEmail,
		UpdatedAt:      setting.UpdatedAt,
	}
}

func (s *SystemSetting) ToDomain() *settings.SystemSetting {
	payload := make(map[string]interface{})
	if len(s.Payload) > 0 {
		_ = json.Unmarshal(s.Payload, &payload)
	}
	return &settings.SystemSetting{
		ID:             s.ID,
		OrganizationID: s.OrganizationID,
		Key:            s.Key,
		Payload:        payload,
		LastUpdatedBy:  s.LastUpdatedBy,
		UpdatedByEmail: s.UpdatedByEmail,
		UpdatedAt:      s.UpdatedAt,
		CreatedAt:      s.CreatedAt,
	}
}

type AuditLog struct {
	BaseModel
	OrganizationID uint    `gorm:"not null;index"`
	UserID         *uint   `gorm:"index"`
	UserEmail      *string `gorm:"size:255"`
	Event          string  `gorm:"size:128;not null;index"`
	Metadata       []byte  `gorm:"type:jsonb"`
}

func NewAuditLogFromDomain(entry *settings.AuditLog) *AuditLog {
	metadata, _ := json.Marshal(entry.Metadata)
	return &AuditLog{
		BaseModel: BaseModel{
			ID: entry.ID,
		},
		OrganizationID: entry.OrganizationID,
		UserID:         entry.UserID,
		UserEmail:      entry.UserEmail,
		Event:          entry.Event,
		Metadata:       metadata,
	}
}

func (a *AuditLog) ToDomain() *settings.AuditLog {
	metadata := make(map[string]interface{})
	if len(a.Metadata) > 0 {
		_ = json.Unmarshal(a.Metadata, &metadata)
	}
	return &settings.AuditLog{
		ID:             a.ID,
		OrganizationID: a.OrganizationID,
		UserID:         a.UserID,
		UserEmail:      a.UserEmail,
		Event:          a.Event,
		Metadata:       metadata,
		CreatedAt:      a.CreatedAt,
	}
}
