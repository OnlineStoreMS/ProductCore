package repo

import (
	"productcore/internal/model"

	"gorm.io/gorm"
)

type DistributionConfigRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewDistributionConfigRepo(db *gorm.DB) *DistributionConfigRepo {
	return &DistributionConfigRepo{db: db, tenantID: 1}
}

func (r *DistributionConfigRepo) WithTenant(tenantID uint64) *DistributionConfigRepo {
	return &DistributionConfigRepo{db: r.db, tenantID: normalizeTenantID(tenantID)}
}

func (r *DistributionConfigRepo) List() ([]model.DistributionPublishConfig, error) {
	var list []model.DistributionPublishConfig
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Order("id DESC").
		Find(&list).Error
	return list, err
}

func (r *DistributionConfigRepo) Get(id uint64) (*model.DistributionPublishConfig, error) {
	var row model.DistributionPublishConfig
	err := r.db.Scopes(scopeTenant(r.tenantID)).First(&row, id).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *DistributionConfigRepo) FindByPlatform(platformTypeID uint64) (*model.DistributionPublishConfig, error) {
	var row model.DistributionPublishConfig
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Where("platform_type_id = ?", platformTypeID).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *DistributionConfigRepo) Create(row *model.DistributionPublishConfig) error {
	row.TenantID = r.tenantID
	return r.db.Create(row).Error
}

func (r *DistributionConfigRepo) Save(row *model.DistributionPublishConfig) error {
	row.TenantID = r.tenantID
	return r.db.Save(row).Error
}

func (r *DistributionConfigRepo) Delete(id uint64) error {
	return r.db.Scopes(scopeTenant(r.tenantID)).Delete(&model.DistributionPublishConfig{}, id).Error
}
