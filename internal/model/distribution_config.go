package model

// DistributionPublishConfig 某个电商平台的铺货配置。同一租户、同一平台只保留一条。
// 当前上架仍使用店铺页面上已选好的默认值，这里先保存，后续再接入执行。
type DistributionPublishConfig struct {
	BaseModel
	TenantID        uint64 `gorm:"uniqueIndex:uk_dist_publish_cfg;not null;default:1" json:"tenantId"`
	PlatformTypeID  uint64 `gorm:"uniqueIndex:uk_dist_publish_cfg;not null" json:"platformTypeId"`
	ShipMode        string `gorm:"size:32;not null" json:"shipMode"`
	SpotShipTime    string `gorm:"size:32;not null" json:"spotShipTime"`
	FreightTemplate string `gorm:"size:128" json:"freightTemplate"`
	AfterSalePolicy string `gorm:"size:32;not null" json:"afterSalePolicy"`
	Enabled         int8   `gorm:"not null;default:1" json:"enabled"`
	Remark          string `gorm:"size:256" json:"remark"`
}

func (DistributionPublishConfig) TableName() string { return "distribution_publish_configs" }
