# ProductCore

多平台电商与线下门店统一的 **商品信息管理系统（PIM）**。

## 技术栈

见 [docs/TECH_STACK.md](./docs/TECH_STACK.md)（Gin + GORM + MySQL/SQLite + Vue3 + Element Plus + Axios + Vite）

## 目录

```
ProductCore/
├── admin/       # 管理后台 API（鉴权、CRUD）
├── api/         # 对外 Open API（只读）
├── internal/    # model / service / repo / config
├── cmd/api/     # 服务入口
├── configs/     # Viper 配置
├── web/         # 管理后台前端
├── deploy/      # Docker MySQL/Redis
└── docs/
```

## 启动

```bash
# 可选：MySQL + Redis
make docker-up

# 后端
make build
./bin/productcore -config configs/config.yaml

# 前端
cd web && npm install && npm run dev
```

- 管理后台：http://localhost:5173  
- Admin API：http://localhost:8090/api/v1/admin/*  
- Open API：http://localhost:8090/api/v1/open/*  

## 进度

- [x] 管理后台 UI（Vue 3 + Element Plus）
- [x] Go REST API（商品/品牌/分类/分组/SKU）
- [x] repo 分层 + Admin/Open API 分离
- [ ] Redis 缓存
- [ ] 图片上传 OSS
- [ ] Swagger 发布
- [ ] 鉴权

## 架构文档

- [TECH_STACK.md](./docs/TECH_STACK.md) — 技术选型与实施规范
- [ROADMAP.md](./docs/ROADMAP.md) — 整体路线图
- [ONLINE_OFFLINE.md](./docs/ONLINE_OFFLINE.md) — 线上/线下双体系
- [SUPPLY_CHAIN.md](./docs/SUPPLY_CHAIN.md) — 供应链与追溯
- [AFTER_SALES.md](./docs/AFTER_SALES.md) — 电商售后
- [STORE_COLLECTION.md](./docs/STORE_COLLECTION.md) — 门店收款
- [OMNI_CHANNEL.md](./docs/OMNI_CHANNEL.md) — 公域私域联动
