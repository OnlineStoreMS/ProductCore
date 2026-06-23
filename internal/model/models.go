package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Brand 品牌
type Brand struct {
	BaseModel
	Name         string `gorm:"size:64;not null" json:"name"`
	Logo         string `gorm:"size:512" json:"logo"`
	FirstLetter  string `gorm:"size:1" json:"firstLetter"`
	Sort         int    `gorm:"default:0" json:"sort"`
	ShowStatus   int8   `gorm:"default:1" json:"showStatus"` // 0隐藏 1显示
	Description  string `gorm:"size:512" json:"description"`
}

func (Brand) TableName() string { return "brands" }

// Category 商品分类（树形）
type Category struct {
	BaseModel
	ParentID   uint64 `gorm:"default:0;index" json:"parentId"`
	Name       string `gorm:"size:64;not null" json:"name"`
	Level      int    `gorm:"default:0" json:"level"`
	Sort       int    `gorm:"default:0" json:"sort"`
	ShowStatus int8   `gorm:"default:1" json:"showStatus"`
	Icon       string `gorm:"size:512" json:"icon"`
}

func (Category) TableName() string { return "categories" }

// ProductGroup 商品分组
type ProductGroup struct {
	BaseModel
	Name        string `gorm:"size:64;not null" json:"name"`
	Description string `gorm:"size:512" json:"description"`
	Sort        int    `gorm:"default:0" json:"sort"`
}

func (ProductGroup) TableName() string { return "product_groups" }

// Product SPU 商品
type Product struct {
	BaseModel
	Name           string  `gorm:"size:200;not null" json:"name"`
	SubTitle       string  `gorm:"size:255" json:"subTitle"`
	ProductSn      string  `gorm:"size:64;uniqueIndex" json:"productSn"`
	BrandID        uint64  `gorm:"index" json:"brandId"`
	CategoryID     uint64  `gorm:"index" json:"categoryId"`
	Pic            string  `gorm:"size:512" json:"pic"`
	AlbumPics      string  `gorm:"type:text" json:"-"` // JSON array
	ProductVideo   string  `gorm:"size:512" json:"productVideo"`
	Price          float64 `gorm:"type:decimal(10,2);default:0" json:"price"`
	OriginalPrice  float64 `gorm:"type:decimal(10,2);default:0" json:"originalPrice"`
	Stock          int     `gorm:"default:0" json:"stock"`
	Unit           string  `gorm:"size:16;default:件" json:"unit"`
	Weight         float64 `gorm:"type:decimal(10,2);default:0" json:"weight"`
	PublishStatus  int8    `gorm:"default:0" json:"publishStatus"`
	VerifyStatus   int8    `gorm:"default:1" json:"verifyStatus"`
	Sort           int     `gorm:"default:0" json:"sort"`
	Sale           int     `gorm:"default:0" json:"sale"`
	Description    string  `gorm:"type:text" json:"description"`
	DetailHTML     string  `gorm:"type:longtext" json:"detailHtml"`
	SkuSpecsJSON   string  `gorm:"type:text" json:"-"` // 规格维度 JSON
	ChannelVisible string  `gorm:"size:16;default:both" json:"channelVisible"` // online|offline|both

	Brand    *Brand    `gorm:"foreignKey:BrandID" json:"-"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"-"`
	Skus     []Sku     `gorm:"foreignKey:ProductID" json:"-"`
}

func (Product) TableName() string { return "products" }

// Sku SKU
type Sku struct {
	BaseModel
	ProductID uint64  `gorm:"index;not null" json:"productId"`
	SkuCode   string  `gorm:"size:64;uniqueIndex;not null" json:"skuCode"`
	SpecData  string  `gorm:"type:text" json:"-"` // JSON: {"颜色":"红","尺码":"L"}
	Price     float64 `gorm:"type:decimal(10,2);default:0" json:"price"`
	CostPrice float64 `gorm:"type:decimal(10,2);default:0" json:"costPrice"`
	Stock     int     `gorm:"default:0" json:"stock"`
	Pic       string  `gorm:"size:512" json:"pic"`
	Sale      int     `gorm:"default:0" json:"sale"`
}

func (Sku) TableName() string { return "product_skus" }

// ProductGroupRelation 商品-分组关联
type ProductGroupRelation struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	ProductID uint64 `gorm:"index;not null"`
	GroupID   uint64 `gorm:"index;not null"`
}

func (ProductGroupRelation) TableName() string { return "product_group_relations" }

// PlatformShop 平台店铺（预留）
type PlatformShop struct {
	BaseModel
	Name          string `gorm:"size:128;not null" json:"name"`
	SourceChannel string `gorm:"size:32;not null" json:"sourceChannel"`
	ShopCode      string `gorm:"size:64" json:"shopCode"`
	Status        int8   `gorm:"default:1" json:"status"`
}

func (PlatformShop) TableName() string { return "platform_shops" }

// PlatformSkuMapping 平台 SKU 映射（预留）
type PlatformSkuMapping struct {
	BaseModel
	SkuID            uint64 `gorm:"index;not null" json:"skuId"`
	PlatformShopID   uint64 `gorm:"index;not null" json:"platformShopId"`
	PlatformSkuID    string `gorm:"size:128" json:"platformSkuId"`
	PlatformProductID string `gorm:"size:128" json:"platformProductId"`
}

func (PlatformSkuMapping) TableName() string { return "platform_sku_mappings" }
