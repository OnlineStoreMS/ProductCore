# ProductCore

多平台电商与线下门店统一的 **商品信息管理系统（PIM）**。

## 技术栈

见 [docs/TECH_STACK.md](./docs/TECH_STACK.md)（Gin + GORM + PostgreSQL + Vue3 + Element Plus + Axios + Vite）

## 目录

```
ProductCore/
├── admin/       # 管理后台 API（鉴权、CRUD）
├── api/         # 对外 Open API（只读）
├── internal/    # model / service / repo / config
├── cmd/api/     # 服务入口
├── configs/     # Viper 配置
├── web/         # 管理后台前端
├── deploy/      # Docker PostgreSQL/Redis
└── docs/
```

## 启动

```bash
# 1. 创建/重置 PostgreSQL 用户与库（密码须与 config.yaml 中 postgres_dsn 一致）
make init-db APP_PASSWORD=你的密码 SUDO=1

# 2. 验证连接
psql -h 127.0.0.1 -U productcore -d productcore -c "SELECT 1"

# 3. 配置连接
cp configs/config.example.yaml configs/config.yaml
# 编辑 postgres_dsn 中的 user / password

# 4. 启动后端（自动建表 + 空库时写入演示数据）
make build
./bin/productcore -config configs/config.yaml

# 4. 前端
cd web && npm install && npm run dev
```

可选 Docker 依赖（无本地 PG 时）：`make docker-up`

- 管理后台：http://localhost:5173  
- Admin API：http://localhost:8090/api/v1/admin/*  
- Open API：http://localhost:8090/api/v1/open/*  

## 进度

- [x] 管理后台 UI（Vue 3 + Element Plus）
- [x] Go REST API（商品/品牌/分类/分组/SKU）
- [x] repo 分层 + Admin/Open API 分离
- [x] Redis 商品缓存 + 变更事件（Stream）
- [x] 本地/MinIO 图片上传
- [x] Swagger 文档 + Admin 鉴权（可选开启）
- [x] WangEditor 商品详情编辑器

## 架构文档

- [TECH_STACK.md](./docs/TECH_STACK.md) — 技术选型与实施规范
- [PRODUCT_EDIT_SCHEMA.md](./docs/PRODUCT_EDIT_SCHEMA.md) — 编辑页反向字段设计（Frontend → Backend）
- [ROADMAP.md](./docs/ROADMAP.md) — 整体路线图
- [ONLINE_OFFLINE.md](./docs/ONLINE_OFFLINE.md) — 线上/线下双体系
- [SUPPLY_CHAIN.md](./docs/SUPPLY_CHAIN.md) — 供应链与追溯
- [AFTER_SALES.md](./docs/AFTER_SALES.md) — 电商售后
- [STORE_COLLECTION.md](./docs/STORE_COLLECTION.md) — 门店收款
- [OMNI_CHANNEL.md](./docs/OMNI_CHANNEL.md) — 公域私域联动
