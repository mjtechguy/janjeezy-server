package settings

import (
	"context"
	"time"
)

type AuditService struct {
	repo AuditLogRepository
}

func NewAuditService(repo AuditLogRepository) *AuditService {
	return &AuditService{repo: repo}
}

type RecordAuditInput struct {
	OrganizationID uint
	UserID         *uint
	UserEmail      *string
	Event          string
	Metadata       map[string]interface{}
}

func (s *AuditService) Record(ctx context.Context, input RecordAuditInput) error {
	entry := &AuditLog{
		OrganizationID: input.OrganizationID,
		UserID:         input.UserID,
		UserEmail:      input.UserEmail,
		Event:          input.Event,
		Metadata:       input.Metadata,
		CreatedAt:      time.Now(),
	}
	return s.repo.Create(ctx, entry)
}

type ListAuditLogsInput struct {
	OrganizationID uint
	AfterID        *uint
	Limit          int
}

type AuditLogList struct {
	Logs  []*AuditLog
	Total int64
}

func (s *AuditService) List(ctx context.Context, input ListAuditLogsInput) (*AuditLogList, error) {
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 20
	}
	filter := AuditLogFilter{
		OrganizationID: input.OrganizationID,
		AfterID:        input.AfterID,
		Limit:          input.Limit,
	}
	items, err := s.repo.FindByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &AuditLogList{
		Logs:  items,
		Total: total,
	}, nil
}
