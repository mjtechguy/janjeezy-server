package settingsrepo

import (
    "context"
    "encoding/json"

    "gorm.io/gorm/clause"
    "menlo.ai/indigo-api-gateway/app/domain/settings"
    "menlo.ai/indigo-api-gateway/app/infrastructure/database/dbschema"
    "menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/transaction"
)

type AuditRepository struct {
	db *transaction.Database
}

func NewAuditRepository(db *transaction.Database) settings.AuditLogRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, entry *settings.AuditLog) error {
	db := r.db.GetTx(ctx)
	metadata, _ := json.Marshal(entry.Metadata)
	model := dbschema.AuditLog{
		BaseModel: dbschema.BaseModel{
			ID: entry.ID,
		},
		OrganizationID: entry.OrganizationID,
		UserID:         entry.UserID,
		UserEmail:      entry.UserEmail,
		Event:          entry.Event,
		Metadata:       metadata,
	}
	return db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&model).Error
}

func (r *AuditRepository) FindByFilter(ctx context.Context, filter settings.AuditLogFilter) ([]*settings.AuditLog, error) {
	db := r.db.GetTx(ctx)
	query := db.WithContext(ctx).Model(&dbschema.AuditLog{}).
		Where("organization_id = ?", filter.OrganizationID).
		Order("id DESC")
	if filter.AfterID != nil {
		query = query.Where("id < ?", *filter.AfterID)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	var rows []dbschema.AuditLog
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]*settings.AuditLog, 0, len(rows))
	for i := range rows {
		result = append(result, rows[i].ToDomain())
	}
	return result, nil
}

func (r *AuditRepository) Count(ctx context.Context, filter settings.AuditLogFilter) (int64, error) {
	db := r.db.GetTx(ctx)
	var count int64
	if err := db.WithContext(ctx).Model(&dbschema.AuditLog{}).
		Where("organization_id = ?", filter.OrganizationID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
