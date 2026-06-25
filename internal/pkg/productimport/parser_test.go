package productimport_test

import (
	"testing"

	"productcore/internal/pkg/productimport"
)

func TestParseZipSample(t *testing.T) {
	zipPath := "../../../data/show/淘宝_商品ID_723068433145.zip"
	pkg, cleanup, err := productimport.ParseZip(zipPath)
	if err != nil {
		t.Fatalf("ParseZip: %v", err)
	}
	defer cleanup()

	if pkg.Source != "淘宝" {
		t.Fatalf("source = %q, want 淘宝", pkg.Source)
	}
	if pkg.MaterialCode != "723068433145" {
		t.Fatalf("materialCode = %q, want 723068433145", pkg.MaterialCode)
	}
	if len(pkg.MainPics) != 5 {
		t.Fatalf("main pics = %d, want 5", len(pkg.MainPics))
	}
	if len(pkg.Skus) != 22 {
		t.Fatalf("skus = %d, want 22", len(pkg.Skus))
	}
	if len(pkg.DetailPics) != 16 {
		t.Fatalf("detail pics = %d, want 16", len(pkg.DetailPics))
	}
	if len(pkg.Videos) != 1 {
		t.Fatalf("videos = %d, want 1", len(pkg.Videos))
	}
	if pkg.Skus[0].SpecName != "商品规格" {
		t.Fatalf("spec name = %q", pkg.Skus[0].SpecName)
	}
	if pkg.Skus[0].Code != "SKU01" {
		t.Fatalf("first sku code = %q", pkg.Skus[0].Code)
	}
}
