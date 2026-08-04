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
	TenantID     uint64 `gorm:"index;not null;default:1" json:"tenantId"`
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
	TenantID   uint64 `gorm:"index;not null;default:1" json:"tenantId"`
	ParentID   uint64 `gorm:"default:0;index" json:"parentId"`
	Name       string `gorm:"size:64;not null" json:"name"`
	Level      int    `gorm:"default:0" json:"level"`
	Sort       int    `gorm:"default:0" json:"sort"`
	ShowStatus int8   `gorm:"default:1" json:"showStatus"`
	Icon       string `gorm:"size:512" json:"icon"`
}

func (Category) TableName() string { return "categories" }

// ProductGroup 商品分组（树形，支持子分组）
type ProductGroup struct {
	BaseModel
	TenantID    uint64 `gorm:"index;not null;default:1" json:"tenantId"`
	ParentID    uint64 `gorm:"default:0;index" json:"parentId"`
	Name        string `gorm:"size:64;not null" json:"name"`
	Description string `gorm:"size:512" json:"description"`
	Level       int    `gorm:"default:0" json:"level"`
	Sort        int    `gorm:"default:0" json:"sort"`
}

func (ProductGroup) TableName() string { return "product_groups" }

// Product SPU 商品
type Product struct {
	BaseModel
	TenantID       uint64  `gorm:"index;not null;default:1" json:"tenantId"`
	Name           string  `gorm:"size:200;not null" json:"name"`
	SubTitle       string  `gorm:"size:255" json:"subTitle"`
	MaterialCode   string  `gorm:"size:64" json:"materialCode"` // 来源资料编码，选填，非空时唯一
	Source         string  `gorm:"size:64" json:"source"`                   // 商品来源
	ProductSn      string  `gorm:"size:64" json:"productSn"`                // 货号（可选）
	BrandID        uint64  `gorm:"index" json:"brandId"`
	CategoryID     uint64  `gorm:"index" json:"categoryId"`
	Pic            string  `gorm:"size:512" json:"pic"`
	AlbumPics      string  `gorm:"type:text" json:"-"` // JSON array
	Pics34JSON     string  `gorm:"type:text" json:"-"` // 3:4 主图 JSON array
	ProductVideo   string  `gorm:"size:512" json:"productVideo"`
	Price          float64 `gorm:"type:decimal(10,2);default:0" json:"price"`
	OriginalPrice  float64 `gorm:"type:decimal(10,2);default:0" json:"originalPrice"`
	Stock          int     `gorm:"default:0" json:"stock"`
	Unit           string  `gorm:"size:16;default:件" json:"unit"`
	Weight         float64 `gorm:"type:decimal(10,2);default:0" json:"weight"`
	PublishStatus  int8    `gorm:"default:0" json:"publishStatus"`
	IsDraft        int8    `gorm:"default:0;index" json:"isDraft"` // 1=草稿箱
	VerifyStatus   int8    `gorm:"default:1" json:"verifyStatus"`
	Sort           int     `gorm:"default:0" json:"sort"`
	Sale           int     `gorm:"default:0" json:"sale"`
	Description    string  `gorm:"type:text" json:"description"`
	DetailHTML     string  `gorm:"type:text" json:"detailHtml"`
	SkuSpecsJSON   string  `gorm:"type:text" json:"-"` // 规格维度 JSON
	MediaJSON      string  `gorm:"type:text" json:"-"` // 扩展媒体 JSON: pics34/videos/materials/detailPics
	ChannelVisible string  `gorm:"size:16;default:both" json:"channelVisible"` // online|offline|both

	Brand    *Brand    `gorm:"foreignKey:BrandID" json:"-"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"-"`
	Skus     []Sku     `gorm:"foreignKey:ProductID" json:"-"`
}

func (Product) TableName() string { return "products" }

