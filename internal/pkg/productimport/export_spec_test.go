package productimport

import (
	"testing"

	"productcore/internal/dto"
)

func TestExportSpecNameUsesSkuSpecs(t *testing.T) {
	specs := []dto.SkuSpecDTO{
		{Name: "", Values: nil},
		{Name: "规格", Values: []dto.SkuSpecValueDTO{{Value: "红色"}}},
	}
	skus := []dto.SkuDTO{
		{Specs: map[string]string{"商品规格": "红色"}},
	}
	if got := exportSpecName(specs, skus); got != "规格" {
		t.Fatalf("exportSpecName = %q, want 规格", got)
	}
}

func TestExportSpecNamePrefersSkuSpecsOverSpecsKeys(t *testing.T) {
	specs := []dto.SkuSpecDTO{{Name: "规格", Values: []dto.SkuSpecValueDTO{{Value: "红色"}}}}
	skus := []dto.SkuDTO{
		{Specs: map[string]string{"商品规格": "红色"}},
	}
	if got := exportSpecName(specs, skus); got != "规格" {
		t.Fatalf("exportSpecName = %q, want 规格", got)
	}
}

func TestExportSpecValueKeyMismatch(t *testing.T) {
	specs := []dto.SkuSpecDTO{{Name: "规格", Values: []dto.SkuSpecValueDTO{{Value: "【国产】RT70-140mm"}}}}
	sku := dto.SkuDTO{Specs: map[string]string{"商品规格": "【国产】RT70-140mm"}}
	if got := exportSpecValue(specs, sku); got != "【国产】RT70-140mm" {
		t.Fatalf("exportSpecValue = %q, want spec value", got)
	}
}

func TestBuildSkuFileNameUsesActualSpecName(t *testing.T) {
	specs := []dto.SkuSpecDTO{{Name: "规格", Values: []dto.SkuSpecValueDTO{{Value: "红色"}}}}
	skus := []dto.SkuDTO{{Specs: map[string]string{"商品规格": "红色"}}}
	specName := exportSpecName(specs, skus)
	specValue := exportSpecValue(specs, skus[0])
	got := BuildSkuFileName(1, specName, specValue, ".jpg")
	want := "SKU01_规格(红色).jpg"
	if got != want {
		t.Fatalf("BuildSkuFileName = %q, want %q", got, want)
	}
}

func TestBuildSkuFileNameSlashToUnderscore(t *testing.T) {
	got := BuildSkuFileName(1, "规格", "红/蓝", ".jpg")
	want := "SKU01_规格(红_蓝).jpg"
	if got != want {
		t.Fatalf("BuildSkuFileName = %q, want %q", got, want)
	}
}
