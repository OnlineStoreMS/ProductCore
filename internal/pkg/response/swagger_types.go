package response

import "productcore/internal/dto"

// Swagger 文档专用响应包装（swag 引用，运行时仍使用 Body + PageResult）

type ProductPageData struct {
	List     []dto.ProductDTO `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

type ProductPageResp struct {
	Code    int             `json:"code" example:"200"`
	Message string          `json:"message" example:"success"`
	Data    ProductPageData `json:"data"`
}

type ProductResp struct {
	Code    int            `json:"code" example:"200"`
	Message string         `json:"message" example:"success"`
	Data    dto.ProductDTO `json:"data"`
}

type ProductSkusResp struct {
	Code    int                `json:"code" example:"200"`
	Message string             `json:"message" example:"success"`
	Data    dto.ProductSkusDTO `json:"data"`
}

type BatchResultResp struct {
	Code    int             `json:"code" example:"200"`
	Message string          `json:"message" example:"success"`
	Data    dto.BatchResult `json:"data"`
}

type BrandListResp struct {
	Code    int            `json:"code" example:"200"`
	Message string         `json:"message" example:"success"`
	Data    []dto.BrandDTO `json:"data"`
}

type BrandResp struct {
	Code    int          `json:"code" example:"200"`
	Message string       `json:"message" example:"success"`
	Data    dto.BrandDTO `json:"data"`
}

type CategoryTreeResp struct {
	Code    int               `json:"code" example:"200"`
	Message string            `json:"message" example:"success"`
	Data    []dto.CategoryDTO `json:"data"`
}

type CategoryResp struct {
	Code    int             `json:"code" example:"200"`
	Message string          `json:"message" example:"success"`
	Data    dto.CategoryDTO `json:"data"`
}

type ProductGroupListResp struct {
	Code    int                   `json:"code" example:"200"`
	Message string                `json:"message" example:"success"`
	Data    []dto.ProductGroupDTO `json:"data"`
}

type ProductGroupResp struct {
	Code    int                  `json:"code" example:"200"`
	Message string               `json:"message" example:"success"`
	Data    dto.ProductGroupDTO  `json:"data"`
}

type ProductKeywordListResp struct {
	Code    int                     `json:"code" example:"200"`
	Message string                  `json:"message" example:"success"`
	Data    []dto.ProductKeywordDTO `json:"data"`
}

type ProductKeywordResp struct {
	Code    int                   `json:"code" example:"200"`
	Message string                `json:"message" example:"success"`
	Data    dto.ProductKeywordDTO `json:"data"`
}

type ListedShopListResp struct {
	Code    int                 `json:"code" example:"200"`
	Message string              `json:"message" example:"success"`
	Data    []dto.ListedShopDTO `json:"data"`
}

type PlatformTypeListResp struct {
	Code    int                        `json:"code" example:"200"`
	Message string                     `json:"message" example:"success"`
	Data    []dto.PlatformShopTypeDTO  `json:"data"`
}

type PlatformTypeResp struct {
	Code    int                       `json:"code" example:"200"`
	Message string                    `json:"message" example:"success"`
	Data    dto.PlatformShopTypeDTO   `json:"data"`
}

type PlatformShopPageData struct {
	List     []dto.PlatformShopDTO `json:"list"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
}

type PlatformShopPageResp struct {
	Code    int                  `json:"code" example:"200"`
	Message string               `json:"message" example:"success"`
	Data    PlatformShopPageData `json:"data"`
}

type PlatformShopResp struct {
	Code    int                  `json:"code" example:"200"`
	Message string               `json:"message" example:"success"`
	Data    dto.PlatformShopDTO  `json:"data"`
}

type UploadURLData struct {
	URL string `json:"url"`
}

type UploadURLResp struct {
	Code    int           `json:"code" example:"200"`
	Message string        `json:"message" example:"success"`
	Data    UploadURLData `json:"data"`
}

type BatchUploadFailedItem struct {
	Filename string `json:"filename"`
	Message  string `json:"message"`
}

type BatchUploadData struct {
	URLs   []string                `json:"urls"`
	Failed []BatchUploadFailedItem `json:"failed"`
}

type BatchUploadResp struct {
	Code    int             `json:"code" example:"200"`
	Message string          `json:"message" example:"success"`
	Data    BatchUploadData `json:"data"`
}

type EmptyResp struct {
	Code    int    `json:"code" example:"200"`
	Message string `json:"message" example:"success"`
}
