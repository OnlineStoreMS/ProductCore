package dto

// IngestProductCollectRequest 浏览器扩展人工点至尊宝后回传的商品草稿。
type IngestProductCollectRequest struct {
	ProductURL string      `json:"productUrl"`
	Platform   string      `json:"platform"`
	Product    *ProductDTO `json:"product" binding:"required"`
}

type ProductCollectTaskDTO struct {
	ID           uint64 `json:"id"`
	ProductURL   string `json:"productUrl"`
	Platform     string `json:"platform"`
	PlatformName string `json:"platformName"`
	AgentJobID   uint64 `json:"agentJobId"`
	AgentID      uint64 `json:"agentId"`
	AgentName    string `json:"agentName"`
	ProductID    uint64 `json:"productId,omitempty"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	CreatedAt    string `json:"createdAt"`
}
