package dto

type SkuSpecDTO struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type SkuDTO struct {
	ID        uint64            `json:"id,omitempty"`
	SkuCode   string            `json:"skuCode"`
	Specs     map[string]string `json:"specs"`
	Price     float64           `json:"price"`
	CostPrice float64           `json:"costPrice"`
	Stock     int               `json:"stock"`
	Pic       string            `json:"pic,omitempty"`
}

type ProductDTO struct {
	ID             uint64       `json:"id,omitempty"`
	Name           string       `json:"name"`
	SubTitle       string       `json:"subTitle"`
	ProductSn      string       `json:"productSn"`
	BrandID        uint64       `json:"brandId"`
	BrandName      string       `json:"brandName,omitempty"`
	CategoryID     uint64       `json:"categoryId"`
	CategoryName   string       `json:"categoryName,omitempty"`
	GroupIDs       []uint64     `json:"groupIds"`
	Pic            string       `json:"pic"`
	AlbumPics      []string     `json:"albumPics"`
	ProductVideo   string       `json:"productVideo,omitempty"`
	Price          float64      `json:"price"`
	OriginalPrice  float64      `json:"originalPrice"`
	Stock          int          `json:"stock"`
	Unit           string       `json:"unit"`
	Weight         float64      `json:"weight"`
	PublishStatus  int8         `json:"publishStatus"`
	VerifyStatus   int8         `json:"verifyStatus"`
	Sort           int          `json:"sort"`
	Sale           int          `json:"sale,omitempty"`
	SkuCount       int          `json:"skuCount,omitempty"`
	Description    string       `json:"description"`
	DetailHTML     string       `json:"detailHtml"`
	SkuSpecs       []SkuSpecDTO `json:"skuSpecs"`
	Skus           []SkuDTO     `json:"skus"`
	ChannelVisible string       `json:"channelVisible,omitempty"`
	CreateTime     string       `json:"createTime,omitempty"`
	UpdateTime     string       `json:"updateTime,omitempty"`
}

type ProductQuery struct {
	Keyword       string `form:"keyword"`
	BrandID       uint64 `form:"brandId"`
	CategoryID    uint64 `form:"categoryId"`
	GroupID       uint64 `form:"groupId"`
	PublishStatus *int8  `form:"publishStatus"`
	Page          int    `form:"page"`
	PageSize      int    `form:"pageSize"`
}

type BrandDTO struct {
	ID           uint64 `json:"id,omitempty"`
	Name         string `json:"name" binding:"required"`
	Logo         string `json:"logo"`
	FirstLetter  string `json:"firstLetter"`
	Sort         int    `json:"sort"`
	ShowStatus   int8   `json:"showStatus"`
	ProductCount int64  `json:"productCount,omitempty"`
}

type CategoryDTO struct {
	ID           uint64        `json:"id,omitempty"`
	ParentID     uint64        `json:"parentId"`
	Name         string        `json:"name" binding:"required"`
	Level        int           `json:"level"`
	Sort         int           `json:"sort"`
	ShowStatus   int8          `json:"showStatus"`
	ProductCount int64         `json:"productCount,omitempty"`
	Children     []CategoryDTO `json:"children,omitempty"`
}

type ProductGroupDTO struct {
	ID           uint64 `json:"id,omitempty"`
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	Sort         int    `json:"sort"`
	ProductCount int64  `json:"productCount,omitempty"`
	CreateTime   string `json:"createTime,omitempty"`
}
