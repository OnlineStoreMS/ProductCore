-- 商品来源（淘宝、抖音、闲鱼、小红书等）
ALTER TABLE products ADD COLUMN IF NOT EXISTS source VARCHAR(64) DEFAULT '';

COMMENT ON COLUMN products.source IS '商品来源';
