-- 商品草稿箱 + 资料编码改为选填（非空时唯一）
ALTER TABLE products ADD COLUMN IF NOT EXISTS is_draft SMALLINT NOT NULL DEFAULT 0;

DROP INDEX IF EXISTS idx_products_material_code;
DROP INDEX IF EXISTS uni_products_material_code;
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_material_code
  ON products (material_code)
  WHERE material_code IS NOT NULL AND material_code <> '';

COMMENT ON COLUMN products.is_draft IS '1=草稿箱，保存后变为0';
COMMENT ON COLUMN products.material_code IS '来源资料编码，选填，非空时唯一，导入时用于匹配更新';

-- 历史自动生成的草稿编码归入草稿箱并清空资料编码
UPDATE products
SET is_draft = 1, material_code = ''
WHERE deleted_at IS NULL
  AND material_code LIKE 'DRAFT\_%' ESCAPE '\';
