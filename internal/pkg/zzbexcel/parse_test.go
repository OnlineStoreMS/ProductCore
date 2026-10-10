package zzbexcel

import (
	"os"
	"testing"
)

func TestParseHTMLTable(t *testing.T) {
	html := `<html><body><table><thead><tr><th>标题</th><th>宝贝ID</th><th>宝贝链接</th><th>图片地址</th><th>价格</th><th>销量</th><th>评论数</th><th>月成交笔数</th><th>月代销</th><th>发货时间</th><th>上架时间</th><th>类目</th><th>标签</th><th>店铺名</th><th>店铺链接</th></tr></thead><tbody><tr><td>链条</td><td>646794414807</td><td>https://item.taobao.com/item.htm?id=646794414807</td><td>https://gw.alicdn.com/a.jpg</td><td></td><td>5000</td><td></td><td></td><td></td><td></td><td></td><td></td><td></td><td>微笑单车</td><td>https://shop.taobao.com/</td></tr></tbody></table></body></html>`
	rows, err := Parse([]byte(html))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("len = %d", len(rows))
	}
	if rows[0].ItemID != "646794414807" || rows[0].Title != "链条" || rows[0].Sales != "5000" || rows[0].PicURL == "" {
		t.Fatalf("%+v", rows[0])
	}
}

func TestParseSampleFile(t *testing.T) {
	path := "../../../.tmp/至尊宝_导出excel (1).xlsx"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skip(err)
	}
	rows, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) < 50 {
		t.Fatalf("rows = %d", len(rows))
	}
	if rows[0].ItemID != "646794414807" || rows[0].Sales != "5000" {
		t.Fatalf("%+v", rows[0])
	}
}
