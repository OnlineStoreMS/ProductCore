package repo

import "gorm.io/gorm"

// Repos 聚合数据访问层（对应 TECH_STACK internal/repo）
type Repos struct {
	Product      *ProductRepo
	Brand        *BrandRepo
	Category     *CategoryRepo
	Group        *GroupRepo
	PlatformType   *PlatformTypeRepo
	PlatformShop   *PlatformShopRepo
	PlatformListing *PlatformListingRepo
}

func New(db *gorm.DB) *Repos {
	return &Repos{
		Product:         NewProductRepo(db),
		Brand:           NewBrandRepo(db),
		Category:        NewCategoryRepo(db),
		Group:           NewGroupRepo(db),
		PlatformType:    NewPlatformTypeRepo(db),
		PlatformShop:    NewPlatformShopRepo(db),
		PlatformListing: NewPlatformListingRepo(db),
	}
}
