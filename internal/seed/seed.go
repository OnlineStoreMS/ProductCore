package seed

import (
	"log"

	"productcore/internal/model"
	"productcore/internal/pkg/util"

	"gorm.io/gorm"
)

func Run(db *gorm.DB) {
	var count int64
	if err := db.Model(&model.Brand{}).Count(&count).Error; err != nil {
		log.Printf("seed check failed: %v", err)
		return
	}
	if count > 0 {
		log.Printf("seed skipped: brands table already has %d row(s)", count)
		return
	}
	log.Println("seeding demo data...")
	if err := db.Transaction(seedDemoData); err != nil {
		log.Printf("seed failed: %v", err)
		return
	}
	log.Println("seed completed")
}

// Force 清空演示表并重新 seed（开发用）
func Force(db *gorm.DB) error {
	log.Println("force seed: clearing demo tables...")
	tables := []string{
		"product_group_relations", "product_skus", "products",
		"product_groups", "categories", "brands",
	}
	for _, t := range tables {
		if err := db.Exec("TRUNCATE TABLE " + t + " RESTART IDENTITY CASCADE").Error; err != nil {
			return err
		}
	}
	log.Println("seeding demo data...")
	return db.Transaction(seedDemoData)
}

func seedDemoData(db *gorm.DB) error {
	brands := []model.Brand{
		{Name: "Apple", FirstLetter: "A", Sort: 100, ShowStatus: 1},
		{Name: "Nike", FirstLetter: "N", Sort: 90, ShowStatus: 1},
		{Name: "优衣库", FirstLetter: "Y", Sort: 80, ShowStatus: 1},
		{Name: "小米", FirstLetter: "X", Sort: 70, ShowStatus: 1},
	}
	if err := db.Create(&brands).Error; err != nil {
		return err
	}

	catDigital := model.Category{Name: "数码电器", Level: 0, Sort: 100, ShowStatus: 1}
	if err := db.Create(&catDigital).Error; err != nil {
		return err
	}
	catPhone := model.Category{ParentID: catDigital.ID, Name: "手机通讯", Level: 1, Sort: 10, ShowStatus: 1}
	if err := db.Create(&catPhone).Error; err != nil {
		return err
	}
	catSmartPhone := model.Category{ParentID: catPhone.ID, Name: "智能手机", Level: 2, Sort: 1, ShowStatus: 1}
	if err := db.Create(&catSmartPhone).Error; err != nil {
		return err
	}
	catShoes := model.Category{ParentID: catDigital.ID, Name: "运动鞋", Level: 1, Sort: 5, ShowStatus: 1}
	if err := db.Create(&catShoes).Error; err != nil {
		return err
	}

	groups := []model.ProductGroup{
		{Name: "春季上新", Description: "2026 春季主推", Sort: 10},
		{Name: "爆款精选", Description: "高转化商品", Sort: 9},
		{Name: "待上架抖店", Description: "待同步抖店", Sort: 8},
	}
	if err := db.Create(&groups).Error; err != nil {
		return err
	}

	specs := util.ToJSON([]map[string]interface{}{
		{"name": "颜色", "values": []string{"黑白", "全黑"}},
		{"name": "尺码", "values": []string{"41", "42"}},
	})
	product := model.Product{
		Name:           "Air Max 270 运动鞋",
		SubTitle:       "经典气垫跑鞋，舒适透气",
		MaterialCode:   "NKAM270",
		Source:         "淘宝",
		ProductSn:      "NK-AM270-001",
		BrandID:        brands[1].ID,
		CategoryID:     catShoes.ID,
		Pic:            "https://picsum.photos/seed/nike1/200/200",
		AlbumPics:      util.ToJSON([]string{"https://picsum.photos/seed/nike1/400/400"}),
		Price:          899,
		OriginalPrice:  1299,
		Stock:          75,
		Unit:           "双",
		Weight:         850,
		PublishStatus:  1,
		VerifyStatus:   1,
		Sort:           100,
		Description:    "Nike Air Max 270",
		DetailHTML:     "<p>商品详情</p>",
		SkuSpecsJSON:   specs,
		ChannelVisible: "both",
	}
	if err := db.Create(&product).Error; err != nil {
		return err
	}

	skus := []model.Sku{
		{ProductID: product.ID, SkuCode: "NK270BW41", SpecData: `{"颜色":"黑白","尺码":"41"}`, Price: 899, CostPrice: 450, Stock: 30, Pic: "https://picsum.photos/seed/nike1/100/100"},
		{ProductID: product.ID, SkuCode: "NK270BK42", SpecData: `{"颜色":"全黑","尺码":"42"}`, Price: 899, CostPrice: 450, Stock: 45, Pic: "https://picsum.photos/seed/nike2/100/100"},
	}
	if err := db.Create(&skus).Error; err != nil {
		return err
	}

	if err := db.Create([]model.ProductGroupRelation{
		{ProductID: product.ID, GroupID: groups[0].ID},
		{ProductID: product.ID, GroupID: groups[1].ID},
	}).Error; err != nil {
		return err
	}

	product2 := model.Product{
		Name:           "纯棉圆领 T 恤",
		SubTitle:       "100% 纯棉",
		MaterialCode:   "UQT001",
		Source:         "抖音",
		ProductSn:      "UQ-T001",
		BrandID:        brands[2].ID,
		CategoryID:     catDigital.ID,
		Pic:            "https://picsum.photos/seed/tshirt/200/200",
		Price:          79,
		OriginalPrice:  99,
		Stock:          200,
		Unit:           "件",
		PublishStatus:  0,
		VerifyStatus:   1,
		ChannelVisible: "offline",
	}
	if err := db.Create(&product2).Error; err != nil {
		return err
	}
	return db.Create(&model.ProductGroupRelation{ProductID: product2.ID, GroupID: groups[0].ID}).Error
}
