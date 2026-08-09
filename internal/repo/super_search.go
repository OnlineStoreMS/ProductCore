package repo

import (
	"strconv"
	"strings"
)

type SkuSearchRow struct {
	SkuID          uint64
	SkuCode        string
	SpecData       string
	Price          float64
	Stock          int
	Pic            string
	SortOrder      int
	ProductID      uint64
	ProductName    string
	MaterialCode   string
	ProductSn      string
	ProductPic     string
	BrandID        uint64
	CategoryID     uint64
	PublishStatus  int8
}

func (r *ProductRepo) searchSkus(keyword string, page, pageSize int) ([]SkuSearchRow, int64, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, 0, nil
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}

	kw := "%" + keyword + "%"
	base := r.db.Table("product_skus AS s").
		Joins("JOIN products p ON p.id = s.product_id AND p.deleted_at IS NULL AND p.is_draft = 0").
		Where("s.deleted_at IS NULL AND s.tenant_id = ? AND p.tenant_id = ?", normalizeTenantID(r.tenantID), normalizeTenantID(r.tenantID))

	// 仅匹配当前 SKU 字段 + 商品基础信息；不匹配 p.sku_specs_json，
	// 避免兄弟规格文案（如「不含…飞轮…」）把同商品其它 SKU 一并搜出。
	whereSQL := `(s.sku_code LIKE ? OR s.spec_data LIKE ? OR p.name LIKE ? OR p.product_sn LIKE ? OR p.material_code LIKE ? OR p.source LIKE ?`
	args := []interface{}{kw, kw, kw, kw, kw, kw}
	if id, err := strconv.ParseUint(keyword, 10, 64); err == nil && id > 0 {
		whereSQL += ` OR p.id = ? OR s.id = ?`
		args = append(args, id, id)
	}
	whereSQL += `)`
	filtered := base.Where(whereSQL, args...)

	var total int64
	if err := filtered.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []SkuSearchRow
	offset := (page - 1) * pageSize
	err := filtered.
		Select(`s.id AS sku_id, s.sku_code, s.spec_data, s.price, s.stock, s.pic, s.sort_order,
			p.id AS product_id, p.name AS product_name, p.material_code, p.product_sn,
			p.pic AS product_pic, p.brand_id, p.category_id, p.publish_status`).
		Order("p.id DESC, s.sort_order ASC, s.id ASC").
		Offset(offset).
		Limit(pageSize).
		Scan(&rows).Error
	return rows, total, err
}

func (r *ProductRepo) SearchSkus(keyword string, page, pageSize int) ([]SkuSearchRow, int64, error) {
	return r.WithTenant(r.tenantID).searchSkus(keyword, page, pageSize)
}
