package settingsrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"menlo.ai/indigo-api-gateway/app/domain/settings"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/dbschema"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database/repository/transaction"
)

type SettingRepository struct {
	db *transaction.Database
}

func NewSettingRepository(db *transaction.Database) settings.SystemSettingRepository {
	return &SettingRepository{db: db}
}

func (r *SettingRepository) FindByKey(ctx context.Context, organizationID uint, key string) (*settings.SystemSetting, error) {
	db := r.db.GetTx(ctx)
	var model dbschema.SystemSetting
	if err := db.WithContext(ctx).
		Where("organization_id = ? AND key = ?", organizationID, key).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, settings.ErrSettingNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *SettingRepository) Upsert(ctx context.Context, setting *settings.SystemSetting) error {
	if setting.Payload == nil {
		setting.Payload = map[string]interface{}{}
	}
	payloadBytes, err := json.Marshal(setting.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	db := r.db.GetTx(ctx)
	model := dbschema.SystemSetting{
		BaseModel: dbschema.BaseModel{
			ID: setting.ID,
		},
		OrganizationID: setting.OrganizationID,
		Key:            setting.Key,
		Payload:        payloadBytes,
		LastUpdatedBy:  setting.LastUpdatedBy,
		UpdatedByEmail: setting.UpdatedByEmail,
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "organization_id"}, {Name: "key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"payload":          model.Payload,
			"last_updated_by":  model.LastUpdatedBy,
			"updated_by_email": model.UpdatedByEmail,
			"updated_at":       gorm.Expr("NOW()"),
		}),
	}).Create(&model).Error
}
