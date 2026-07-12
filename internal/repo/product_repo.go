package repo

import (
	"productcore/internal/dto"
	"productcore/internal/model"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type ProductRepo struct {
	db *gorm.DB
	tenantID uint64
}

func NewProductRepo(db *gorm.DB) *ProductRepo {
	return &ProductRepo{db: db, tenantID: 1}
}

func (r *ProductRepo) WithTenant(tenantID uint64) *ProductRepo {
	return &ProductRepo{db: r.db, tenantID: normalizeTenantID(tenantID)}
}

func (r *ProductRepo) TenantID() uint64 { return normalizeTenantID(r.tenantID) }

func (r *ProductRepo) DB() *gorm.DB { return r.db }

func applyKeywordFilter(tx *gorm.DB, keyword string) *gorm.DB {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return tx
	}
	kw := "%" + keyword + "%"
	if id, err := strconv.ParseUint(keyword, 10, 64); err == nil && id > 0 {
		return tx.Where(
			"name LIKE ? OR product_sn LIKE ? OR material_code LIKE ? OR source LIKE ? OR id = ?",
			kw, kw, kw, kw, id,
		)
	}
	return tx.Where(
		"name LIKE ? OR product_sn LIKE ? OR material_code LIKE ? OR source LIKE ?",
		kw, kw, kw, kw,
	)
}

