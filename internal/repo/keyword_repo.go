package repo

import (
	"productcore/internal/model"

	"gorm.io/gorm"
)

type KeywordRepo struct{ db *gorm.DB }

func NewKeywordRepo(db *gorm.DB) *KeywordRepo { return &KeywordRepo{db: db} }

func (r *KeywordRepo) List(tenantID uint64) ([]model.ProductKeyword, error) {
	var list []model.ProductKeyword
	err := r.db.Scopes(scopeTenant(tenantID)).Order("sort DESC, id ASC").Find(&list).Error
	return list, err
}

func (r *KeywordRepo) GetByID(tenantID, id uint64) (*model.ProductKeyword, error) {
	var k model.ProductKeyword
	if err := r.db.Scopes(scopeTenant(tenantID)).First(&k, id).Error; err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *KeywordRepo) GetByName(tenantID uint64, name string) (*model.ProductKeyword, error) {
	var k model.ProductKeyword
	if err := r.db.Scopes(scopeTenant(tenantID)).Where("name = ?", name).First(&k).Error; err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *KeywordRepo) Create(k *model.ProductKeyword) error { return r.db.Create(k).Error }

func (r *KeywordRepo) Save(k *model.ProductKeyword) error { return r.db.Save(k).Error }

func (r *KeywordRepo) Delete(tenantID, id uint64) error {
	if err := r.db.Where("keyword_id = ?", id).Delete(&model.ProductKeywordRelation{}).Error; err != nil {
		return err
	}
	res := r.db.Scopes(scopeTenant(tenantID)).Delete(&model.ProductKeyword{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *KeywordRepo) CountProducts(tenantID, keywordID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&model.ProductKeywordRelation{}).Where("keyword_id = ?", keywordID).Count(&n).Error
	return n, err
}

func (r *KeywordRepo) ListProductIDs(tenantID, keywordID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.Model(&model.ProductKeywordRelation{}).Where("keyword_id = ?", keywordID).Pluck("product_id", &ids).Error
	return ids, err
}

func (r *KeywordRepo) ReplaceProducts(keywordID uint64, productIDs []uint64) error {
	if err := r.db.Where("keyword_id = ?", keywordID).Delete(&model.ProductKeywordRelation{}).Error; err != nil {
		return err
	}
	for _, pid := range productIDs {
		if pid == 0 {
			continue
		}
		if err := r.db.Create(&model.ProductKeywordRelation{ProductID: pid, KeywordID: keywordID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *KeywordRepo) AddProducts(keywordID uint64, productIDs []uint64) error {
	for _, pid := range productIDs {
		if pid == 0 {
			continue
		}
		var n int64
		if err := r.db.Model(&model.ProductKeywordRelation{}).
			Where("product_id = ? AND keyword_id = ?", pid, keywordID).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		if err := r.db.Create(&model.ProductKeywordRelation{ProductID: pid, KeywordID: keywordID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *KeywordRepo) RemoveProduct(keywordID, productID uint64) error {
	return r.db.Where("keyword_id = ? AND product_id = ?", keywordID, productID).
		Delete(&model.ProductKeywordRelation{}).Error
}
