package repo

import (
	"fmt"
	"strings"

	"productcore/internal/dto"
	"productcore/internal/model"

	"gorm.io/gorm"
)

const collectedProductExists = `EXISTS (
	SELECT 1 FROM products p
	WHERE p.tenant_id = distribution_shop_items.tenant_id
		AND p.material_code = distribution_shop_items.item_id
		AND p.material_code <> ''
		AND p.deleted_at IS NULL
)`

type DistributionRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewDistributionRepo(db *gorm.DB) *DistributionRepo {
	return &DistributionRepo{db: db, tenantID: 1}
}

func (r *DistributionRepo) WithTenant(tenantID uint64) *DistributionRepo {
	return &DistributionRepo{db: r.db, tenantID: normalizeTenantID(tenantID)}
}

func (r *DistributionRepo) ListShops(q dto.DistributionShopQuery) ([]model.DistributionShop, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	tx := r.db.Scopes(scopeTenant(r.tenantID)).Model(&model.DistributionShop{})
	if q.PlatformTypeID > 0 {
		tx = tx.Where("platform_type_id = ?", q.PlatformTypeID)
	}
	if q.Keyword != "" {
		tx = tx.Where("name LIKE ?", "%"+q.Keyword+"%")
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.DistributionShop
	err := tx.Order("id DESC").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&list).Error
	return list, total, err
}

func (r *DistributionRepo) GetShop(id uint64) (*model.DistributionShop, error) {
	var item model.DistributionShop
	err := r.db.Scopes(scopeTenant(r.tenantID)).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *DistributionRepo) CreateShop(item *model.DistributionShop) error {
	item.TenantID = r.tenantID
	return r.db.Create(item).Error
}

func (r *DistributionRepo) SaveShop(item *model.DistributionShop) error {
	return r.db.Save(item).Error
}

func (r *DistributionRepo) DeleteShop(id uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Scopes(scopeTenant(r.tenantID)).Where("shop_id = ?", id).Delete(&model.DistributionShopItem{}).Error; err != nil {
			return err
		}
		res := tx.Scopes(scopeTenant(r.tenantID)).Delete(&model.DistributionShop{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *DistributionRepo) CountItems(shopIDs []uint64) (map[uint64]int64, error) {
	out := map[uint64]int64{}
	if len(shopIDs) == 0 {
		return out, nil
	}
	type row struct {
		ShopID uint64
		Count  int64
	}
	var rows []row
	err := r.db.Model(&model.DistributionShopItem{}).
		Select("shop_id, COUNT(*) AS count").
		Where("tenant_id = ? AND shop_id IN ?", r.tenantID, shopIDs).
		Group("shop_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, item := range rows {
		out[item.ShopID] = item.Count
	}
	return out, nil
}

func (r *DistributionRepo) ListItems(shopID uint64, q dto.DistributionItemQuery) ([]model.DistributionShopItem, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	tx := r.db.Scopes(scopeTenant(r.tenantID)).Model(&model.DistributionShopItem{}).Where("shop_id = ?", shopID)
	if q.Keyword != "" {
		kw := "%" + q.Keyword + "%"
		tx = tx.Where("title LIKE ? OR item_id LIKE ?", kw, kw)
	}
	switch strings.TrimSpace(q.Collected) {
	case "1":
		tx = tx.Where(collectedProductExists)
	case "0":
		tx = tx.Where("NOT " + collectedProductExists)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.DistributionShopItem
	err := tx.Order(itemOrderSQL(q)).Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&list).Error
	return list, total, err
}

func itemOrderSQL(q dto.DistributionItemQuery) string {
	columns := map[string]string{
		"sales":        "sales",
		"price":        "price",
		"monthDeals":   "month_deals",
		"monthConsign": "month_consign",
		"shipTime":     "ship_time",
		"listedAt":     "listed_at",
	}
	col := columns[q.SortBy]
	if col == "" {
		col = "sales"
	}
	dir := "DESC"
	if strings.EqualFold(strings.TrimSpace(q.SortOrder), "asc") {
		dir = "ASC"
	}
	switch col {
	case "sales", "price", "month_deals", "month_consign":
		num := `([0-9]+(\.[0-9]+){0,1})`
		return fmt.Sprintf(
			`(CASE WHEN %[1]s ~ '万' THEN NULLIF(substring(%[1]s from '%[2]s'), '')::numeric * 10000 ELSE NULLIF(substring(%[1]s from '%[2]s'), '')::numeric END) %[3]s NULLS LAST, id DESC`,
			col, num, dir,
		)
	default:
		return fmt.Sprintf(`NULLIF(btrim(%s), '') %s NULLS LAST, id DESC`, col, dir)
	}
}

func (r *DistributionRepo) FindItem(shopID uint64, itemID string) (*model.DistributionShopItem, error) {
	var item model.DistributionShopItem
	err := r.db.Unscoped().Scopes(scopeTenant(r.tenantID)).
		Where("shop_id = ? AND item_id = ?", shopID, itemID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *DistributionRepo) DeleteItemsExcept(shopID uint64, keep []string) (int64, error) {
	tx := r.db.Unscoped().Scopes(scopeTenant(r.tenantID)).Where("shop_id = ?", shopID)
	if len(keep) > 0 {
		tx = tx.Where("item_id NOT IN ?", keep)
	}
	res := tx.Delete(&model.DistributionShopItem{})
	return res.RowsAffected, res.Error
}

func (r *DistributionRepo) CreateItem(item *model.DistributionShopItem) error {
	item.TenantID = r.tenantID
	return r.db.Create(item).Error
}

func (r *DistributionRepo) SaveItem(item *model.DistributionShopItem) error {
	item.DeletedAt = gorm.DeletedAt{}
	return r.db.Unscoped().Save(item).Error
}
