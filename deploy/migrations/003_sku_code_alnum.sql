-- 规格编码 sku_code：仅字母数字，VARCHAR(64)
UPDATE product_skus
SET sku_code = regexp_replace(sku_code, '[^A-Za-z0-9]', '', 'g')
WHERE sku_code ~ '[^A-Za-z0-9]';

ALTER TABLE product_skus
  ALTER COLUMN sku_code TYPE VARCHAR(64);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_product_skus_sku_code_alnum'
  ) THEN
    ALTER TABLE product_skus
      ADD CONSTRAINT chk_product_skus_sku_code_alnum
      CHECK (sku_code ~ '^[A-Za-z0-9]+$' AND char_length(sku_code) > 0);
  END IF;
END $$;

COMMENT ON COLUMN product_skus.sku_code IS 'SKU 规格编码，全局唯一，仅字母与数字';
