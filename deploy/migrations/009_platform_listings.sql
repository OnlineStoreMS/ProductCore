-- 商品铺货记录（SPU 级，手动绑定店铺）

CREATE TABLE IF NOT EXISTS platform_listings (
  id BIGSERIAL PRIMARY KEY,
  product_id BIGINT NOT NULL,
  platform_shop_id BIGINT NOT NULL,
  listing_status SMALLINT NOT NULL DEFAULT 1,
  remark VARCHAR(256) DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  UNIQUE (product_id, platform_shop_id)
);

CREATE INDEX IF NOT EXISTS idx_platform_listings_product ON platform_listings (product_id);
CREATE INDEX IF NOT EXISTS idx_platform_listings_shop ON platform_listings (platform_shop_id);
