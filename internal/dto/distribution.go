package dto

type DistributionShopDTO struct {
	ID               uint64 `json:"id,omitempty"`
	PlatformTypeID   uint64 `json:"platformTypeId"`
	PlatformTypeName string `json:"platformTypeName,omitempty"`
	PlatformTypeLogo string `json:"platformTypeLogo,omitempty"`
	Name             string `json:"name"`
	Remark           string `json:"remark,omitempty"`
	ItemCount        int64  `json:"itemCount"`
	CreateTime       string `json:"createTime,omitempty"`
}

type DistributionShopQuery struct {
	Keyword        string `form:"keyword"`
	PlatformTypeID uint64 `form:"platformTypeId"`
	Page           int    `form:"page"`
	PageSize       int    `form:"pageSize"`
}

type DistributionItemDTO struct {
	ID             uint64 `json:"id"`
	ShopID         uint64 `json:"shopId"`
	ItemID         string `json:"itemId"`
	Title          string `json:"title"`
	ItemURL        string `json:"itemUrl"`
	PicURL         string `json:"picUrl"`
	Price          string `json:"price"`
	Sales          string `json:"sales"`
	CommentCount   string `json:"commentCount"`
	MonthDeals     string `json:"monthDeals"`
	MonthConsign   string `json:"monthConsign"`
	ShipTime       string `json:"shipTime"`
	ListedAt       string `json:"listedAt"`
	Category       string `json:"category"`
	Tags           string `json:"tags"`
	SourceShopName string `json:"sourceShopName"`
	SourceShopURL  string `json:"sourceShopUrl"`
	Collected      bool   `json:"collected"`
	ProductID      uint64 `json:"productId,omitempty"`
}

type DistributionItemQuery struct {
	Keyword   string `form:"keyword"`
	Collected string `form:"collected"` // 1 已采集 0 未采集
	SortBy    string `form:"sortBy"`
	SortOrder string `form:"sortOrder"` // asc desc
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
}

type DistributionImportResult struct {
	Created     int `json:"created"`
	Updated     int `json:"updated"`
	Removed     int `json:"removed"`
	Skipped     int `json:"skipped"`
	Total       int `json:"total"`
	ImageFailed int `json:"imageFailed"`
}
