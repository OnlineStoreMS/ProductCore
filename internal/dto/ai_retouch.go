package dto

type AIRetouchRequest struct {
	ImageURL     string `json:"imageUrl"`
	Prompt       string `json:"prompt"`
	Model        string `json:"model"`
	ConnectionID uint64 `json:"connectionId"`
	Scope        string `json:"scope"`
	Resource     string `json:"resource"`
	ProductID    uint64 `json:"productId"`
	SkuID        uint64 `json:"skuId"`
}
