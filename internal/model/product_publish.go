package model

// ProductPublish 商品通过 WindowsAgent 铺到电商店铺的记录。
type ProductPublish struct {
	BaseModel
	TenantID          uint64 `gorm:"uniqueIndex:uk_product_publish;not null;default:1" json:"tenantId"`
	ProductID         uint64 `gorm:"uniqueIndex:uk_product_publish;index;not null" json:"productId"`
	Platform          string `gorm:"uniqueIndex:uk_product_publish;size:32;not null" json:"platform"`
	PlatformShopID    string `gorm:"uniqueIndex:uk_product_publish;size:128;not null" json:"platformShopId"`
	PlatformShopName  string `gorm:"size:256" json:"platformShopName"`
	AgentID           uint64 `gorm:"index" json:"agentId"`
	AgentJobID        uint64 `gorm:"index" json:"agentJobId"`
	Status            string `gorm:"size:16;index;not null" json:"status"` // running/succeeded/failed
	PlatformProductID string `gorm:"size:64" json:"platformProductId"`
	ErrorMessage      string `gorm:"type:text" json:"errorMessage"`
}

func (ProductPublish) TableName() string { return "product_publishes" }
