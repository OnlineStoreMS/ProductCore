package repo

import (
	"productcore/internal/dto"
	"productcore/internal/model"

	"gorm.io/gorm"
)

type PlatformTypeRepo struct{ db *gorm.DB }

func NewPlatformTypeRepo(db *gorm.DB) *PlatformTypeRepo { return &PlatformTypeRepo{db: db} }

func (r *PlatformTypeRepo) List(keyword string, enabledOnly bool) ([]model.PlatformShopType, error) {
	tx := r.db.Model(&model.PlatformShopType{}).Order("sort DESC, id ASC")
	if keyword != "" {
		kw := "%" + keyword + "%"
		tx = tx.Where("name LIKE ? OR code LIKE ?", kw, kw)
	}
	if enabledOnly {
		tx = tx.Where("enabled = ?", 1)
	}
	var list []model.PlatformShopType
	return list, tx.Find(&list).Error
}

func (r *PlatformTypeRepo) GetByID(id uint64) (*model.PlatformShopType, error) {
	var item model.PlatformShopType
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *PlatformTypeRepo) GetByCode(code string) (*model.PlatformShopType, error) {
	var item model.PlatformShopType
	if err := r.db.Where("code = ?", code).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *PlatformTypeRepo) Create(item *model.PlatformShopType) error { return r.db.Create(item).Error }

func (r *PlatformTypeRepo) Save(item *model.PlatformShopType) error { return r.db.Save(item).Error }

func (r *PlatformTypeRepo) Delete(id uint64) error {
	res := r.db.Delete(&model.PlatformShopType{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *PlatformTypeRepo) CountShops(typeID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&model.PlatformShop{}).Where("platform_type_id = ?", typeID).Count(&n).Error
	return n, err
}

type PlatformShopRepo struct{ db *gorm.DB }

func NewPlatformShopRepo(db *gorm.DB) *PlatformShopRepo { return &PlatformShopRepo{db: db} }

func (r *PlatformShopRepo) List(q dto.PlatformShopQuery) ([]model.PlatformShop, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	tx := r.db.Model(&model.PlatformShop{})
	if q.Keyword != "" {
		kw := "%" + q.Keyword + "%"
		tx = tx.Where("name LIKE ? OR shop_code LIKE ? OR external_shop_id LIKE ?", kw, kw, kw)
	}
	if q.PlatformTypeID > 0 {
		tx = tx.Where("platform_type_id = ?", q.PlatformTypeID)
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PlatformShop
	offset := (q.Page - 1) * q.PageSize
	err := tx.Order("sort DESC, id DESC").Offset(offset).Limit(q.PageSize).Find(&list).Error
	return list, total, err
}

func (r *PlatformShopRepo) GetByID(id uint64) (*model.PlatformShop, error) {
	var item model.PlatformShop
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *PlatformShopRepo) Create(item *model.PlatformShop) error { return r.db.Create(item).Error }

func (r *PlatformShopRepo) Save(item *model.PlatformShop) error { return r.db.Save(item).Error }

func (r *PlatformShopRepo) Delete(id uint64) error {
	res := r.db.Delete(&model.PlatformShop{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
