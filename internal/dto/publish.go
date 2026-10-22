package dto

type PublishShopDTO struct {
	ID                uint64 `json:"id"`
	AgentID           uint64 `json:"agentId"`
	AgentName         string `json:"agentName"`
	Platform          string `json:"platform"`
	PlatformShopID    string `json:"platformShopId"`
	PlatformShopName  string `json:"platformShopName"`
	AgentOnline       bool   `json:"agentOnline"`
	PublishReady      bool   `json:"publishReady"`
	Reason            string `json:"reason,omitempty"`
	PublishStatus     string `json:"publishStatus,omitempty"`
	PlatformProductID string `json:"platformProductId,omitempty"`
}

type PublishShopListDTO struct {
	List []PublishShopDTO `json:"list"`
	Hint string           `json:"hint,omitempty"`
}

type StartPublishRequest struct {
	Platform       string `json:"platform" binding:"required"`
	PlatformShopID string `json:"platformShopId" binding:"required"`
}

type StartPublishUpdateRequest struct {
	Platform          string `json:"platform" binding:"required"`
	PlatformShopID    string `json:"platformShopId" binding:"required"`
	PlatformProductID string `json:"platformProductId" binding:"required"`
	// all 覆盖标题和图片；sku 只更新商品规格。空值按 all。
	Scope string `json:"scope,omitempty"`
	// 全部信息时使用。空则用商品现有标题。
	Title string `json:"title,omitempty"`
}

type ListingTitleResult struct {
	Title string `json:"title"`
}

type StartPublishResult struct {
	ID               uint64 `json:"id"`
	ShopName         string `json:"shopName"`
	Platform         string `json:"platform"`
	PlatformTypeName string `json:"platformTypeName"`
	PublishStatus    string `json:"publishStatus"`
	AgentJobID       uint64 `json:"agentJobId"`
}
