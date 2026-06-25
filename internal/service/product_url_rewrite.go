package service

import (
	"regexp"

	"productcore/internal/dto"
)

var htmlSrcAttrRe = regexp.MustCompile(`(?i)(src\s*=\s*["'])([^"']+)(["'])`)

func rewriteProductURLs(out *dto.ProductDTO, resolve func(string) string) {
	if out == nil || resolve == nil {
		return
	}
	out.Pic = resolve(out.Pic)
	out.ProductVideo = resolve(out.ProductVideo)
	for i, u := range out.AlbumPics {
		out.AlbumPics[i] = resolve(u)
	}
	if out.Media != nil {
		rewriteMediaURLs(out.Media, resolve)
	}
	for i := range out.SkuSpecs {
		for j := range out.SkuSpecs[i].Values {
			out.SkuSpecs[i].Values[j].Pic = resolve(out.SkuSpecs[i].Values[j].Pic)
		}
	}
	for i := range out.Skus {
		out.Skus[i].Pic = resolve(out.Skus[i].Pic)
	}
	out.DetailHTML = rewriteHTMLAssetURLs(out.DetailHTML, resolve)
}

func rewriteMediaURLs(media *dto.ProductMediaDTO, resolve func(string) string) {
	for i, u := range media.Pics34 {
		media.Pics34[i] = resolve(u)
	}
	for i, u := range media.DetailPics {
		media.DetailPics[i] = resolve(u)
	}
	if media.Videos != nil {
		media.Videos.Ratio11 = resolve(media.Videos.Ratio11)
		media.Videos.Ratio34 = resolve(media.Videos.Ratio34)
		media.Videos.Ratio169 = resolve(media.Videos.Ratio169)
		media.Videos.Ratio916 = resolve(media.Videos.Ratio916)
	}
	if media.Materials != nil {
		media.Materials.White = resolve(media.Materials.White)
		media.Materials.Transparent = resolve(media.Materials.Transparent)
		media.Materials.Guide34 = resolve(media.Materials.Guide34)
		media.Materials.Long = resolve(media.Materials.Long)
	}
}

func rewriteHTMLAssetURLs(html string, resolve func(string) string) string {
	if html == "" {
		return html
	}
	return htmlSrcAttrRe.ReplaceAllStringFunc(html, func(match string) string {
		parts := htmlSrcAttrRe.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}
		resolved := resolve(parts[2])
		if resolved == parts[2] {
			return match
		}
		return parts[1] + resolved + parts[3]
	})
}
