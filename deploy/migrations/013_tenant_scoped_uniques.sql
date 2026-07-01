-- 资料编码、SKU 编码改为租户内唯一（多租户隔离）
DROP INDEX IF EXISTS idx_products_material_code;
DROP INDEX IF EXISTS uni_products_material_code;
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_tenant_material_code
  ON products (tenant_id, material_code)
  WHERE material_code IS NOT NULL AND material_code <> '';

DROP INDEX IF EXISTS uni_product_skus_sku_code;
DROP INDEX IF EXISTS idx_product_skus_sku_code;
CREATE UNIQUE INDEX IF NOT EXISTS idx_product_skus_tenant_code
  ON product_skus (tenant_id, sku_code);

COMMENT ON COLUMN products.material_code IS '来源资料编码，选填，非空时在租户内唯一，导入时用于匹配更新';
