package model

// ProductCollectTask 电商商品采集任务。链接下发到 WindowsAgent，结果回写本表。
type ProductCollectTask struct {
	BaseModel
	TenantID     uint64 `gorm:"index;not null;default:1" json:"tenantId"`
	UserID       uint64 `gorm:"index" json:"userId"`
	ProductURL   string `gorm:"size:2048;not null" json:"productUrl"`
	Platform     string `gorm:"size:32;not null" json:"platform"`
	AgentJobID   uint64 `gorm:"index" json:"agentJobId"`
	AgentID      uint64 `json:"agentId"`
	AgentName    string `gorm:"size:128" json:"agentName"`
	Status       string `gorm:"size:16;index;not null" json:"status"`
	ErrorMessage string `gorm:"type:text" json:"errorMessage"`
	ResultJSON   string `gorm:"type:text" json:"resultJson"`
}

func (ProductCollectTask) TableName() string { return "product_collect_tasks" }
