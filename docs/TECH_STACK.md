# ProductCore 技术选型建议

## 项目定位

ProductCore 是**多平台电商商品底库**（PIM - Product Information Management），核心职责是统一管理商品信息，并对外提供 API，后续分发至淘宝、抖音、小红书、闲鱼等渠道。

---

## 推荐技术栈（综合方案）

### 后端：Go

| 组件 | 推荐 | 理由 |
|------|------|------|
| Web 框架 | **Gin** | 生态成熟、性能好、学习成本低，REST API 开发效率高 |
| ORM | **GORM** | 与 mall 类似的关系型模型（SPU/SKU/分类/品牌）迁移友好 |
| 数据库 | **PostgreSQL**（或 MySQL） | 商品 JSON 属性、全文检索扩展能力强 |
| 缓存 | **Redis** | 商品缓存、分布式锁、后续消息队列辅助 |
| API 文档 | **swaggo/swag** | 自动生成 OpenAPI，便于对外 API 对接 |
| 配置 | **Viper** | 多环境配置管理 |
| 依赖注入 | **Wire**（可选） | 中大型项目分层清晰 |
| 文件存储 | **MinIO / 阿里云 OSS** | 商品主图、SKU 图、详情页图片 |
| 消息队列 | **Redis Stream / RabbitMQ**（后期） | 商品变更事件推送到各平台 |

**备选方案：**
- **go-zero**：若后期拆微服务、需要 API 网关与限流，可优先考虑
- **Kratos**：B 站出品，适合长期维护的中大型项目，分层规范

### 前端：Vue 3 管理后台

| 组件 | 推荐 | 理由 |
|------|------|------|
| 框架 | **Vue 3 + TypeScript** | 与现有 mall-admin 技术栈接近，团队迁移成本低 |
| UI 库 | **Element Plus** | 表格、表单、树形分类、上传组件完善，电商后台标配 |
| 构建 | **Vite** | 开发体验好，HMR 快 |
| 路由/状态 | **Vue Router + Pinia** | 标准组合 |
| HTTP | **Axios** | 与 Go API 对接 |
| 富文本 | **WangEditor / TinyMCE** | 商品详情页编辑 |
| SKU 规格 | 自研组件（见 demo） | 参考淘宝/抖店规格矩阵交互 |

**备选方案：**
- **React + Ant Design Pro**：复杂表单与 ProTable 能力更强，适合从零组建新团队
- **Arco Design Vue**：字节系设计，视觉更现代

### 架构分层建议

```
ProductCore/
├── api/          # 对外 Open API（供各电商平台/内部系统调用）
├── admin/        # 管理后台 API（鉴权、CRUD）
├── internal/
│   ├── model/    # SPU、SKU、Category、Brand、ProductGroup
│   ├── service/  # 商品聚合、SKU 矩阵生成、图片处理
│   └── repo/     # 数据访问
├── cmd/api/      # 服务入口
├── configs/      # Viper 配置
├── web/          # 管理前端
└── docs/         # API 文档、数据模型
```

### API 路由

| 路由组 | 路径前缀 | 说明 |
|--------|----------|------|
| Admin | `/api/v1/admin/*` | 管理后台 CRUD，供 `web/` 调用 |
| Open | `/api/v1/open/*` | 只读已上架商品，供 CommerceHub/StoreHub 调用 |

### 开发命令

```bash
make build && ./bin/productcore -config configs/config.yaml
cd web && npm run dev
make docker-up   # MySQL:3307 + Redis:6380
```

### 核心数据模型（参考 mall，面向多平台扩展）

- **SPU（商品）**：名称、分类、品牌、主图、相册、详情、属性模板
- **SKU**：规格组合（颜色+尺码）、价格、库存、SKU 编码、SKU 专属图片
- **Category**：树形分类（支持多级，对接各平台类目映射）
- **Brand**：品牌管理
- **ProductGroup**：商品分组（运营/选品维度）
- **PlatformListing**（后期）：各平台上架状态、平台 SKU ID 映射

---

## 为何不做商城前台

ProductCore 专注 **PIM + API**，不需要 C 端商城页面。各平台（淘宝/抖音等）自有前台，ProductCore 只负责「商品底库 → 同步/上架」。

---

## Demo 说明

`web/` 目录为管理后台 UI Demo，使用 Mock 数据展示：

- 工作台概览
- 商品列表 / 添加编辑（含 SKU 规格矩阵）
- 商品分类（树形）
- 品牌管理
- 商品分组

运行：`cd web && npm install && npm run dev`
