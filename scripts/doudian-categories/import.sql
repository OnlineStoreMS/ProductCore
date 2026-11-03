-- 抖店铺货类目导入（整车及配件 / 骑行装备 / 维修工具）
-- 在目标 OSMS 的 PostgreSQL 上执行，可重复运行：同租户、同父级、同名则跳过。默认租户 1。
--
--   psql -h 127.0.0.1 -U productcore -d productcore -f scripts/doudian-categories/import.sql

BEGIN;

DO $$
DECLARE
  v_tenant BIGINT := 1;
BEGIN

  CREATE TEMP TABLE IF NOT EXISTS tmp_wa_cat (
    parent_path TEXT NOT NULL,
    name TEXT NOT NULL,
    level INT NOT NULL,
    sort INT NOT NULL
  ) ON COMMIT DROP;

  INSERT INTO tmp_wa_cat(parent_path, name, level, sort) VALUES
    ('', '整车及配件', 0, 200),
    ('整车及配件', '自行车及配件', 1, 200),
    ('整车及配件/自行车及配件', '变速器/保护器', 2, 230),
    ('整车及配件/自行车及配件', '车圈/刀圈', 2, 220),
    ('整车及配件/自行车及配件', '脚撑/停车架/骑行台', 2, 210),
    ('整车及配件/自行车及配件', '脚踏/轴承/滚珠', 2, 200),
    ('整车及配件/自行车及配件', '泥除/挡泥板', 2, 190),
    ('整车及配件/自行车及配件', '前叉/避震装置', 2, 180),
    ('整车及配件/自行车及配件', '曲柄', 2, 170),
    ('整车及配件/自行车及配件', '牙盘', 2, 160),
    ('整车及配件/自行车及配件', '中轴', 2, 150),
    ('整车及配件/自行车及配件', '自行车把立', 2, 140),
    ('整车及配件/自行车及配件', '自行车把套', 2, 130),
    ('整车及配件/自行车及配件', '自行车车把', 2, 120),
    ('整车及配件/自行车及配件', '自行车车架', 2, 110),
    ('整车及配件/自行车及配件', '自行车车篮/车筐', 2, 100),
    ('整车及配件/自行车及配件', '自行车链条/链条扣', 2, 90),
    ('整车及配件/自行车及配件', '自行车轮组', 2, 80),
    ('整车及配件/自行车及配件', '自行车内胎', 2, 70),
    ('整车及配件/自行车及配件', '自行车刹车装置', 2, 60),
    ('整车及配件/自行车及配件', '自行车外胎', 2, 50),
    ('整车及配件/自行车及配件', '自行车整车', 2, 40),
    ('整车及配件/自行车及配件', '自行车坐垫/鞍座', 2, 30),
    ('整车及配件/自行车及配件', '座管/座杆', 2, 20),
    ('整车及配件/自行车及配件', '其他自行车配件', 2, 10),
    ('', '摩托车/电动车/自行车用品', 0, 190),
    ('摩托车/电动车/自行车用品', '骑行装备', 1, 200),
    ('摩托车/电动车/自行车用品', '维修工具', 1, 190),
    ('摩托车/电动车/自行车用品/骑行装备', '车铃/喇叭', 2, 430),
    ('摩托车/电动车/自行车用品/骑行装备', '撑伞架', 2, 420),
    ('摩托车/电动车/自行车用品/骑行装备', '车尾箱', 2, 410),
    ('摩托车/电动车/自行车用品/骑行装备', '挡风被', 2, 400),
    ('摩托车/电动车/自行车用品/骑行装备', '打气筒', 2, 390),
    ('摩托车/电动车/自行车用品/骑行装备', '灯架/灯夹', 2, 380),
    ('摩托车/电动车/自行车用品/骑行装备', '电池防盗锁', 2, 370),
    ('摩托车/电动车/自行车用品/骑行装备', '电动车/摩托车车灯', 2, 360),
    ('摩托车/电动车/自行车用品/骑行装备', '电动车/摩托车贴纸', 2, 350),
    ('摩托车/电动车/自行车用品/骑行装备', '电动车/摩托车头盔', 2, 340),
    ('摩托车/电动车/自行车用品/骑行装备', '电动车/摩托车头盔耳机', 2, 330),
    ('摩托车/电动车/自行车用品/骑行装备', '电动车/摩托车音响', 2, 320),
    ('摩托车/电动车/自行车用品/骑行装备', '电动车/摩托车雨披', 2, 310),
    ('摩托车/电动车/自行车用品/骑行装备', '电动车/摩托车遮阳伞', 2, 300),
    ('摩托车/电动车/自行车用品/骑行装备', '电动车/摩托车坐垫', 2, 290),
    ('摩托车/电动车/自行车用品/骑行装备', '电动车/摩托车坐垫套', 2, 280),
    ('摩托车/电动车/自行车用品/骑行装备', '电动车/自行车车锁', 2, 270),
    ('摩托车/电动车/自行车用品/骑行装备', '电动车充电插座', 2, 260),
    ('摩托车/电动车/自行车用品/骑行装备', '电动车充电器', 2, 250),
    ('摩托车/电动车/自行车用品/骑行装备', '货架/尾架', 2, 240),
    ('摩托车/电动车/自行车用品/骑行装备', '脚撑/停车架/骑行台', 2, 230),
    ('摩托车/电动车/自行车用品/骑行装备', '记录仪', 2, 220),
    ('摩托车/电动车/自行车用品/骑行装备', '捆扎带/车把带', 2, 210),
    ('摩托车/电动车/自行车用品/骑行装备', '摩托车车锁', 2, 200),
    ('摩托车/电动车/自行车用品/骑行装备', '摩托车GPS定位', 2, 190),
    ('摩托车/电动车/自行车用品/骑行装备', '摩托车手把套', 2, 180),
    ('摩托车/电动车/自行车用品/骑行装备', '摩托车胎压监测仪', 2, 170),
    ('摩托车/电动车/自行车用品/骑行装备', '摩托车USB充电器', 2, 160),
    ('摩托车/电动车/自行车用品/骑行装备', '骑士护具', 2, 150),
    ('摩托车/电动车/自行车用品/骑行装备', '骑行面罩/帽', 2, 140),
    ('摩托车/电动车/自行车用品/骑行装备', '骑行手机支架', 2, 130),
    ('摩托车/电动车/自行车用品/骑行装备', '骑行眼镜/防风镜', 2, 120),
    ('摩托车/电动车/自行车用品/骑行装备', '水壶架', 2, 110),
    ('摩托车/电动车/自行车用品/骑行装备', '头盔饰配', 2, 100),
    ('摩托车/电动车/自行车用品/骑行装备', '头盔锁', 2, 90),
    ('摩托车/电动车/自行车用品/骑行装备', '袖/腿套/束管带', 2, 80),
    ('摩托车/电动车/自行车用品/骑行装备', '自行车灯', 2, 70),
    ('摩托车/电动车/自行车用品/骑行装备', '自行车码表及配件', 2, 60),
    ('摩托车/电动车/自行车用品/骑行装备', '自行车润滑油', 2, 50),
    ('摩托车/电动车/自行车用品/骑行装备', '自行车贴纸/反光贴', 2, 40),
    ('摩托车/电动车/自行车用品/骑行装备', '自行车头盔', 2, 30),
    ('摩托车/电动车/自行车用品/骑行装备', '自行车坐垫套', 2, 20),
    ('摩托车/电动车/自行车用品/骑行装备', '其他骑行装备', 2, 10),
    ('摩托车/电动车/自行车用品/维修工具', '补胎工具', 2, 60),
    ('摩托车/电动车/自行车用品/维修工具', '多功能钳', 2, 50),
    ('摩托车/电动车/自行车用品/维修工具', '螺丝刀', 2, 40),
    ('摩托车/电动车/自行车用品/维修工具', '撬胎棒', 2, 30),
    ('摩托车/电动车/自行车用品/维修工具', '洗链器', 2, 20),
    ('摩托车/电动车/自行车用品/维修工具', '其他维修工具', 2, 10);

  CREATE TEMP TABLE IF NOT EXISTS tmp_wa_cat_ids (
    path TEXT PRIMARY KEY,
    id BIGINT NOT NULL
  ) ON COMMIT DROP;

  -- 一级
  INSERT INTO categories (tenant_id, parent_id, name, level, sort, show_status, created_at, updated_at)
  SELECT v_tenant, 0, t.name, 0, t.sort, 1, NOW(), NOW()
  FROM tmp_wa_cat t
  WHERE t.level = 0
    AND NOT EXISTS (
      SELECT 1 FROM categories c
      WHERE c.tenant_id = v_tenant AND c.parent_id = 0 AND c.name = t.name AND c.deleted_at IS NULL
    );

  INSERT INTO tmp_wa_cat_ids(path, id)
  SELECT t.name, c.id
  FROM tmp_wa_cat t
  JOIN categories c ON c.tenant_id = v_tenant AND c.parent_id = 0 AND c.name = t.name AND c.deleted_at IS NULL
  WHERE t.level = 0
  ON CONFLICT (path) DO NOTHING;

  -- 二级
  INSERT INTO categories (tenant_id, parent_id, name, level, sort, show_status, created_at, updated_at)
  SELECT v_tenant, p.id, t.name, 1, t.sort, 1, NOW(), NOW()
  FROM tmp_wa_cat t
  JOIN tmp_wa_cat_ids p ON p.path = t.parent_path
  WHERE t.level = 1
    AND NOT EXISTS (
      SELECT 1 FROM categories c
      WHERE c.tenant_id = v_tenant AND c.parent_id = p.id AND c.name = t.name AND c.deleted_at IS NULL
    );

  INSERT INTO tmp_wa_cat_ids(path, id)
  SELECT t.parent_path || '/' || t.name, c.id
  FROM tmp_wa_cat t
  JOIN tmp_wa_cat_ids p ON p.path = t.parent_path
  JOIN categories c ON c.tenant_id = v_tenant AND c.parent_id = p.id AND c.name = t.name AND c.deleted_at IS NULL
  WHERE t.level = 1
  ON CONFLICT (path) DO NOTHING;

  -- 三级
  INSERT INTO categories (tenant_id, parent_id, name, level, sort, show_status, created_at, updated_at)
  SELECT v_tenant, p.id, t.name, 2, t.sort, 1, NOW(), NOW()
  FROM tmp_wa_cat t
  JOIN tmp_wa_cat_ids p ON p.path = t.parent_path
  WHERE t.level = 2
    AND NOT EXISTS (
      SELECT 1 FROM categories c
      WHERE c.tenant_id = v_tenant AND c.parent_id = p.id AND c.name = t.name AND c.deleted_at IS NULL
    );
END
$$;

COMMIT;
