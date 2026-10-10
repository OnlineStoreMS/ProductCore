package model

// DistributionShop 铺货店铺。商品从至尊宝导出表导入，不写入商品库。
type DistributionShop struct {
	BaseModel
	TenantID       uint64 `gorm:"index;not null;default:1" json:"tenantId"`
	PlatformTypeID uint64 `gorm:"index;not null" json:"platformTypeId"`
	Name           string `gorm:"size:128;not null" json:"name"`
	Remark         string `gorm:"size:512" json:"remark"`
}

func (DistributionShop) TableName() string { return "distribution_shops" }

// DistributionShopItem 铺货店铺中的一条店铺商品。
type DistributionShopItem struct {
	BaseModel
	TenantID       uint64 `gorm:"uniqueIndex:uk_dist_shop_item;not null;default:1" json:"tenantId"`
	ShopID         uint64 `gorm:"uniqueIndex:uk_dist_shop_item;index;not null" json:"shopId"`
	ItemID         string `gorm:"uniqueIndex:uk_dist_shop_item;size:64;not null" json:"itemId"`
	Title          string `gorm:"size:300" json:"title"`
	ItemURL        string `gorm:"size:1024" json:"itemUrl"`
	PicURL         string `gorm:"size:1024" json:"picUrl"`
	Price          string `gorm:"size:64" json:"price"`
	Sales          string `gorm:"size:64" json:"sales"`
	CommentCount   string `gorm:"size:64" json:"commentCount"`
	MonthDeals     string `gorm:"size:64" json:"monthDeals"`
	MonthConsign   string `gorm:"size:64" json:"monthConsign"`
	ShipTime       string `gorm:"size:64" json:"shipTime"`
	ListedAt       string `gorm:"size:64" json:"listedAt"`
	Category       string `gorm:"size:128" json:"category"`
	Tags           string `gorm:"size:256" json:"tags"`
	SourceShopName string `gorm:"size:128" json:"sourceShopName"`
	SourceShopURL  string `gorm:"size:1024" json:"sourceShopUrl"`
}

func (DistributionShopItem) TableName() string { return "distribution_shop_items" }
