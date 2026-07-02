package productimport

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildSkuPlaceholderFileName(t *testing.T) {
	got := BuildSkuPlaceholderFileName(7, "规格", "红/蓝")
	want := "SKU07_规格(红_蓝)"
	if got != want {
		t.Fatalf("BuildSkuPlaceholderFileName = %q, want %q", got, want)
	}
	if filepath.Ext(got) != "" {
		t.Fatalf("placeholder should have no extension, got %q", filepath.Ext(got))
	}
}

func TestParseSkuSpecPartsNestedParens(t *testing.T) {
	specName, specValue, err := parseSkuSpecParts("商品规格(Sigeyi Axo功率计贴纸可指定颜色--(5选1))")
	if err != nil {
		t.Fatalf("parseSkuSpecParts: %v", err)
	}
	if specName != "商品规格" {
		t.Fatalf("specName = %q, want 商品规格", specName)
	}
	if specValue != "Sigeyi Axo功率计贴纸可指定颜色--(5选1)" {
		t.Fatalf("specValue = %q", specValue)
	}
}

func TestParseSkuDirSkipsExtensionlessPlaceholder(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SKU01_规格(红色)"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKU02_规格(蓝色).jpg"), []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}

	skus, err := parseSkuDir(dir)
	if err != nil {
		t.Fatalf("parseSkuDir: %v", err)
	}
	if len(skus) != 1 {
		t.Fatalf("skus = %d, want 1", len(skus))
	}
	if skus[0].Index != 2 || skus[0].SpecValue != "蓝色" {
		t.Fatalf("unexpected sku: %+v", skus[0])
	}
}

func TestParseSkuDirSkipsNonImageExtension(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SKU01_规格(红色).txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKU02_规格(蓝色).jpg"), []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}

	skus, err := parseSkuDir(dir)
	if err != nil {
		t.Fatalf("parseSkuDir: %v", err)
	}
	if len(skus) != 1 || skus[0].Index != 2 {
		t.Fatalf("unexpected skus: %+v", skus)
	}
}
