package repo

import (
	"productcore/internal/model"

	"gorm.io/gorm"
)

type BrandRepo struct{ db *gorm.DB }

func NewBrandRepo(db *gorm.DB) *BrandRepo { return &BrandRepo{db: db} }

func (r *BrandRepo) List(tenantID uint64, keyword string) ([]model.Brand, error) {
	tx := r.db.Scopes(scopeTenant(tenantID)).Model(&model.Brand{}).Order("sort DESC, id ASC")
	if keyword != "" {
		tx = tx.Where("name LIKE ?", "%"+keyword+"%")
	}
	var list []model.Brand
	return list, tx.Find(&list).Error
}

func (r *BrandRepo) GetByID(tenantID, id uint64) (*model.Brand, error) {
	var b model.Brand
	if err := r.db.Scopes(scopeTenant(tenantID)).First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BrandRepo) Create(b *model.Brand) error { return r.db.Create(b).Error }

func (r *BrandRepo) Save(b *model.Brand) error { return r.db.Save(b).Error }

func (r *BrandRepo) Delete(tenantID, id uint64) error {
	res := r.db.Scopes(scopeTenant(tenantID)).Delete(&model.Brand{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *BrandRepo) CountProducts(tenantID, brandID uint64) (int64, error) {
	var n int64
	err := r.db.Scopes(scopeTenant(tenantID)).Model(&model.Product{}).Where("brand_id = ?", brandID).Count(&n).Error
	return n, err
}

type CategoryRepo struct{ db *gorm.DB }

func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db: db} }

func (r *CategoryRepo) ListAll(tenantID uint64) ([]model.Category, error) {
	var list []model.Category
	err := r.db.Scopes(scopeTenant(tenantID)).Order("sort DESC, id ASC").Find(&list).Error
	return list, err
}

func (r *CategoryRepo) GetByID(tenantID, id uint64) (*model.Category, error) {
	var c model.Category
	if err := r.db.Scopes(scopeTenant(tenantID)).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepo) Create(c *model.Category) error { return r.db.Create(c).Error }

func (r *CategoryRepo) Save(c *model.Category) error { return r.db.Save(c).Error }

func (r *CategoryRepo) Delete(tenantID, id uint64) error {
	res := r.db.Scopes(scopeTenant(tenantID)).Delete(&model.Category{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CategoryRepo) CountChildren(tenantID, parentID uint64) (int64, error) {
	var n int64
	err := r.db.Scopes(scopeTenant(tenantID)).Model(&model.Category{}).Where("parent_id = ?", parentID).Count(&n).Error
	return n, err
}

func (r *CategoryRepo) CountProducts(tenantID, categoryID uint64) (int64, error) {
	var n int64
	err := r.db.Scopes(scopeTenant(tenantID)).Model(&model.Product{}).Where("category_id = ?", categoryID).Count(&n).Error
	return n, err
}

type GroupRepo struct{ db *gorm.DB }

func NewGroupRepo(db *gorm.DB) *GroupRepo { return &GroupRepo{db: db} }

func (r *GroupRepo) List(tenantID uint64) ([]model.ProductGroup, error) {
	var list []model.ProductGroup
	err := r.db.Scopes(scopeTenant(tenantID)).Order("sort DESC, id ASC").Find(&list).Error
	return list, err
}

func (r *GroupRepo) GetByID(tenantID, id uint64) (*model.ProductGroup, error) {
	var g model.ProductGroup
	if err := r.db.Scopes(scopeTenant(tenantID)).First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GroupRepo) Create(g *model.ProductGroup) error { return r.db.Create(g).Error }

func (r *GroupRepo) Save(g *model.ProductGroup) error { return r.db.Save(g).Error }

func (r *GroupRepo) Delete(tenantID, id uint64) error {
	if err := r.db.Where("group_id = ?", id).Delete(&model.ProductGroupRelation{}).Error; err != nil {
		return err
	}
	res := r.db.Scopes(scopeTenant(tenantID)).Delete(&model.ProductGroup{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GroupRepo) CountProducts(tenantID, groupID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&model.ProductGroupRelation{}).Where("group_id = ?", groupID).Count(&n).Error
	return n, err
}

func (r *GroupRepo) ListProductIDs(tenantID, groupID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.Model(&model.ProductGroupRelation{}).Where("group_id = ?", groupID).Pluck("product_id", &ids).Error
	return ids, err
}

func (r *BrandRepo) GetName(id uint64) string {
	var b model.Brand
	if r.db.First(&b, id).Error != nil {
		return ""
	}
	return b.Name
}

func (r *CategoryRepo) GetName(id uint64) string {
	var c model.Category
	if r.db.First(&c, id).Error != nil {
		return ""
	}
	return c.Name
}
