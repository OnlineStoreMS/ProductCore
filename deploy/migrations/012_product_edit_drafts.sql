-- 已发布商品的编辑草稿（与草稿箱 is_draft 分离，不影响商品列表展示）
CREATE TABLE IF NOT EXISTS product_edit_drafts (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL UNIQUE REFERENCES products(id) ON DELETE CASCADE,
    payload_json TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_edit_drafts_product_id ON product_edit_drafts(product_id);

COMMENT ON TABLE product_edit_drafts IS '已发布商品编辑中的临时草稿，finalize 后合并入主表并删除';
COMMENT ON COLUMN product_edit_drafts.payload_json IS 'ProductDTO JSON（含 SKU/规格/媒体等）';
