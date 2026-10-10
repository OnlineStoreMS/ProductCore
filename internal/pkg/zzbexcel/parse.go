package zzbexcel

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"strings"
)

// Row 至尊宝「导出 excel」里的一行店铺商品。该文件实际是 HTML 表格。
type Row struct {
	Title          string
	ItemID         string
	ItemURL        string
	PicURL         string
	Price          string
	Sales          string
	CommentCount   string
	MonthDeals     string
	MonthConsign   string
	ShipTime       string
	ListedAt       string
	Category       string
	Tags           string
	SourceShopName string
	SourceShopURL  string
}

var (
	rowRE  = regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
	cellRE = regexp.MustCompile(`(?is)<t[dh][^>]*>(.*?)</t[dh]>`)
	tagRE  = regexp.MustCompile(`(?is)<[^>]+>`)
)

// Parse 解析至尊宝导出的 HTML 表格（扩展名常为 .xls / .xlsx）。
func Parse(data []byte) ([]Row, error) {
	text := strings.TrimSpace(string(bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})))
	if text == "" {
		return nil, fmt.Errorf("文件是空的")
	}
	if !strings.Contains(strings.ToLower(text), "<table") {
		return nil, fmt.Errorf("无法识别至尊宝导出表，请上传至尊宝导出的 Excel")
	}
	rawRows := rowRE.FindAllStringSubmatch(text, -1)
	if len(rawRows) < 2 {
		return nil, fmt.Errorf("表格里没有商品数据")
	}
	header := cells(rawRows[0][1])
	idx := map[string]int{}
	for i, name := range header {
		key := strings.TrimSpace(name)
		if key != "" {
			idx[key] = i
		}
	}
	titleAt, titleOK := firstIndex(idx, "标题")
	idAt, idOK := firstIndex(idx, "宝贝ID", "商品ID")
	if !titleOK || !idOK {
		return nil, fmt.Errorf("表头需包含「标题」和「宝贝ID」")
	}
	col := func(names ...string) int {
		at, ok := firstIndex(idx, names...)
		if !ok {
			return -1
		}
		return at
	}
	linkAt := col("宝贝链接", "商品链接")
	picAt := col("图片地址", "图片", "主图")
	priceAt := col("价格")
	salesAt := col("销量")
	commentAt := col("评论数")
	dealsAt := col("月成交笔数")
	consignAt := col("月代销")
	shipAt := col("发货时间")
	listedAt := col("上架时间")
	categoryAt := col("类目")
	tagsAt := col("标签")
	shopAt := col("店铺名")
	shopURLAt := col("店铺链接")

	out := make([]Row, 0, len(rawRows)-1)
	for _, raw := range rawRows[1:] {
		line := cells(raw[1])
		itemID := cellAt(line, idAt)
		title := cellAt(line, titleAt)
		if itemID == "" && title == "" {
			continue
		}
		if itemID == "" {
			continue
		}
		out = append(out, Row{
			Title:          title,
			ItemID:         itemID,
			ItemURL:        cellAt(line, linkAt),
			PicURL:         cellAt(line, picAt),
			Price:          cellAt(line, priceAt),
			Sales:          cellAt(line, salesAt),
			CommentCount:   cellAt(line, commentAt),
			MonthDeals:     cellAt(line, dealsAt),
			MonthConsign:   cellAt(line, consignAt),
			ShipTime:       cellAt(line, shipAt),
			ListedAt:       cellAt(line, listedAt),
			Category:       cellAt(line, categoryAt),
			Tags:           cellAt(line, tagsAt),
			SourceShopName: cellAt(line, shopAt),
			SourceShopURL:  cellAt(line, shopURLAt),
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("表格里没有可导入的商品")
	}
	return out, nil
}

func firstIndex(idx map[string]int, names ...string) (int, bool) {
	for _, name := range names {
		if at, ok := idx[name]; ok {
			return at, true
		}
	}
	return 0, false
}

func cells(rowHTML string) []string {
	found := cellRE.FindAllStringSubmatch(rowHTML, -1)
	out := make([]string, 0, len(found))
	for _, m := range found {
		text := tagRE.ReplaceAllString(m[1], "")
		text = html.UnescapeString(text)
		text = strings.TrimSpace(strings.ReplaceAll(text, "\u00a0", " "))
		out = append(out, text)
	}
	return out
}

func cellAt(line []string, at int) string {
	if at < 0 || at >= len(line) {
		return ""
	}
	return strings.TrimSpace(line[at])
}