func (r *ProductRepo) List(q dto.ProductQuery) ([]model.Product, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	tx := r.db.Scopes(scopeTenant(r.tenantID)).Model(&model.Product{}).Where("is_draft = ?", 0)
	tx = applyKeywordFilter(tx, q.Keyword)
	if q.BrandID > 0 {
		tx = tx.Where("brand_id = ?", q.BrandID)
	}
	if len(q.CategoryIDs) > 0 {
		tx = tx.Where("category_id IN ?", q.CategoryIDs)
	} else if q.CategoryID > 0 {
		tx = tx.Where("category_id = ?", q.CategoryID)
	}
	if q.PublishStatus != nil {
		tx = tx.Where("publish_status = ?", *q.PublishStatus)
	}
	if q.GroupID > 0 {
		tx = tx.Joins("JOIN product_group_relations pgr ON pgr.product_id = products.id").
			Where("pgr.group_id = ?", q.GroupID)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Product
	offset := (q.Page - 1) * q.PageSize
	err := tx.Order("sort DESC, id DESC").Offset(offset).Limit(q.PageSize).Find(&list).Error
	return list, total, err
}

func (r *ProductRepo) applyListFilters(tx *gorm.DB, q dto.ProductQuery) *gorm.DB {
	tx = applyKeywordFilter(tx, q.Keyword)
	if q.BrandID > 0 {
		tx = tx.Where("brand_id = ?", q.BrandID)
	}
	if len(q.CategoryIDs) > 0 {
		tx = tx.Where("category_id IN ?", q.CategoryIDs)
	} else if q.CategoryID > 0 {
		tx = tx.Where("category_id = ?", q.CategoryID)
	}
	if q.PublishStatus != nil {
		tx = tx.Where("publish_status = ?", *q.PublishStatus)
	}
	if q.GroupID > 0 {
		tx = tx.Joins("JOIN product_group_relations pgr ON pgr.product_id = products.id").
			Where("pgr.group_id = ?", q.GroupID)
	}
	return tx
}

func (r *ProductRepo) ListTrashed(q dto.ProductQuery) ([]model.Product, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	tx := r.db.Unscoped().Scopes(scopeTenant(r.tenantID)).Model(&model.Product{}).Where("deleted_at IS NOT NULL")
	tx = r.applyListFilters(tx, q)
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Product
	offset := (q.Page - 1) * q.PageSize
	err := tx.Order("deleted_at DESC, id DESC").Offset(offset).Limit(q.PageSize).Find(&list).Error
	return list, total, err
}

func (r *ProductRepo) GetByIDUnscoped(id uint64) (*model.Product, error) {
	var p model.Product
	if err := r.db.Unscoped().Scopes(scopeTenant(r.tenantID)).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) ForceDeleteProduct(id uint64) error {
	res := r.db.Unscoped().Scopes(scopeTenant(r.tenantID)).Delete(&model.Product{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ProductRepo) GetByID(id uint64) (*model.Product, error) {
	var p model.Product
	if err := r.db.Scopes(scopeTenant(r.tenantID)).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) GetByMaterialCode(code string) (*model.Product, error) {
	var p model.Product
	if err := r.db.Scopes(scopeTenant(r.tenantID)).Where("material_code = ?", code).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// GetByMaterialCodeUnscoped 含已软删记录，供导入按资料编码覆盖
func (r *ProductRepo) GetByMaterialCodeUnscoped(code string) (*model.Product, error) {
	var p model.Product
	if err := r.db.Unscoped().Scopes(scopeTenant(r.tenantID)).Where("material_code = ?", code).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) RestoreProduct(id uint64) error {
	return r.db.Unscoped().Scopes(scopeTenant(r.tenantID)).Model(&model.Product{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *ProductRepo) CountByMaterialCode(code string, excludeID uint64) (int64, error) {
	tx := r.db.Scopes(scopeTenant(r.tenantID)).Model(&model.Product{}).Where("material_code = ?", code)
	if excludeID > 0 {
		tx = tx.Where("id <> ?", excludeID)
	}
	var count int64
	err := tx.Count(&count).Error
	return count, err
}

func (r *ProductRepo) CountBySN(sn string, excludeID uint64) (int64, error) {
	tx := r.db.Scopes(scopeTenant(r.tenantID)).Model(&model.Product{}).Where("product_sn = ?", sn)
	if excludeID > 0 {
		tx = tx.Where("id <> ?", excludeID)
	}
	var count int64
	err := tx.Count(&count).Error
	return count, err
}

func (r *ProductRepo) Create(p *model.Product) error {
	p.TenantID = r.TenantID()
	return r.db.Create(p).Error
}

// CreateDraft 创建草稿商品，brand_id/category_id 留空（NULL）
func (r *ProductRepo) CreateDraft(p *model.Product) error {
	p.TenantID = r.TenantID()
	return r.db.Omit("BrandID", "CategoryID").Create(p).Error
}

func (r *ProductRepo) UpdateFields(id uint64, fields map[string]interface{}) error {
	return r.db.Scopes(scopeTenant(r.tenantID)).Model(&model.Product{}).Where("id = ?", id).Updates(fields).Error
}

func (r *ProductRepo) Delete(id uint64) error {
	res := r.db.Scopes(scopeTenant(r.tenantID)).Delete(&model.Product{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ProductRepo) UpdatePublishStatus(id uint64, status int8) error {
	res := r.db.Scopes(scopeTenant(r.tenantID)).Model(&model.Product{}).Where("id = ?", id).Update("publish_status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ProductRepo) ListSkus(productID uint64) ([]model.Sku, error) {
	var skus []model.Sku
	err := r.db.Scopes(scopeTenant(r.tenantID)).Where("product_id = ?", productID).
		Order("sort_order ASC, id ASC").
		Find(&skus).Error
	return skus, err
}

func (r *ProductRepo) CountSkus(productID uint64) (int64, error) {
	var n int64
	err := r.db.Scopes(scopeTenant(r.tenantID)).Model(&model.Sku{}).Where("product_id = ?", productID).Count(&n).Error
	return n, err
}

func (r *ProductRepo) DeleteSkusByProduct(productID uint64) error {
	// 物理删除：GORM 默认软删会保留 sku_code，导致同商品再次保存时触发唯一约束冲突
	return r.db.Unscoped().Scopes(scopeTenant(r.tenantID)).Where("product_id = ?", productID).Delete(&model.Sku{}).Error
}

func (r *ProductRepo) CreateSku(sku *model.Sku) error {
	sku.TenantID = r.TenantID()
	return r.db.Create(sku).Error
}

func (r *ProductRepo) CountSkuByCode(code string, excludeProductID uint64) (int64, error) {
	tx := r.db.Scopes(scopeTenant(r.tenantID)).Model(&model.Sku{}).Where("sku_code = ?", code)
	if excludeProductID > 0 {
		tx = tx.Where("product_id <> ?", excludeProductID)
	}
	var n int64
	err := tx.Count(&n).Error
	return n, err
}

func (r *ProductRepo) ListGroupIDs(productID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.Model(&model.ProductGroupRelation{}).Where("product_id = ?", productID).Pluck("group_id", &ids).Error
	return ids, err
}

func (r *ProductRepo) DeleteGroupRelations(productID uint64) error {
	return r.db.Where("product_id = ?", productID).Delete(&model.ProductGroupRelation{}).Error
}

func (r *ProductRepo) CreateGroupRelation(productID, groupID uint64) error {
	return r.db.Create(&model.ProductGroupRelation{ProductID: productID, GroupID: groupID}).Error
}

func (r *ProductRepo) Transaction(fn func(tx *ProductRepo) error) error {
	return r.db.Transaction(func(txDB *gorm.DB) error {
		return fn(&ProductRepo{db: txDB, tenantID: r.tenantID})
	})
}
