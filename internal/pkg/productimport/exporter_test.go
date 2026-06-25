package productimport_test

import (
	"strings"
	"testing"

	"productcore/internal/dto"
	"productcore/internal/pkg/productimport"
)

func TestBuildExportFolderName(t *testing.T) {
	got := productimport.BuildExportFolderName(123)
	want := "商品中心_商品ID_123"
	if got != want {
		t.Fatalf("BuildExportFolderName = %q, want %q", got, want)
	}
}

func TestBuildSkuFileName(t *testing.T) {
	got := productimport.BuildSkuFileName(1, "商品规格", "红色", ".jpg")
	want := "SKU01_商品规格(红色).jpg"
	if got != want {
		t.Fatalf("BuildSkuFileName = %q, want %q", got, want)
	}
}

func TestBuildSkuDetailText(t *testing.T) {
	product := &dto.ProductDTO{
		ID:           123,
		Name:         "测试商品",
		MaterialCode: "MC001",
		Source:       "淘宝",
		BrandName:    "品牌A",
		CategoryName: "分类B",
		SkuSpecs:     []dto.SkuSpecDTO{{Name: "颜色", Values: []dto.SkuSpecValueDTO{{Value: "红"}}}},
		Skus: []dto.SkuDTO{
			{SkuCode: "MC001SKU01", Specs: map[string]string{"颜色": "红"}, Price: 99, Stock: 10, Weight: 0.5},
		},
	}
	text := productimport.BuildSkuDetailText(product, "颜色")
	if !strings.Contains(text, "MC001SKU01") {
		t.Fatalf("missing sku code in detail text")
	}
	if !strings.Contains(text, "SKU 明细") {
		t.Fatalf("missing header in detail text")
	}
}
