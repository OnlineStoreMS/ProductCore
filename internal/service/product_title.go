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
长度必须控制在 60 个字符以内，并且不少于 16 个字符。一个汉字或中文符号算 2 个字符，一个英文字母、数字或英文符号算 1 个字符。中文标题大约不超过 30 个字。
写成一句完整标题，不要超长，不要在词中间停住。
不要出现禧玛诺，也不要出现 Shimano 的任何大小写，例如 SHIMANO、shimano、Shimano。把这个品牌换成商品实际对应的车型：公路车用「公路自行车」，山地车用「山地自行车」，折叠车用「折叠自行车」，通勤或城市车用「通勤自行车」，儿童车用「儿童自行车」。资料里同时有多种车，或看不出具体车型时，用「自行车」，不要猜成某一种。
保留型号、速别和关键规格。不要编造资料里没有的功能、认证、赠品或型号。
不要出现「副本」「包邮」「最」「第一」「顶级」「国家级」这类词。`

func listingTitleBackground(product *dto.ProductDTO, categoryPath []string) string {
	var b strings.Builder
	b.WriteString("请为下面的商品写一个不超过 60 个字符的抖店标题。汉字算 2 个字符。\n")
	b.WriteString(listingSceneLine(product, categoryPath))
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
	raw = stripListingBrand(raw)
	return trimDouyinTitle(strings.TrimSpace(raw), 60)
}

var (
	shimanoLatinRe = regexp.MustCompile(`(?i)shimano`)
	shimanoHanRe   = regexp.MustCompile(`禧玛诺|禧马诺|喜玛诺|喜马诺`)
	roadSceneRe    = regexp.MustCompile(`公路|(?i)\broad\b`)
	mtbSceneRe     = regexp.MustCompile(`山地|越野|(?i)\bmtb\b`)
)

func listingSceneLine(product *dto.ProductDTO, categoryPath []string) string {
	scenes := listingScenes(product, categoryPath)
	if len(scenes) == 1 && scenes[0] == "自行车" {
		return "品牌替换：用「自行车」。资料里没有明确车型，不要写成公路自行车或山地自行车。\n"
	}
	if len(scenes) > 1 {
		return "品牌替换：资料同时涉及" + strings.Join(scenes, "、") + "，用「自行车」概括，不要保留禧玛诺或 Shimano。\n"
	}
	return "品牌替换：用「" + scenes[0] + "」替换禧玛诺和 Shimano。\n"
}

func listingScenes(product *dto.ProductDTO, categoryPath []string) []string {
	var bits []string
	bits = append(bits, product.Name, product.SubTitle, strings.Join(categoryPath, " "))
	for _, spec := range product.SkuSpecs {
		bits = append(bits, spec.Name)
		for _, v := range spec.Values {
			bits = append(bits, v.Value)
		}
	}
	scenes := detectRidingScenes(strings.Join(bits, " "))
	if len(scenes) == 0 {
		scenes = detectRidingScenes(plainText(product.Description + " " + product.DetailHTML))
	}
	if len(scenes) == 0 {
		return []string{"自行车"}
	}
	return scenes
}

func detectRidingScenes(text string) []string {
	var scenes []string
	if roadSceneRe.MatchString(text) {
		scenes = append(scenes, "公路自行车")
	}
	if mtbSceneRe.MatchString(text) {
		scenes = append(scenes, "山地自行车")
	}
	if strings.Contains(text, "折叠") {
		scenes = append(scenes, "折叠自行车")
	}
	if strings.Contains(text, "通勤") || strings.Contains(text, "城市车") || strings.Contains(text, "城市自行车") {
		scenes = append(scenes, "通勤自行车")
	}
	if strings.Contains(text, "儿童") || strings.Contains(text, "童车") || strings.Contains(text, "小孩") {
		scenes = append(scenes, "儿童自行车")
	}
	return scenes
}

func stripListingBrand(s string) string {
	s = shimanoLatinRe.ReplaceAllString(s, "")
	s = shimanoHanRe.ReplaceAllString(s, "")
	return strings.Join(strings.Fields(s), " ")
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
