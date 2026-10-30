package dto

type AISettingDTO struct {
	DefaultImageModel        string `json:"defaultImageModel"`
	DefaultImageConnectionID uint64 `json:"defaultImageConnectionId"`
	DefaultImageLabel        string `json:"defaultImageLabel,omitempty"`
}

type UpdateAISettingRequest struct {
	DefaultImageModel        string `json:"defaultImageModel"`
	DefaultImageConnectionID uint64 `json:"defaultImageConnectionId"`
}

type AIImageModelDTO struct {
	ID             uint64 `json:"id"`
	ConnectionID   uint64 `json:"connectionId"`
	ConnectionName string `json:"connectionName"`
	ProviderCode   string `json:"providerCode"`
	ModelID        string `json:"modelId"`
	DisplayName    string `json:"displayName"`
	IsDefault      int    `json:"isDefault"`
}
