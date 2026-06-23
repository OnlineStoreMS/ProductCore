package repo

import (
	"productcore/internal/dto"
	"productcore/internal/model"

	"gorm.io/gorm"
)

type ProductRepo struct {
	db *gorm.DB
}

func NewProductRepo(db *gorm.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) DB() *gorm.DB { return r.db }

func (r *ProductRepo) List(q dto.ProductQuery) ([]model.Product, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	tx := r.db.Model(&model.Product{})
	if q.Keyword != "" {
		kw := "%" + q.Keyword + "%"
		tx = tx.Where("name LIKE ? OR product_sn LIKE ?", kw, kw)
	}
	if q.BrandID > 0 {
		tx = tx.Where("brand_id = ?", q.BrandID)
	}
	if q.CategoryID > 0 {
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

func (r *ProductRepo) GetByID(id uint64) (*model.Product, error) {
	var p model.Product
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) CountBySN(sn string, excludeID uint64) (int64, error) {
	tx := r.db.Model(&model.Product{}).Where("product_sn = ?", sn)
	if excludeID > 0 {
		tx = tx.Where("id <> ?", excludeID)
	}
	var count int64
	err := tx.Count(&count).Error
	return count, err
}

func (r *ProductRepo) Create(p *model.Product) error {
	return r.db.Create(p).Error
}

func (r *ProductRepo) UpdateFields(id uint64, fields map[string]interface{}) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).Updates(fields).Error
}

func (r *ProductRepo) Delete(id uint64) error {
	res := r.db.Delete(&model.Product{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ProductRepo) UpdatePublishStatus(id uint64, status int8) error {
	res := r.db.Model(&model.Product{}).Where("id = ?", id).Update("publish_status", status)
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
	err := r.db.Where("product_id = ?", productID).Find(&skus).Error
	return skus, err
}

func (r *ProductRepo) CountSkus(productID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&model.Sku{}).Where("product_id = ?", productID).Count(&n).Error
	return n, err
}

func (r *ProductRepo) DeleteSkusByProduct(productID uint64) error {
	return r.db.Where("product_id = ?", productID).Delete(&model.Sku{}).Error
}

func (r *ProductRepo) CreateSku(sku *model.Sku) error {
	return r.db.Create(sku).Error
}

func (r *ProductRepo) CountSkuByCode(code string) (int64, error) {
	var n int64
	err := r.db.Model(&model.Sku{}).Where("sku_code = ?", code).Count(&n).Error
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
		return fn(&ProductRepo{db: txDB})
	})
}
