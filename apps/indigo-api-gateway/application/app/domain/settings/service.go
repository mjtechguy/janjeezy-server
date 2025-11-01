package settings

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	repo SystemSettingRepository
}

func NewService(repo SystemSettingRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) defaultSMTPSettings() *SMTPSettings {
	return &SMTPSettings{
		Enabled:     false,
		Host:        "",
		Port:        587,
		Username:    "",
		FromEmail:   "",
		HasPassword: false,
	}
}

func (s *Service) GetSMTPSettings(ctx context.Context, organizationID uint) (*SMTPSettings, error) {
	setting, err := s.repo.FindByKey(ctx, organizationID, SettingKeySMTP)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return s.defaultSMTPSettings(), nil
		}
		return nil, err
	}

	payload := setting.Payload
	result := s.defaultSMTPSettings()
	if enabled, ok := payload["enabled"].(bool); ok {
		result.Enabled = enabled
	}
	if host, ok := payload["host"].(string); ok {
		result.Host = host
	}
	if port, ok := payload["port"].(float64); ok {
		result.Port = int(port)
	}
	if username, ok := payload["username"].(string); ok {
		result.Username = username
	}
	if fromEmail, ok := payload["from_email"].(string); ok {
		result.FromEmail = fromEmail
	}
	if hasPassword, ok := payload["has_password"].(bool); ok {
		result.HasPassword = hasPassword
	}
	if password, ok := payload["password"].(string); ok {
		result.Password = password
		if password != "" {
			result.HasPassword = true
		}
	}
	return result, nil
}

type UpdateSMTPSettingsInput struct {
	Enabled    bool
	Host       string
	Port       int
	Username   string
	Password   *string
	FromEmail  string
	ActorID    *uint
	ActorEmail *string
}

func (s *Service) UpdateSMTPSettings(ctx context.Context, organizationID uint, input UpdateSMTPSettingsInput) (*SMTPSettings, error) {
	if strings.TrimSpace(input.Host) == "" {
		return nil, fmt.Errorf("host is required")
	}
	if input.Port <= 0 {
		return nil, fmt.Errorf("port must be positive")
	}
	if strings.TrimSpace(input.FromEmail) == "" {
		return nil, fmt.Errorf("from_email is required")
	}

	existing, err := s.repo.FindByKey(ctx, organizationID, SettingKeySMTP)
	var existingPayload map[string]interface{}
	if err != nil {
		if !errors.Is(err, ErrSettingNotFound) {
			return nil, err
		}
		existingPayload = map[string]interface{}{}
	} else {
		existingPayload = existing.Payload
	}

	var existingPassword string
	hasPassword := false
	if existingPayload != nil {
		if value, ok := existingPayload["password"].(string); ok {
			existingPassword = value
		}
		if value, ok := existingPayload["has_password"].(bool); ok {
			hasPassword = value
		}
	}

	payload := map[string]interface{}{
		"enabled":    input.Enabled,
		"host":       strings.TrimSpace(input.Host),
		"port":       input.Port,
		"username":   strings.TrimSpace(input.Username),
		"from_email": strings.TrimSpace(input.FromEmail),
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}
	if input.Password != nil {
		passwordValue := strings.TrimSpace(*input.Password)
		payload["password"] = passwordValue
		hasPassword = passwordValue != ""
	} else {
		payload["password"] = existingPassword
	}
	payload["has_password"] = hasPassword

	setting := &SystemSetting{
		OrganizationID: organizationID,
		Key:            SettingKeySMTP,
		Payload:        payload,
		LastUpdatedBy:  input.ActorID,
		UpdatedByEmail: input.ActorEmail,
	}

	if err := s.repo.Upsert(ctx, setting); err != nil {
		return nil, err
	}

	result := &SMTPSettings{
		Enabled:     input.Enabled,
		Host:        payload["host"].(string),
		Port:        input.Port,
		Username:    payload["username"].(string),
		FromEmail:   payload["from_email"].(string),
		HasPassword: payload["has_password"].(bool),
	}
	if input.Password != nil {
		result.Password = strings.TrimSpace(*input.Password)
	}
	return result, nil
}

func (s *Service) defaultWorkspaceQuota() *WorkspaceQuotaConfig {
	return &WorkspaceQuotaConfig{
		DefaultLimit: 10,
		Overrides:    []WorkspaceQuotaOverride{},
	}
}

func (s *Service) GetWorkspaceQuota(ctx context.Context, organizationID uint) (*WorkspaceQuotaConfig, error) {
	setting, err := s.repo.FindByKey(ctx, organizationID, SettingKeyWorkspaceQuota)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return s.defaultWorkspaceQuota(), nil
		}
		return nil, err
	}

	payload := setting.Payload
	result := s.defaultWorkspaceQuota()
	if defaultLimit, ok := payload["default_limit"].(float64); ok {
		result.DefaultLimit = int(defaultLimit)
	}
	if overrides, ok := payload["overrides"].([]interface{}); ok {
		configOverrides := make([]WorkspaceQuotaOverride, 0, len(overrides))
		for _, item := range overrides {
			if entry, ok := item.(map[string]interface{}); ok {
				userID, _ := entry["user_public_id"].(string)
				limitVal, _ := entry["limit"].(float64)
				if strings.TrimSpace(userID) == "" || limitVal <= 0 {
					continue
				}
				configOverrides = append(configOverrides, WorkspaceQuotaOverride{
					UserPublicID: userID,
					Limit:        int(limitVal),
				})
			}
		}
		result.Overrides = configOverrides
	}
	return result, nil
}

type UpdateWorkspaceQuotaInput struct {
	DefaultLimit int
	Overrides    []WorkspaceQuotaOverride
	ActorID      *uint
	ActorEmail   *string
}

func (s *Service) UpdateWorkspaceQuota(ctx context.Context, organizationID uint, input UpdateWorkspaceQuotaInput) (*WorkspaceQuotaConfig, error) {
	if input.DefaultLimit <= 0 {
		return nil, fmt.Errorf("default_limit must be positive")
	}
	cleanOverrides := make([]WorkspaceQuotaOverride, 0, len(input.Overrides))
	for _, override := range input.Overrides {
		if strings.TrimSpace(override.UserPublicID) == "" {
			continue
		}
		if override.Limit <= 0 {
			return nil, fmt.Errorf("override limit must be positive")
		}
		cleanOverrides = append(cleanOverrides, WorkspaceQuotaOverride{
			UserPublicID: strings.TrimSpace(override.UserPublicID),
			Limit:        override.Limit,
		})
	}

	payload := map[string]interface{}{
		"default_limit": input.DefaultLimit,
		"overrides":     cleanOverrides,
		"updated_at":    time.Now().UTC().Format(time.RFC3339),
	}

	setting := &SystemSetting{
		OrganizationID: organizationID,
		Key:            SettingKeyWorkspaceQuota,
		Payload:        payload,
		LastUpdatedBy:  input.ActorID,
		UpdatedByEmail: input.ActorEmail,
	}
	if err := s.repo.Upsert(ctx, setting); err != nil {
		return nil, err
	}

	return &WorkspaceQuotaConfig{
		DefaultLimit: input.DefaultLimit,
		Overrides:    cleanOverrides,
	}, nil
}
