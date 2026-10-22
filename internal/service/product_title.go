package service

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"

	"productcore/internal/dto"
	"productcore/internal/integrations/aimodel"
)

const listingTitleModel = "deepseek-flash"

var htmlTagRe = regexp.MustCompile(`(?s)<[^>]*>`)

func (s *ProductPublishService) SuggestListingTitle(ctx context.Context, productID uint64) (string, error) {
	product, err := s.productSvc.Get(productID)
	if err != nil {
		return "", err
	}
	return s.optimizeListingTitle(ctx, product, s.categoryPath(product.CategoryID))
}

func (s *ProductPublishService) optimizeListingTitle(ctx context.Context, product *dto.ProductDTO, categoryPath []string) (string, error) {
	if s.ai == nil {
		return "", fmt.Errorf("未配置 DeepSeek，无法优化标题")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	raw, err := s.ai.Chat(ctx, aimodel.ChatInput{
		TenantID:  s.tenantID,
		Model:     listingTitleModel,
		System:    listingTitleSystem,
		User:      listingTitleBackground(product, categoryPath),
		MaxTokens: 200,
	})
	if err != nil {
		return "", err
	}
	title := cleanGeneratedTitle(raw)
	if douyinTitleWeight(title) < 16 {
		return "", fmt.Errorf("生成的标题太短，请重试")
	}
	return title, nil
}

const listingTitleSystem = `你是抖店电商标题优化助手。根据商品资料写一个新的商品标题。
只输出一行标题，不要解释，不要引号，不要序号。
计字规则：一个汉字或中文符号算 2，一个英文字母、数字或英文符号算 1，总长度必须在 16 到 60 之间。
保留资料里的品牌、品类和关键型号。可以重组卖点，不要编造资料里没有的功能、认证、赠品或型号。
不要出现「副本」「包邮」「最」「第一」「顶级」「国家级」这类词。`

func listingTitleBackground(product *dto.ProductDTO, categoryPath []string) string {
	var b strings.Builder
	b.WriteString("请为下面的商品写一个抖店标题。\n")
	writeLine(&b, "现有标题", product.Name)
	writeLine(&b, "副标题", product.SubTitle)
	writeLine(&b, "品牌", product.BrandName)
	writeLine(&b, "分类", strings.Join(categoryPath, " > "))
	writeLine(&b, "货号", product.ProductSn)
	writeLine(&b, "资料编码", product.MaterialCode)
	writeLine(&b, "来源", product.Source)
	writeLine(&b, "单位", product.Unit)
	if product.Price > 0 {
		fmt.Fprintf(&b, "售价：%.0f\n", product.Price)
	}
	desc := plainText(product.Description)
	if desc == "" {
		desc = plainText(product.DetailHTML)
	}
	writeLine(&b, "商品描述", clipRunes(desc, 2500))
	if len(product.SkuSpecs) > 0 {
		b.WriteString("规格：\n")
		for _, spec := range product.SkuSpecs {
			values := make([]string, 0, len(spec.Values))
			for _, v := range spec.Values {
				if name := strings.TrimSpace(v.Value); name != "" {
					values = append(values, name)
				}
			}
			if len(values) > 80 {
				values = append(values[:80], "…")
			}
			fmt.Fprintf(&b, "- %s：%s\n", strings.TrimSpace(spec.Name), strings.Join(values, "、"))
		}
	}
	if len(product.Skus) > 0 {
		b.WriteString("SKU：\n")
		specName := "商品规格"
		if len(product.SkuSpecs) > 0 && strings.TrimSpace(product.SkuSpecs[0].Name) != "" {
			specName = strings.TrimSpace(product.SkuSpecs[0].Name)
		}
		limit := len(product.Skus)
		if limit > 80 {
			limit = 80
		}
		for _, sku := range product.Skus[:limit] {
			fmt.Fprintf(&b, "- %s 售价 %.0f 库存 %d\n", skuLabel(sku, specName), sku.Price, sku.Stock)
		}
		if len(product.Skus) > limit {
			fmt.Fprintf(&b, "- 其余 %d 个规格略\n", len(product.Skus)-limit)
		}
	}
	return b.String()
}

func writeLine(b *strings.Builder, label, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	fmt.Fprintf(b, "%s：%s\n", label, value)
}

func plainText(raw string) string {
	raw = htmlTagRe.ReplaceAllString(raw, " ")
	raw = html.UnescapeString(raw)
	return strings.Join(strings.Fields(raw), " ")
}

func clipRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	n := 0
	for i := range s {
		if n == max {
			return s[:i] + "…"
		}
		n++
	}
	return s
}

func cleanGeneratedTitle(raw string) string {
	raw = strings.TrimSpace(raw)
	if i := strings.IndexAny(raw, "\r\n"); i >= 0 {
		raw = raw[:i]
	}
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "\"'“”‘’「」『』")
	raw = strings.TrimSpace(raw)
	for _, prefix := range []string{"标题：", "标题:", "新标题：", "新标题:"} {
		raw = strings.TrimPrefix(raw, prefix)
	}
	raw = strings.ReplaceAll(raw, "副本", "")
	return trimDouyinTitle(strings.TrimSpace(raw), 60)
}

func douyinTitleWeight(s string) int {
	n := 0
	for _, r := range s {
		if r > 127 {
			n += 2
		} else {
			n++
		}
	}
	return n
}
