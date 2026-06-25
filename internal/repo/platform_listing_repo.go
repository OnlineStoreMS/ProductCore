package repo

import (
	"productcore/internal/dto"
	"productcore/internal/model"

	"gorm.io/gorm"
)

type PlatformListingRepo struct{ db *gorm.DB }

func NewPlatformListingRepo(db *gorm.DB) *PlatformListingRepo {
	return &PlatformListingRepo{db: db}
}

type listedShopRow struct {
	ProductID        uint64
	ShopID           uint64
	ShopName         string
	PlatformTypeName string
	PlatformTypeLogo string
	SourceChannel    string
}

func (r *PlatformListingRepo) ShopsByProductIDs(productIDs []uint64) (map[uint64][]dto.ListedShopDTO, error) {
	out := make(map[uint64][]dto.ListedShopDTO)
	if len(productIDs) == 0 {
		return out, nil
	}
	var rows []listedShopRow
	err := r.db.Table("platform_listings pl").
		Select(`pl.product_id, ps.id AS shop_id, ps.name AS shop_name,
			pst.name AS platform_type_name, pst.logo AS platform_type_logo, ps.source_channel`).
		Joins("JOIN platform_shops ps ON ps.id = pl.platform_shop_id AND ps.deleted_at IS NULL").
		Joins("JOIN platform_shop_types pst ON pst.id = ps.platform_type_id AND pst.deleted_at IS NULL").
		Where("pl.product_id IN ? AND pl.deleted_at IS NULL", productIDs).
		Order("ps.sort DESC, ps.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.ProductID] = append(out[row.ProductID], dto.ListedShopDTO{
			ShopID: row.ShopID, ShopName: row.ShopName,
			PlatformTypeName: row.PlatformTypeName, PlatformTypeLogo: row.PlatformTypeLogo,
			SourceChannel: row.SourceChannel,
		})
	}
	return out, nil
}

func (r *PlatformListingRepo) ReplaceProductListings(productID uint64, shopIDs []uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("product_id = ?", productID).Delete(&model.PlatformListing{}).Error; err != nil {
			return err
		}
		if len(shopIDs) == 0 {
			return nil
		}
		seen := make(map[uint64]struct{}, len(shopIDs))
		items := make([]model.PlatformListing, 0, len(shopIDs))
		for _, sid := range shopIDs {
			if sid == 0 {
				continue
			}
			if _, dup := seen[sid]; dup {
				continue
			}
			seen[sid] = struct{}{}
			items = append(items, model.PlatformListing{
				ProductID: productID, PlatformShopID: sid, ListingStatus: 1,
			})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *PlatformListingRepo) DeleteByProduct(productID uint64) error {
	return r.db.Where("product_id = ?", productID).Delete(&model.PlatformListing{}).Error
}

func (r *PlatformListingRepo) DeleteByShop(shopID uint64) error {
	return r.db.Where("platform_shop_id = ?", shopID).Delete(&model.PlatformListing{}).Error
}

func (r *PlatformListingRepo) CountProductsByShopIDs(shopIDs []uint64) (map[uint64]int64, error) {
	out := make(map[uint64]int64)
	if len(shopIDs) == 0 {
		return out, nil
	}
	type row struct {
		PlatformShopID uint64
		Count          int64
	}
	var rows []row
	err := r.db.Model(&model.PlatformListing{}).
		Select("platform_shop_id, COUNT(*) AS count").
		Where("platform_shop_id IN ?", shopIDs).
		Group("platform_shop_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, item := range rows {
		out[item.PlatformShopID] = item.Count
	}
	return out, nil
}

func (r *PlatformListingRepo) ListProductIDsByShop(shopID uint64, keyword string, page, pageSize int) ([]uint64, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	tx := r.db.Table("platform_listings pl").
		Joins("JOIN products p ON p.id = pl.product_id AND p.deleted_at IS NULL AND p.is_draft = 0").
		Where("pl.platform_shop_id = ? AND pl.deleted_at IS NULL", shopID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		tx = tx.Where("p.name LIKE ? OR p.material_code LIKE ? OR p.product_sn LIKE ?", kw, kw, kw)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var ids []uint64
	offset := (page - 1) * pageSize
	err := tx.Select("p.id").
		Order("pl.id DESC").
		Offset(offset).Limit(pageSize).
		Pluck("p.id", &ids).Error
	return ids, total, err
}
