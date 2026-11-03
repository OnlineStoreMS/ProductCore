package service

import (
	"testing"

	"productcore/internal/dto"
	"productcore/internal/integrations/agentscenter"
)

func testShop() *agentscenter.ShopItem {
	return &agentscenter.ShopItem{
		Platform:         "doudian",
		PlatformShopID:   "s1",
		PlatformShopName: "测试店",
	}
}

func TestBuildPublishParamsSingleSpecUsesProductSpec(t *testing.T) {
	p := &dto.ProductDTO{
		ID:   1,
		Name: "单规格商品",
		Pic:  "http://x/a.jpg",
		SkuSpecs: []dto.SkuSpecDTO{
			{Name: "颜色", Values: []dto.SkuSpecValueDTO{{Value: "红"}, {Value: "蓝"}}},
		},
		Skus: []dto.SkuDTO{
			{SkuCode: "A", Specs: map[string]string{"颜色": "红"}, Price: 10, Stock: 1},
			{SkuCode: "B", Specs: map[string]string{"颜色": "蓝"}, Price: 12, Stock: 2},
		},
	}
	got, err := buildPublishParams(p, testShop(), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.SpecName != "商品规格" {
		t.Fatalf("SpecName = %q, want 商品规格", got.SpecName)
	}
	if len(got.SpecTypes) != 1 || got.SpecTypes[0].Name != "商品规格" {
		t.Fatalf("SpecTypes = %+v, want one 商品规格", got.SpecTypes)
	}
	if len(got.Skus) != 2 {
		t.Fatalf("skus = %d, want 2", len(got.Skus))
	}
	if got.Skus[0].Name != "红" {
		t.Fatalf("sku0 name = %q, want 红", got.Skus[0].Name)
	}
}

func TestBuildPublishParamsMultiSpecUsesActualNames(t *testing.T) {
	p := &dto.ProductDTO{
		ID:   141,
		Name: "多规格商品",
		Pic:  "http://x/a.jpg",
		SkuSpecs: []dto.SkuSpecDTO{
			{Name: "颜色", Values: []dto.SkuSpecValueDTO{{Value: "黑"}, {Value: "白"}}},
			{Name: "尺寸", Values: []dto.SkuSpecValueDTO{{Value: "S"}, {Value: "M"}}},
		},
		Skus: []dto.SkuDTO{
			{SkuCode: "A", Specs: map[string]string{"颜色": "黑", "尺寸": "S"}, Price: 10, Stock: 1, Pic: "http://x/black.jpg"},
			{SkuCode: "B", Specs: map[string]string{"颜色": "黑", "尺寸": "M"}, Price: 11, Stock: 2, Pic: "http://x/black.jpg"},
			{SkuCode: "C", Specs: map[string]string{"颜色": "白", "尺寸": "S"}, Price: 12, Stock: 3, Pic: "http://x/white.jpg"},
			{SkuCode: "D", Specs: map[string]string{"颜色": "白", "尺寸": "M"}, Price: 13, Stock: 4, Pic: "http://x/white.jpg"},
		},
	}
	got, err := buildPublishParams(p, testShop(), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.SpecTypes) != 2 {
		t.Fatalf("SpecTypes len = %d, want 2", len(got.SpecTypes))
	}
	if got.SpecTypes[0].Name != "颜色分类" || got.SpecTypes[1].Name != "尺寸" {
		t.Fatalf("SpecTypes names = %q/%q, want 颜色分类/尺寸", got.SpecTypes[0].Name, got.SpecTypes[1].Name)
	}
	if got.SpecName != "颜色分类" {
		t.Fatalf("SpecName = %q, want 颜色分类", got.SpecName)
	}
	if len(got.Skus) != 4 {
		t.Fatalf("skus = %d, want 4 (must not collapse by color)", len(got.Skus))
	}
	sku0 := got.Skus[0]
	if sku0.Specs["颜色分类"] != "黑" || sku0.Specs["尺寸"] != "S" {
		t.Fatalf("sku0 specs = %+v", sku0.Specs)
	}
	if sku0.Name != "黑 S" {
		t.Fatalf("sku0 name = %q, want 黑 S", sku0.Name)
	}
}
