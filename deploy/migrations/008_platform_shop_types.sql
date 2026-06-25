-- 电商平台店铺类型与店铺（SQLite/Postgres 由 GORM AutoMigrate 建表，此文件供 Postgres 部署参考）

CREATE TABLE IF NOT EXISTS platform_shop_types (
  id BIGSERIAL PRIMARY KEY,
  code VARCHAR(32) NOT NULL UNIQUE,
  name VARCHAR(64) NOT NULL,
  logo VARCHAR(512) DEFAULT '',
  sort INT NOT NULL DEFAULT 0,
  enabled SMALLINT NOT NULL DEFAULT 1,
  is_builtin SMALLINT NOT NULL DEFAULT 0,
  remark VARCHAR(256) DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

-- platform_shops 表由 GORM 扩展字段：platform_type_id, external_shop_id, remark, sort
