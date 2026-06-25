package dto

type SuperSearchQuery struct {
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

type SuperSearchItemDTO struct {
	ProductID      uint64            `json:"productId"`
	ProductName    string            `json:"productName"`
	MaterialCode   string            `json:"materialCode"`
	ProductSn      string            `json:"productSn"`
	ProductPic     string            `json:"productPic"`
	BrandName      string            `json:"brandName"`
	CategoryName   string            `json:"categoryName"`
	PublishStatus  int8              `json:"publishStatus"`
	SkuID          uint64            `json:"skuId"`
	SkuCode        string            `json:"skuCode"`
	Specs          map[string]string `json:"specs"`
	SpecLabel      string            `json:"specLabel"`
	Price          float64           `json:"price"`
	Stock          int               `json:"stock"`
	Pic            string            `json:"pic"`
	ListedShops    []ListedShopDTO   `json:"listedShops"`
	ListedShopCount int              `json:"listedShopCount"`
}
