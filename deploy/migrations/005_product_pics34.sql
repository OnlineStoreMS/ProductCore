-- 3:4 主图独立存储（JSON 字符串数组，最多 5 张 URL）
ALTER TABLE products ADD COLUMN IF NOT EXISTS pics34_json TEXT DEFAULT '';

COMMENT ON COLUMN products.pics34_json IS '3:4 主图 URL 列表 JSON 数组，与 media_json.pics34 解耦后以此为准';
