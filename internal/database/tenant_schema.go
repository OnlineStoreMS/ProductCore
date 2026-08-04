package database

import "gorm.io/gorm"

func ensureTenantSchema(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	tables := []string{
		"brands", "categories", "product_groups", "product_keywords", "products", "product_skus",
		"platform_shops", "platform_listings",
	}
	for _, t := range tables {
		if err := db.Exec(`ALTER TABLE ` + t + ` ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 1`).Error; err != nil {
			return err
		}
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_brands_tenant ON brands(tenant_id)`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_products_tenant ON products(tenant_id)`).Error; err != nil {
		return err
	}
	if err := db.Exec(`DROP INDEX IF EXISTS idx_products_material_code`).Error; err != nil {
		return err
	}
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_products_tenant_material_code
		ON products (tenant_id, material_code)
		WHERE material_code IS NOT NULL AND material_code <> '';
	`).Error; err != nil {
		return err
	}
	if err := db.Exec(`DROP INDEX IF EXISTS uni_product_skus_sku_code`).Error; err != nil {
		return err
	}
	if err := db.Exec(`DROP INDEX IF EXISTS idx_product_skus_sku_code`).Error; err != nil {
		return err
	}
	return db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_product_skus_tenant_code
		ON product_skus (tenant_id, sku_code);
	`).Error
}
