package seed

import (
	"log"

	"productcore/internal/model"

	"gorm.io/gorm"
)

func SeedPlatformTypes(db *gorm.DB) {
	var count int64
	if err := db.Model(&model.PlatformShopType{}).Count(&count).Error; err != nil {
		log.Printf("platform type seed check failed: %v", err)
		return
	}
	if count > 0 {
		return
	}
	types := []model.PlatformShopType{
		{Code: "douyin", Name: "抖音", Sort: 100, Enabled: 1, IsBuiltin: 1},
		{Code: "taobao", Name: "淘宝", Sort: 90, Enabled: 1, IsBuiltin: 1},
		{Code: "pdd", Name: "拼多多", Sort: 80, Enabled: 1, IsBuiltin: 1},
		{Code: "xhs", Name: "小红书", Sort: 70, Enabled: 1, IsBuiltin: 1},
		{Code: "xianyu", Name: "闲鱼", Sort: 60, Enabled: 1, IsBuiltin: 1},
		{Code: "custom", Name: "自定义店铺", Sort: 10, Enabled: 1, IsBuiltin: 1, Remark: "通用自定义渠道"},
	}
	if err := db.Create(&types).Error; err != nil {
		log.Printf("platform type seed failed: %v", err)
		return
	}
	log.Printf("seeded %d platform shop types", len(types))
}