// Sku SKU
type Sku struct {
	BaseModel
	TenantID  uint64  `gorm:"index;not null;default:1" json:"tenantId"`
	ProductID uint64  `gorm:"index;not null" json:"productId"`
	SkuCode   string  `gorm:"size:64;not null" json:"skuCode"` // 租户内唯一，仅 A-Za-z0-9
	SpecData  string  `gorm:"type:text" json:"-"`                          // JSON: {"颜色":"红","尺码":"L"}
	SortOrder int     `gorm:"default:0;index" json:"sortOrder"`            // 列表/导出序号，从 0 起
	Price       float64 `gorm:"type:decimal(10,2);default:0" json:"price"`
	CostPrice   float64 `gorm:"type:decimal(10,2);default:0" json:"costPrice"`
	MarketPrice float64 `gorm:"type:decimal(10,2);default:0" json:"marketPrice"`
	Stock       int     `gorm:"default:0" json:"stock"`
	Weight      float64 `gorm:"type:decimal(10,2);default:0" json:"weight"` // 克(g)
	Pic         string  `gorm:"size:512" json:"pic"`
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

// ProductKeyword 商品关键词（细分类标签）
type ProductKeyword struct {
	BaseModel
	TenantID    uint64 `gorm:"index;not null;default:1" json:"tenantId"`
	Name        string `gorm:"size:64;not null" json:"name"`
	Description string `gorm:"size:512" json:"description"`
	Sort        int    `gorm:"default:0" json:"sort"`
}

func (ProductKeyword) TableName() string { return "product_keywords" }

// ProductKeywordRelation 商品-关键词关联
type ProductKeywordRelation struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	ProductID uint64 `gorm:"index;not null;uniqueIndex:uk_product_keyword"`
	KeywordID uint64 `gorm:"index;not null;uniqueIndex:uk_product_keyword"`
}

func (ProductKeywordRelation) TableName() string { return "product_keyword_relations" }

// PlatformShopType 电商平台类型（抖店、淘宝、拼多多等）
type PlatformShopType struct {
	BaseModel
	Code      string `gorm:"size:32;uniqueIndex;not null" json:"code"`
	Name      string `gorm:"size:64;not null" json:"name"`
	Logo      string `gorm:"size:512" json:"logo"`
	Sort      int    `gorm:"default:0" json:"sort"`
	Enabled   int8   `gorm:"default:1" json:"enabled"`
	IsBuiltin int8   `gorm:"default:0" json:"isBuiltin"`
	Remark    string `gorm:"size:256" json:"remark"`
}

func (PlatformShopType) TableName() string { return "platform_shop_types" }

// PlatformShop 平台店铺
type PlatformShop struct {
	BaseModel
	TenantID       uint64 `gorm:"index;not null;default:1" json:"tenantId"`
	PlatformTypeID uint64 `gorm:"index;not null" json:"platformTypeId"`
	Name           string `gorm:"size:128;not null" json:"name"`
	SourceChannel  string `gorm:"size:32;not null" json:"sourceChannel"`
	ShopCode       string `gorm:"size:64" json:"shopCode"`
	ExternalShopID string `gorm:"size:128" json:"externalShopId"`
	Status         int8   `gorm:"default:1" json:"status"`
	Remark         string `gorm:"size:512" json:"remark"`
	Sort           int    `gorm:"default:0" json:"sort"`
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

// PlatformListing 商品在各店铺的铺货记录（SPU 级）
type PlatformListing struct {
	BaseModel
	TenantID       uint64 `gorm:"index;not null;default:1" json:"tenantId"`
	ProductID      uint64 `gorm:"uniqueIndex:idx_tenant_product_shop;not null" json:"productId"`
	PlatformShopID uint64 `gorm:"uniqueIndex:idx_tenant_product_shop;not null" json:"platformShopId"`
	ListingStatus  int8   `gorm:"default:1" json:"listingStatus"`
	Remark         string `gorm:"size:256" json:"remark"`
}

func (PlatformListing) TableName() string { return "platform_listings" }

// ProductEditDraft 已发布商品的编辑草稿（auto-save，不影响 products.is_draft）
type ProductEditDraft struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID   uint64    `gorm:"uniqueIndex;not null" json:"productId"`
	PayloadJSON string    `gorm:"type:text;not null" json:"-"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (ProductEditDraft) TableName() string { return "product_edit_drafts" }
