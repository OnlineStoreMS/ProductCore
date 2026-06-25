package repo

import (
	"productcore/internal/dto"
	"productcore/internal/model"
	"time"

	"gorm.io/gorm"
)

// DraftListEntry 草稿箱条目（新建草稿 + 已发布商品的编辑草稿）
type DraftListEntry struct {
	Product      model.Product
	DraftSavedAt time.Time
	IsEditOnly   bool // true=已发布商品的未合并编辑草稿
}

func (r *ProductRepo) ListDraftEntries(q dto.ProductQuery) ([]DraftListEntry, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}

	tx := r.db.Model(&model.Product{}).
		Joins("LEFT JOIN product_edit_drafts ed ON ed.product_id = products.id AND products.is_draft = 0").
		Where("products.is_draft = 1 OR ed.id IS NOT NULL")
	tx = applyKeywordFilter(tx, q.Keyword)

	countTx := tx.Session(&gorm.Session{}).Select("products.id")
	var total int64
	if err := countTx.Distinct("products.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type scanRow struct {
		model.Product
		DraftSavedAt time.Time `gorm:"column:draft_saved_at"`
		IsEditOnly   bool      `gorm:"column:is_edit_only"`
	}
	var rows []scanRow
	offset := (q.Page - 1) * q.PageSize
	err := tx.Select(`
		products.*,
		CASE WHEN products.is_draft = 1 THEN products.updated_at ELSE ed.updated_at END AS draft_saved_at,
		CASE WHEN products.is_draft = 1 THEN FALSE ELSE TRUE END AS is_edit_only
	`).
		Order("draft_saved_at DESC, products.id DESC").
		Offset(offset).Limit(q.PageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]DraftListEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, DraftListEntry{
			Product:      row.Product,
			DraftSavedAt: row.DraftSavedAt,
			IsEditOnly:   row.IsEditOnly,
		})
	}
	return out, total, nil
}
