-- ProductCore: 商品编辑页反向设计 — 媒体与 SKU 扩展字段
-- 适用于已有 PostgreSQL 库；新库由 GORM AutoMigrate 自动建列

ALTER TABLE products ADD COLUMN IF NOT EXISTS media_json TEXT DEFAULT '';

ALTER TABLE product_skus ADD COLUMN IF NOT EXISTS market_price DECIMAL(10,2) NOT NULL DEFAULT 0;
ALTER TABLE product_skus ADD COLUMN IF NOT EXISTS weight DECIMAL(10,2) NOT NULL DEFAULT 0;

COMMENT ON COLUMN products.media_json IS 'SPU 扩展媒体 JSON: pics34/videos/materials/detailPics';
COMMENT ON COLUMN product_skus.market_price IS 'SKU 市场价（划线价）';
COMMENT ON COLUMN product_skus.weight IS 'SKU 重量，单位克(g)';
