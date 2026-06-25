package repo

import (
	"productcore/internal/model"
	"time"

	"gorm.io/gorm"
)

func (r *ProductRepo) GetEditDraft(productID uint64) (*model.ProductEditDraft, error) {
	var d model.ProductEditDraft
	if err := r.db.Where("product_id = ?", productID).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *ProductRepo) UpsertEditDraft(productID uint64, payloadJSON string) error {
	var existing model.ProductEditDraft
	err := r.db.Where("product_id = ?", productID).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(&model.ProductEditDraft{
				ProductID:   productID,
				PayloadJSON: payloadJSON,
			}).Error
		}
		return err
	}
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"payload_json": payloadJSON,
		"updated_at":   time.Now(),
	}).Error
}

func (r *ProductRepo) DeleteEditDraft(productID uint64) error {
	return r.db.Where("product_id = ?", productID).Delete(&model.ProductEditDraft{}).Error
}

func (r *ProductRepo) EditDraftFlags(productIDs []uint64) (map[uint64]bool, error) {
	out := make(map[uint64]bool, len(productIDs))
	if len(productIDs) == 0 {
		return out, nil
	}
	var ids []uint64
	if err := r.db.Model(&model.ProductEditDraft{}).
		Where("product_id IN ?", productIDs).
		Pluck("product_id", &ids).Error; err != nil {
		return nil, err
	}
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}
