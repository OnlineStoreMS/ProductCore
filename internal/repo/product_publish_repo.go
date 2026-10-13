package repo

import (
	"errors"

	"productcore/internal/model"

	"gorm.io/gorm"
)

type ProductPublishRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewProductPublishRepo(db *gorm.DB) *ProductPublishRepo {
	return &ProductPublishRepo{db: db, tenantID: 1}
}

func (r *ProductPublishRepo) WithTenant(tenantID uint64) *ProductPublishRepo {
	cp := *r
	cp.tenantID = normalizeTenantID(tenantID)
	return &cp
}

func (r *ProductPublishRepo) scoped() *gorm.DB {
	return r.db.Scopes(scopeTenant(r.tenantID))
}

func (r *ProductPublishRepo) Get(productID uint64, platform, platformShopID string) (*model.ProductPublish, error) {
	var row model.ProductPublish
	err := r.db.Unscoped().
		Where("tenant_id = ? AND product_id = ? AND platform = ? AND platform_shop_id = ?",
			r.tenantID, productID, platform, platformShopID).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ProductPublishRepo) Save(row *model.ProductPublish) error {
	row.TenantID = r.tenantID
	row.DeletedAt = gorm.DeletedAt{}
	if row.ID == 0 {
		return r.db.Create(row).Error
	}
	return r.db.Unscoped().Save(row).Error
}

func (r *ProductPublishRepo) ListByProductIDs(productIDs []uint64) ([]model.ProductPublish, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}
	var rows []model.ProductPublish
	err := r.scoped().
		Where("product_id IN ?", productIDs).
		Order("id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *ProductPublishRepo) ListOpen() ([]model.ProductPublish, error) {
	var rows []model.ProductPublish
	err := r.scoped().
		Where("status = ? AND agent_job_id > 0", "running").
		Find(&rows).Error
	return rows, err
}

func IsPublishNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
