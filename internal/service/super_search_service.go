package service

import (
	"encoding/json"
	"sort"
	"strings"

	"productcore/internal/dto"
)

func (s *ProductService) SuperSearch(q dto.SuperSearchQuery) ([]dto.SuperSearchItemDTO, int64, error) {
	keyword := strings.TrimSpace(q.Keyword)
	if keyword == "" {
		return []dto.SuperSearchItemDTO{}, 0, nil
	}
	rows, total, err := s.repo.SearchSkus(keyword, q.Page, q.PageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return []dto.SuperSearchItemDTO{}, total, nil
	}

	productIDs := make([]uint64, 0, len(rows))
	seenProducts := make(map[uint64]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seenProducts[row.ProductID]; ok {
			continue
		}
		seenProducts[row.ProductID] = struct{}{}
		productIDs = append(productIDs, row.ProductID)
	}

	shopMap, err := s.meta.PlatformListing.ShopsByProductIDs(productIDs)
	if err != nil {
		return nil, 0, err
	}

	out := make([]dto.SuperSearchItemDTO, 0, len(rows))
	for _, row := range rows {
		specs := map[string]string{}
		_ = json.Unmarshal([]byte(row.SpecData), &specs)
		item := dto.SuperSearchItemDTO{
			ProductID:     row.ProductID,
			ProductName:   row.ProductName,
			MaterialCode:  row.MaterialCode,
			ProductSn:     row.ProductSn,
			ProductPic:    row.ProductPic,
			PublishStatus: row.PublishStatus,
			SkuID:         row.SkuID,
			SkuCode:       row.SkuCode,
			Specs:         specs,
			SpecLabel:     formatSpecsLabel(specs),
			Price:         row.Price,
			Stock:         row.Stock,
			Pic:           row.Pic,
		}
		if row.BrandID > 0 {
			item.BrandName = s.meta.Brand.GetName(row.BrandID)
		}
		if row.CategoryID > 0 {
			item.CategoryName = s.meta.Category.GetName(row.CategoryID)
		}
		shops := shopMap[row.ProductID]
		if shops == nil {
			shops = []dto.ListedShopDTO{}
		}
		item.ListedShops = shops
		item.ListedShopCount = len(shops)
		if s.store != nil {
			if item.Pic != "" {
				item.Pic = s.store.ResolvePublicURL(item.Pic)
			}
			if item.ProductPic != "" {
				item.ProductPic = s.store.ResolvePublicURL(item.ProductPic)
			}
			for i := range item.ListedShops {
				if item.ListedShops[i].PlatformTypeLogo != "" {
					item.ListedShops[i].PlatformTypeLogo = s.store.ResolvePublicURL(item.ListedShops[i].PlatformTypeLogo)
				}
			}
		}
		out = append(out, item)
	}
	return out, total, nil
}

func formatSpecsLabel(specs map[string]string) string {
	if len(specs) == 0 {
		return "-"
	}
	keys := make([]string, 0, len(specs))
	for k := range specs {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := strings.TrimSpace(specs[k])
		if v == "" {
			continue
		}
		parts = append(parts, k+": "+v)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " / ")
}
