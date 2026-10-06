package service

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"productcore/internal/dto"
	"productcore/internal/pkg/util"
	"productcore/internal/storage"
)

var reCollectImgSrc = regexp.MustCompile(`(?i)(src=["'])(https?://[^"']+)(["'])`)

func needsRehost(raw string) bool {
	s := strings.TrimSpace(raw)
	if s == "" {
		return false
	}
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	host := strings.ToLower(u.Host)
	return strings.Contains(host, "alicdn.com") ||
		strings.Contains(host, "taobao.com") ||
		strings.Contains(host, "tmall.com") ||
		strings.Contains(host, "alibaba.com") ||
		strings.Contains(host, "aliyuncs.com")
}

// ingestRemoteImages 把采集到的外链图片下载到自有存储，并回写商品与 SKU。
func (s *ProductService) ingestRemoteImages(id uint64, in *dto.ProductDTO) (ok, failed int, err error) {
	if s == nil || s.store == nil || id == 0 || in == nil {
		return 0, 0, nil
	}
	cache := map[string]string{}
	var firstErr string
	resolve := func(raw, resource string) string {
		raw = strings.TrimSpace(raw)
		if raw == "" || !needsRehost(raw) {
			return raw
		}
		if uploaded, seen := cache[raw]; seen {
			return uploaded
		}
		uploaded, upErr := storage.UploadFromURL(s.store, raw, storage.UploadOptions{
			Scope:     "product",
			ProductID: id,
			Resource:  resource,
		})
		if upErr != nil {
			failed++
			if firstErr == "" {
				firstErr = upErr.Error()
			}
			cache[raw] = ""
			return ""
		}
		ok++
		cache[raw] = uploaded
		return uploaded
	}

	in.Pic = resolve(in.Pic, "main")
	in.AlbumPics = compactRemoteURLs(resolveList(in.AlbumPics, "album", resolve))
	in.AlbumPics = dropCoverDuplicate(in.Pic, in.AlbumPics)
	if in.Media != nil {
		in.Media.DetailPics = compactRemoteURLs(resolveList(in.Media.DetailPics, "detail", resolve))
		in.Media.Pics34 = compactRemoteURLs(resolveList(in.Media.Pics34, "pics34", resolve))
		if in.Media.Materials != nil {
			m := in.Media.Materials
			m.White = resolve(m.White, "material_white")
			m.Transparent = resolve(m.Transparent, "material_transparent")
			m.Guide34 = resolve(m.Guide34, "material_guide34")
			m.Long = resolve(m.Long, "material_long")
		}
	}
	in.DetailHTML = rewriteCollectImgSrc(in.DetailHTML, func(raw string) string {
		return resolve(raw, "detail")
	})
	for i := range in.Skus {
		in.Skus[i].Pic = resolve(in.Skus[i].Pic, "sku_pic")
	}
	for i := range in.SkuSpecs {
		for j := range in.SkuSpecs[i].Values {
			in.SkuSpecs[i].Values[j].Pic = resolve(in.SkuSpecs[i].Values[j].Pic, "sku_pic")
		}
	}
	if ok == 0 && failed == 0 {
		return 0, 0, nil
	}

	pics34 := dto.Pics34FromMedia(in.Media)
	media := dto.ProductMediaWithoutPics34(in.Media)
	mediaJSON := ""
	if media != nil {
		mediaJSON = util.ToJSON(media)
	}
	skuSpecsJSON := "[]"
	if len(in.Skus) > 0 {
		skuSpecsJSON = util.ToJSON(dto.CompactSkuSpecs(in.SkuSpecs))
	}
	if err := s.repo.UpdateFields(id, map[string]interface{}{
		"pic":            in.Pic,
		"album_pics":     util.ToJSON(in.AlbumPics),
		"pics34_json":    util.ToJSON(pics34),
		"detail_html":    in.DetailHTML,
		"sku_specs_json": skuSpecsJSON,
		"media_json":     mediaJSON,
	}); err != nil {
		return ok, failed, err
	}
	for _, sku := range in.Skus {
		code := strings.TrimSpace(sku.SkuCode)
		if code == "" {
			continue
		}
		if err := s.repo.UpdateSkuPic(id, code, sku.Pic); err != nil {
			return ok, failed, err
		}
	}
	if failed > 0 {
		return ok, failed, fmt.Errorf("%s", firstErr)
	}
	return ok, failed, nil
}

func resolveList(list []string, resource string, resolve func(string, string) string) []string {
	out := make([]string, len(list))
	for i, raw := range list {
		out[i] = resolve(raw, resource)
	}
	return out
}

func dropCoverDuplicate(cover string, list []string) []string {
	cover = strings.TrimSpace(cover)
	if cover == "" {
		return list
	}
	out := make([]string, 0, len(list))
	for _, raw := range list {
		if strings.TrimSpace(raw) == cover {
			continue
		}
		out = append(out, raw)
	}
	return out
}

func compactRemoteURLs(list []string) []string {
	out := make([]string, 0, len(list))
	for _, raw := range list {
		if strings.TrimSpace(raw) != "" {
			out = append(out, raw)
		}
	}
	return out
}

func rewriteCollectImgSrc(html string, resolve func(string) string) string {
	if strings.TrimSpace(html) == "" {
		return html
	}
	return reCollectImgSrc.ReplaceAllStringFunc(html, func(s string) string {
		m := reCollectImgSrc.FindStringSubmatch(s)
		if len(m) != 4 {
			return s
		}
		next := resolve(m[2])
		if next == "" {
			return ""
		}
		return m[1] + next + m[3]
	})
}
