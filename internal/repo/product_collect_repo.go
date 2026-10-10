package repo

import (
	"productcore/internal/model"

	"gorm.io/gorm"
)

type ProductCollectRepo struct{ db *gorm.DB }

func NewProductCollectRepo(db *gorm.DB) *ProductCollectRepo {
	return &ProductCollectRepo{db: db}
}

func (r *ProductCollectRepo) Create(task *model.ProductCollectTask) error {
	return r.db.Create(task).Error
}

func (r *ProductCollectRepo) Save(task *model.ProductCollectTask) error {
	return r.db.Save(task).Error
}

func (r *ProductCollectRepo) Clear(tenantID uint64) error {
	return r.db.Unscoped().Scopes(scopeTenant(tenantID)).Delete(&model.ProductCollectTask{}).Error
}

func (r *ProductCollectRepo) List(tenantID uint64, page, pageSize int) ([]model.ProductCollectTask, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	query := func() *gorm.DB {
		return r.db.Scopes(scopeTenant(tenantID)).Model(&model.ProductCollectTask{})
	}
	var total int64
	if err := query().Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ProductCollectTask
	err := query().Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
