package repo

import (
	"errors"

	"productcore/internal/model"

	"gorm.io/gorm"
)

type AISettingRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewAISettingRepo(db *gorm.DB) *AISettingRepo {
	return &AISettingRepo{db: db, tenantID: 1}
}

func (r *AISettingRepo) WithTenant(tenantID uint64) *AISettingRepo {
	cp := *r
	cp.tenantID = NormalizeTenantID(tenantID)
	return &cp
}

func (r *AISettingRepo) Get() (*model.AISetting, error) {
	var row model.AISetting
	err := r.db.Where("tenant_id = ?", r.tenantID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &model.AISetting{TenantID: r.tenantID}, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *AISettingRepo) Save(row *model.AISetting) error {
	row.TenantID = r.tenantID
	var existing model.AISetting
	err := r.db.Where("tenant_id = ?", r.tenantID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(row).Error
	}
	if err != nil {
		return err
	}
	row.ID = existing.ID
	return r.db.Save(row).Error
}
