package util

import (
	"math/rand"
	"regexp"
	"strings"
	"time"
)

const SkuCodeMaxLen = 64

var skuCodePattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// SanitizeSkuCode 只保留字母与数字
func SanitizeSkuCode(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
		if b.Len() >= SkuCodeMaxLen {
			break
		}
	}
	return b.String()
}

// IsValidSkuCode 校验规格编码格式
func IsValidSkuCode(code string) bool {
	code = strings.TrimSpace(code)
	if code == "" || len(code) > SkuCodeMaxLen {
		return false
	}
	return skuCodePattern.MatchString(code)
}

const skuCodeRandomLen = 8

var skuCodeRandomChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// NextRandomSkuCode 生成随机规格编码，确保在 used 内唯一
func NextRandomSkuCode(used map[string]struct{}) string {
	for attempt := 0; attempt < 10000; attempt++ {
		b := make([]byte, skuCodeRandomLen)
		for i := range b {
			b[i] = skuCodeRandomChars[rand.Intn(len(skuCodeRandomChars))]
		}
		code := string(b)
		if _, dup := used[code]; dup {
			continue
		}
		used[code] = struct{}{}
		return code
	}
	fallback := SanitizeSkuCode(time.Now().Format("150405999"))
	if _, dup := used[fallback]; !dup {
		used[fallback] = struct{}{}
	}
	return fallback
}
