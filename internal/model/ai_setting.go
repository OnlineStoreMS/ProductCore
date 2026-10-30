package model

import "time"

// AISetting 租户级 AI 能力默认配置（修图模型等）。
type AISetting struct {
	ID                       uint64    `gorm:"primaryKey"`
	TenantID                 uint64    `gorm:"uniqueIndex;not null;default:1"`
	DefaultImageModel        string    `gorm:"size:128;not null;default:''"`
	DefaultImageConnectionID uint64    `gorm:"not null;default:0"`
	UpdatedAt                time.Time `gorm:"autoUpdateTime"`
	CreatedAt                time.Time `gorm:"autoCreateTime"`
}

func (AISetting) TableName() string { return "ai_settings" }
