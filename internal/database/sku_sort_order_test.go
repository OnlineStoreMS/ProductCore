package database_test

import (
	"testing"

	"productcore/internal/config"
	"productcore/internal/database"
	"productcore/internal/model"
)

func TestListSkusOrderedBySortOrder(t *testing.T) {
	db, err := database.Connect(&config.DatabaseConfig{
		Driver:     "sqlite",
		SQLitePath: t.TempDir() + "/test.db",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}

	p := &model.Product{Name: "p", Unit: "件", VerifyStatus: 1, ChannelVisible: "both"}
	if err := db.Create(p).Error; err != nil {
		t.Fatal(err)
	}

	skus := []model.Sku{
		{ProductID: p.ID, SkuCode: "A001", SortOrder: 2, SpecData: `{}`},
		{ProductID: p.ID, SkuCode: "A002", SortOrder: 0, SpecData: `{}`},
		{ProductID: p.ID, SkuCode: "A003", SortOrder: 1, SpecData: `{}`},
	}
	for i := range skus {
		if err := db.Create(&skus[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	var got []model.Sku
	if err := db.Where("product_id = ?", p.ID).Order("sort_order ASC, id ASC").Find(&got).Error; err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d", len(got))
	}
	wantCodes := []string{"A002", "A003", "A001"}
	for i, code := range wantCodes {
		if got[i].SkuCode != code {
			t.Fatalf("index %d = %q, want %q", i, got[i].SkuCode, code)
		}
	}
}
