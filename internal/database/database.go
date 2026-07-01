package database

import (
	"fmt"
	"os"
	"path/filepath"

	"productcore/internal/config"
	"productcore/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case "postgres":
		dialector = postgres.Open(cfg.PostgresDSN)
	case "sqlite":
		if err := os.MkdirAll(filepath.Dir(cfg.SQLitePath), 0o755); err != nil {
			return nil, err
		}
		dialector = sqlite.Open(cfg.SQLitePath)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s (supported: postgres, sqlite)", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.Brand{},
		&model.Category{},
		&model.ProductGroup{},
		&model.Product{},
		&model.Sku{},
		&model.ProductGroupRelation{},
		&model.PlatformShopType{},
		&model.PlatformShop{},
		&model.PlatformListing{},
		&model.PlatformSkuMapping{},
		&model.ProductEditDraft{},
	); err != nil {
		return err
	}
	if err := backfillSkuSortOrder(db); err != nil {
		return err
	}
	if err := ensureTenantSchema(db); err != nil {
		return err
	}
	return ensureProductSchema(db)
}

func ensureProductSchema(db *gorm.DB) error {
	if err := ensureProductDraftColumns(db); err != nil {
		return err
	}
	switch db.Dialector.Name() {
	case "postgres":
		if err := db.Exec(`ALTER TABLE products ADD COLUMN IF NOT EXISTS is_draft SMALLINT NOT NULL DEFAULT 0`).Error; err != nil {
			return err
		}
		// 资料编码唯一索引由 ensureTenantSchema 按 (tenant_id, material_code) 维护
		return nil
	default:
		return nil
	}
}

func ensureProductDraftColumns(db *gorm.DB) error {
	switch db.Dialector.Name() {
	case "postgres":
		return db.Exec(`
			ALTER TABLE products ALTER COLUMN brand_id DROP NOT NULL;
			ALTER TABLE products ALTER COLUMN category_id DROP NOT NULL;
		`).Error
	default:
		return nil
	}
}
