package dto

type SkuSpecValueDTO struct {
	Value  string `json:"value"`
	Remark string `json:"remark,omitempty"`
	Pic    string `json:"pic,omitempty"`
}

type SkuSpecDTO struct {
	Name   string            `json:"name"`
	Values []SkuSpecValueDTO `json:"values"`
}

type ProductVideosDTO struct {
	Ratio11  string `json:"ratio11,omitempty"`
	Ratio34  string `json:"ratio34,omitempty"`
	Ratio169 string `json:"ratio169,omitempty"`
	Ratio916 string `json:"ratio916,omitempty"`
}

type ProductMaterialsDTO struct {
	White       string `json:"white,omitempty"`
	Transparent string `json:"transparent,omitempty"`
	Guide34     string `json:"guide34,omitempty"`
	Long        string `json:"long,omitempty"`
}

type ProductMediaDTO struct {
	Pics34     []string             `json:"pics34,omitempty"`
	Videos     *ProductVideosDTO    `json:"videos,omitempty"`
	Materials  *ProductMaterialsDTO `json:"materials,omitempty"`
	DetailPics []string             `json:"detailPics,omitempty"`
}

type SkuDTO struct {
	ID          uint64            `json:"id,omitempty"`
	SkuCode     string            `json:"skuCode"`
	Specs       map[string]string `json:"specs"`
	Price       float64           `json:"price"`
	CostPrice   float64           `json:"costPrice"`
	MarketPrice float64           `json:"marketPrice,omitempty"`
	Stock       int               `json:"stock"`
	Weight      float64           `json:"weight,omitempty"`
	Pic         string            `json:"pic,omitempty"`
}

type ProductDTO struct {
	ID             uint64           `json:"id,omitempty"`
	Name           string           `json:"name"`
	SubTitle       string           `json:"subTitle"`
	MaterialCode   string           `json:"materialCode"`
	Source         string           `json:"source"`
	ProductSn      string           `json:"productSn"`
	BrandID        uint64           `json:"brandId"`
	BrandName      string           `json:"brandName,omitempty"`
	CategoryID     uint64           `json:"categoryId"`
	CategoryName   string           `json:"categoryName,omitempty"`
	GroupIDs       []uint64         `json:"groupIds"`
	Pic            string           `json:"pic"`
	AlbumPics      []string         `json:"albumPics"`
	ProductVideo   string           `json:"productVideo,omitempty"`
	Media          *ProductMediaDTO `json:"media,omitempty"`
	Price          float64          `json:"price"`
	OriginalPrice  float64          `json:"originalPrice"`
	Stock          int              `json:"stock"`
	Unit           string           `json:"unit"`
	Weight         float64          `json:"weight"`
	PublishStatus  int8             `json:"publishStatus"`
	IsDraft        int8             `json:"isDraft"`
	HasEditDraft   bool             `json:"hasEditDraft,omitempty"` // 已发布商品有未合并的编辑草稿
	DraftSavedAt   string           `json:"draftSavedAt,omitempty"` // 草稿箱：上次草稿保存时间
	VerifyStatus   int8             `json:"verifyStatus"`
	Sort           int              `json:"sort"`
	Sale           int              `json:"sale,omitempty"`
	SkuCount       int              `json:"skuCount,omitempty"`
	Description    string           `json:"description"`
	DetailHTML     string           `json:"detailHtml"`
	SkuSpecs       []SkuSpecDTO     `json:"skuSpecs"`
	Skus           []SkuDTO         `json:"skus"`
	ChannelVisible string           `json:"channelVisible,omitempty"`
	ListedShops    []ListedShopDTO  `json:"listedShops,omitempty"`
	ListedShopCount int             `json:"listedShopCount,omitempty"`
	CreateTime     string           `json:"createTime,omitempty"`
	UpdateTime     string           `json:"updateTime,omitempty"`
	DeleteTime     string           `json:"deleteTime,omitempty"`
	Finalize       bool             `json:"finalize,omitempty"` // true=完成保存并移出草稿箱
}

type ProductQuery struct {
	Keyword       string   `form:"keyword"`
	BrandID       uint64   `form:"brandId"`
	CategoryID    uint64   `form:"categoryId"`
	CategoryIDs   []uint64 `form:"-"` // 内部：含子孙分类
	GroupID       uint64   `form:"groupId"`
	PublishStatus *int8    `form:"publishStatus"`
	Page          int      `form:"page"`
	PageSize      int      `form:"pageSize"`
}

type BatchIDsRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1,dive,gt=0"`
}

type BatchResult struct {
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

type UpdateSkusRequest struct {
	Skus     []SkuDTO     `json:"skus"`
	SkuSpecs []SkuSpecDTO `json:"skuSpecs,omitempty"`
}

// ProductSkusDTO SKU 管理轻量视图（不含详情、媒体等大字段）
type ProductSkusDTO struct {
	ID            uint64       `json:"id"`
	Name          string       `json:"name"`
	MaterialCode  string       `json:"materialCode"`
	SkuCount      int          `json:"skuCount"`
	Price         float64      `json:"price"`
	OriginalPrice float64      `json:"originalPrice"`
	Stock         int          `json:"stock"`
	SkuSpecs      []SkuSpecDTO `json:"skuSpecs"`
	Skus          []SkuDTO     `json:"skus"`
}

type UpdatePublishStatusRequest struct {
	PublishStatus int8 `json:"publishStatus"`
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
	Icon         string        `json:"icon,omitempty"`
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
