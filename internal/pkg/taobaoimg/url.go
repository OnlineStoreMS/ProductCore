package taobaoimg

import (
	"net/url"
	"regexp"
	"strings"

	"productcore/internal/dto"
)

var (
	sizeJpg  = regexp.MustCompile(`_\d+x\d+\.jpg`)
	sizePng  = regexp.MustCompile(`_\d+x\d+\.png`)
	reImgSrc = regexp.MustCompile(`(?i)(src=["'])(https?://[^"']+)(["'])`)
)

// Unwrap 把阿里 CDN 的 webp / 缩略图地址还原成原始 jpg/png。
// 例：.../O1CN01xxx_!!123.jpg_400x400q90.jpg_.webp → .../O1CN01xxx_!!123.jpg
func Unwrap(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "//") {
		s = "http:" + s
	}
	s = strings.ReplaceAll(s, ".jpg_.webp", ".jpg")
	if strings.Contains(s, "alicdn.com") {
		s = sizeJpg.ReplaceAllString(s, "")
		s = sizePng.ReplaceAllString(s, "")
	}
	if i := strings.Index(s, ".png_"); i >= 0 {
		s = s[:i+4]
	}
	if i := strings.Index(s, ".jpg_"); i >= 0 {
		s = s[:i+4]
	}
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return strings.TrimSpace(raw)
	}
	return s
}

func unwrapList(list []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(list))
	for _, raw := range list {
		u := Unwrap(raw)
		if u == "" {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

func unwrapHTML(html string) string {
	return reImgSrc.ReplaceAllStringFunc(html, func(s string) string {
		m := reImgSrc.FindStringSubmatch(s)
		if len(m) != 4 {
			return s
		}
		return m[1] + Unwrap(m[2]) + m[3]
	})
}

// UnwrapProduct 采集入库前把主图/详情/SKU 图还原成原始 jpg。
func UnwrapProduct(p *dto.ProductDTO) {
	if p == nil {
		return
	}
	p.Pic = Unwrap(p.Pic)
	p.AlbumPics = unwrapList(p.AlbumPics)
	if p.Pic == "" && len(p.AlbumPics) > 0 {
		p.Pic = p.AlbumPics[0]
	}
	if p.Media != nil {
		p.Media.Pics34 = unwrapList(p.Media.Pics34)
		p.Media.DetailPics = unwrapList(p.Media.DetailPics)
	}
	p.DetailHTML = unwrapHTML(p.DetailHTML)
	for i := range p.Skus {
		p.Skus[i].Pic = Unwrap(p.Skus[i].Pic)
	}
	for i := range p.SkuSpecs {
		for j := range p.SkuSpecs[i].Values {
			p.SkuSpecs[i].Values[j].Pic = Unwrap(p.SkuSpecs[i].Values[j].Pic)
		}
	}
}
