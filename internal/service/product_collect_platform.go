package service

import (
	"fmt"
	"net/url"
	"strings"
)

// DetectCollectPlatform 识别商品链接所属平台。当前只接受淘宝、天猫（含短链 tb.cn）。
func DetectCollectPlatform(raw string) (platform, normalized string, err error) {
	normalized = strings.TrimSpace(raw)
	if normalized == "" {
		return "", "", fmt.Errorf("%w", ErrUnsupportedProductURL)
	}
	if len(normalized) > 2048 {
		return "", "", fmt.Errorf("%w", ErrUnsupportedProductURL)
	}
	u, parseErr := url.Parse(normalized)
	if parseErr != nil || u.Host == "" {
		return "", "", fmt.Errorf("%w", ErrUnsupportedProductURL)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", "", fmt.Errorf("%w", ErrUnsupportedProductURL)
	}
	host := strings.ToLower(u.Hostname())
	if hostIs(host, "taobao.com") || hostIs(host, "tmall.com") || hostIs(host, "tb.cn") {
		return "taobao", normalized, nil
	}
	return "", "", fmt.Errorf("%w", ErrUnsupportedProductURL)
}

func hostIs(host, root string) bool {
	return host == root || strings.HasSuffix(host, "."+root)
}
