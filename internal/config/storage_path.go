package config

import (
	"path/filepath"
	"strings"
)

// EffectivePrefix 统一资源前缀（local / minio 目录规则一致）
func (c *StorageConfig) EffectivePrefix() string {
	if p := strings.Trim(c.Prefix, "/"); p != "" {
		return p
	}
	if p := strings.Trim(c.MinIO.Prefix, "/"); p != "" {
		return p
	}
	return "uploads"
}

// LocalContentDir 本地存储内容根目录（等价于 minio bucket 内的 prefix 目录）
// 支持 legacy：local_path 已设为 ./data/uploads 时不再重复拼接 prefix
func (c *StorageConfig) LocalContentDir() string {
	base := c.LocalPath
	if base == "" {
		base = "./data/uploads"
	}
	prefix := c.EffectivePrefix()
	clean := filepath.Clean(base)
	if filepath.Base(clean) == prefix {
		return clean
	}
	return filepath.Join(clean, prefix)
}

// ObjectKeyPrefix 生成对象 key 时使用的前缀段（local 内容根已含 prefix 目录，故为空）
func (c *StorageConfig) ObjectKeyPrefix(driver string) string {
	if driver == "local" || driver == "" {
		return ""
	}
	return c.EffectivePrefix()
}
