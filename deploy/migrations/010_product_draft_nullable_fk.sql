-- 草稿商品创建时尚未选择品牌/分类，允许 NULL（外键仅校验非空值）
ALTER TABLE products ALTER COLUMN brand_id DROP NOT NULL;
ALTER TABLE products ALTER COLUMN category_id DROP NOT NULL;

UPDATE products SET brand_id = NULL WHERE brand_id = 0;
UPDATE products SET category_id = NULL WHERE category_id = 0;
