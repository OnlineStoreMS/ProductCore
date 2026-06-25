package database

import (
	"productcore/internal/model"

	"gorm.io/gorm"
)

// backfillSkuSortOrder 为历史 SKU 按 id 顺序补 sort_order，保证列表/导出顺序稳定。
func backfillSkuSortOrder(db *gorm.DB) error {
	var productIDs []uint64
	if err := db.Model(&model.Sku{}).Distinct("product_id").Pluck("product_id", &productIDs).Error; err != nil {
		return err
	}
	for _, productID := range productIDs {
		var skus []model.Sku
		if err := db.Where("product_id = ?", productID).Order("id ASC").Find(&skus).Error; err != nil {
			return err
		}
		needsBackfill := false
		for i, sku := range skus {
			if sku.SortOrder != i {
				needsBackfill = true
				break
			}
		}
		if !needsBackfill {
			continue
		}
		for i, sku := range skus {
			if err := db.Model(&model.Sku{}).Where("id = ?", sku.ID).Update("sort_order", i).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
