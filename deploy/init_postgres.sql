-- 首次使用前，在 PostgreSQL 中创建数据库（需先连接到 postgres 库）
--
--   psql -U postgres -f deploy/init_postgres.sql
--
-- 或在 psql 交互模式中执行下方语句。

CREATE DATABASE productcore
  ENCODING 'UTF8'
  TEMPLATE template0;
