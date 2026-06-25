-- 货号改为可选；资料编码唯一
DROP INDEX IF EXISTS idx_products_product_sn;
DROP INDEX IF EXISTS uni_products_product_sn;

CREATE UNIQUE INDEX IF NOT EXISTS idx_products_material_code ON products (material_code);

COMMENT ON COLUMN products.material_code IS '资料编码，商品唯一标识，导入时用于匹配更新';
COMMENT ON COLUMN products.product_sn IS '货号，商家商品编号（可选）';
