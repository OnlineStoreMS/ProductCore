package dto

type ListedShopDTO struct {
	ShopID           uint64 `json:"shopId"`
	ShopName         string `json:"shopName"`
	PlatformTypeName string `json:"platformTypeName,omitempty"`
	PlatformTypeLogo string `json:"platformTypeLogo,omitempty"`
	SourceChannel    string `json:"sourceChannel,omitempty"`
}

type UpdateProductListingsRequest struct {
	ShopIDs []uint64 `json:"shopIds"`
}

type PlatformShopTypeDTO struct {
	ID         uint64 `json:"id,omitempty"`
	Code       string `json:"code" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Logo       string `json:"logo"`
	Sort       int    `json:"sort"`
	Enabled    int8   `json:"enabled"`
	IsBuiltin  int8   `json:"isBuiltin,omitempty"`
	Remark     string `json:"remark"`
	ShopCount  int64  `json:"shopCount,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
}

type PlatformShopQuery struct {
	Keyword        string `form:"keyword"`
	PlatformTypeID uint64 `form:"platformTypeId"`
	Status         *int8  `form:"status"`
	Page           int    `form:"page"`
	PageSize       int    `form:"pageSize"`
}

type PlatformShopDTO struct {
	ID                 uint64 `json:"id,omitempty"`
	PlatformTypeID     uint64 `json:"platformTypeId" binding:"required"`
	PlatformTypeName   string `json:"platformTypeName,omitempty"`
	PlatformTypeLogo   string `json:"platformTypeLogo,omitempty"`
	Name               string `json:"name" binding:"required"`
	SourceChannel      string `json:"sourceChannel,omitempty"`
	ShopCode           string `json:"shopCode"`
	ExternalShopID     string `json:"externalShopId"`
	Status             int8   `json:"status"`
	Remark             string `json:"remark"`
	Sort               int    `json:"sort"`
	ListedProductCount int64  `json:"listedProductCount,omitempty"`
	CreateTime         string `json:"createTime,omitempty"`
	UpdateTime         string `json:"updateTime,omitempty"`
}
