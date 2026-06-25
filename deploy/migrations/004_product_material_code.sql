-- 资料编码（与货号 product_sn 区分）
ALTER TABLE products ADD COLUMN IF NOT EXISTS material_code VARCHAR(64) DEFAULT '';

COMMENT ON COLUMN products.material_code IS '资料编码，用于内部资料管理与规格编码前缀';
COMMENT ON COLUMN products.product_sn IS '货号，商家商品编号';
