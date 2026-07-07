# ProductCore 整体架构与演进规划

> 目标：先做好商品底库（PIM），再扩展为「**商品 → 供应商 → 采购 → 仓储 → 订单 → 发货 → 物流**」的统一业务中台，支持每一张订单全链路追溯。供应链与追溯详见 [SUPPLY_CHAIN.md](./SUPPLY_CHAIN.md)。

---

## 1. 总体愿景

```
                    ┌─────────────────────────────────────────────────────┐
                    │              统一管理后台 (Vue)                       │
                    │ 商品│供应商│采购│入库│出库│订单│发货│售后│追溯│门店│维修  │
                    └────────────────────────┬────────────────────────────┘
                                             │
                    ┌────────────────────────▼────────────────────────────┐
                    │                   API Gateway / BFF                  │
                    └────────────────────────┬────────────────────────────┘
     ┌──────────┬──────────┬───────────────┼───────────────┬──────────┬──────────┐
     ▼          ▼          ▼               ▼               ▼          ▼          ▼
ProductCore   VMS/PMS     WMS            OrderCenter    Fulfill    Logistics  TraceHub
 商品/PIM   供应商/采购    仓储            订单/OMS       打包发货    快递轨迹   全链路追溯
     │          │          │               │               │          │          │
     └──────────┴──────────┴──── SKU / 批次 / 库存流水 / trace_id ──┴──────────┘
                                             │
                              Integration Hub（各电商平台）
```

**核心原则：**

1. **商品先行**：ProductCore 是底库，采购/仓储/订单全部引用中央 SKU
2. **批次追溯**：出库关联批次 → 入库 → 采购单 → 供应商，支持打包视频等证据
3. **trace_id 贯穿**：一张销售订单从下单到采购/入库/出库/物流/视频，同一追溯链
4. **订单统一、来源多样**：一套 OMS + 标准来源字段（电商平台/门店/维修）
5. **先模块化单体，后按需拆分**：Go monorepo，Integration / Worker 可独立进程
6. **事件驱动同步**：状态变更、库存变动、平台回传通过消息队列解耦

---

## 2. 领域划分（Bounded Context）

| 领域 | 职责 | 与 ProductCore 关系 |
|------|------|---------------------|
| **ProductCore (PIM)** | SPU/SKU、分类、品牌、分组、详情、平台商品映射 | 当前重点 |
| **Vendor (VMS)** | 供应商档案、供货 SKU、价目、交期 | 采购引用 |
| **Procurement (PMS)** | 采购申请、采购单、到货 | 可关联销售单 trace_id |
| **Warehouse (WMS)** | 仓库、入库、出库、盘点、调拨 | 批次入库生成 lot |
| **Inventory (IMS)** | 批次库存、占用、扣减、流水 | 出库 FIFO 扣批次 |
| **OrderCenter (OMS)** | 销售订单汇聚、状态机、售后 | 引用 sku_id，记录来源渠道 |
| **Fulfillment** | 拣货、打包、打包视频、发货 | 关联 outbound + trace |
| **Logistics (TMS)** | 运单、轨迹、平台回传 | 关联 outbound |
| **TraceHub** | trace_id、单据图谱、操作时间线 | 跨模块追溯入口 |
| **AfterSale (RMA)** | 电商售后、退货签收、开箱视频、退款回传 | 继承 SO 的 trace_id 与 shop |
| **Integration Hub** | 各平台 API/Webhook | 订单/售后拉取、发货与凭证回传 |
| **Store (StoreCore / OSMS)** | 物理门店、收银台、门店销售/服务单、门店库存与采购、监控 | 引用 ProductCore SKU；供应商来自 SupplyCore |

---

## 3. 分阶段路线图

> **当前优先级（2026-06）：** IAM Phase C 暂停；按 [VMS_PMS_DESIGN.md](./VMS_PMS_DESIGN.md) 推进供应链 **M1→M3**（供应商 / 采购），再进入 OMS。详见 [SUPPLY_CHAIN.md](./SUPPLY_CHAIN.md)。

### Phase 1 — 商品底库（进行中）

- [ ] ProductCore Go API：商品 CRUD、SKU 矩阵、分类/品牌/分组
- [ ] 图片上传 OSS、商品详情
- [ ] 管理后台联调（替换 Demo Mock）
- [ ] 预留 `platform_listing`、`platform_sku_mapping` 表（暂可不接 API）

**交付标准：** 可在后台完整维护商品，SKU 数据准确，API 可供内部调用。

---

### Phase 1.5 — 供应商 + 采购（VMS/PMS，下一步）

详见 [VMS_PMS_DESIGN.md](./VMS_PMS_DESIGN.md)。

- [ ] SKU 自营/代发、多供应商报价、发货地、拿货价
- [ ] 采购单、付款、物流、附件（销售单/付款截图）
- [ ] 按 SKU 查询供货方案（为后续订单选供做准备）

**交付标准：** 可在后台维护供应商与 SKU 供货关系；可创建采购单并跟踪付款、物流与凭证。

---

### Phase 2 — OMS 核心 + 单渠道验证（约 2~3 个月）

优先选 **1 个** 你方订单量最大或 API 最成熟的渠道（建议：微信商城小程序 或 抖店）做闭环。

- [ ] 统一订单模型（见第 4 节）
- [ ] 订单状态机：待付款 → 待发货 → 已发货 → 已完成 / 已关闭 / 售后中
- [ ] 管理后台：订单列表、详情、备注、人工改价（如需要）
- [ ] 与 ProductCore SKU 映射校验
- [ ] 第一个渠道 Adapter：订单拉取或 Webhook 接入

**交付标准：** 该渠道订单可在后台统一查看和处理。

---

### Phase 3 — 发货与物流回传（约 1~2 个月）

- [ ] 发货单（支持拆单、多包裹）
- [ ] 快递公司 + 物流单号录入/电子面单
- [ ] 物流轨迹查询（快递100 / 菜鸟等）
- [ ] **发货回传 Adapter**：按平台要求上传 `delivery_sn` 到抖店/淘宝等
- [ ] 库存占用 → 发货扣减

**交付标准：** 后台发货后，对应平台订单状态同步为已发货，物流单号可见。

---

### Phase 4 — 多渠道订单汇聚（约 3~4 个月）

按优先级逐个接入：

| 优先级 | 渠道 | 接入要点 |
|--------|------|----------|
| P0 | 抖店 | 订单 API、发货回传、售后 |
| P0 | 淘宝/天猫 | 聚石塔/开放平台、子账号授权 |
| P1 | 微信商城小程序 | 自建订单，支付回调 |
| P1 | 小红书 | 订单与发货 API |
| P2 | 视频号小店 | 微信生态，可与小程序共用部分逻辑 |
| P2 | 闲鱼 | 接口相对特殊，可能需人工+半自动 |

每接一个渠道 = 实现一个 **Adapter**，不改 OMS 核心表结构。

---

### Phase 5 — StoreCore / OSMS 门店管理（进行中）

> 独立应用：**StoreCore**（`/home/asialeaf/projects/StoreCore`），API `:8094`、Web `:5179`，与 ProductCore / UserCore / SupplyCore 同栈（Go + Vue + ACR）。文档内历史名称 **StoreHub** 与本项目 **StoreCore** 指同一门店域。

**与电商 OMS 的边界：**

| 场景 | 系统 | 特征 |
|------|------|------|
| 即时零售 | StoreCore 收银台 | 现场选品结算，现金/静态二维码等，电子小票 |
| 线下订货/派送 | StoreCore 销售订单 | 手动建单，订货后提货、送货上门、发快递 |
| 维修/预约服务 | StoreCore 服务工单 | 中高端自行车类目，维修/保养/预约 |
| 平台电商订单 | OMS（Phase 2+） | 抖店/淘宝等，`source_channel != store` |

**六大模块（M1 已脚手架，按序深化）：**

#### 5.1 收银台（POS）

- [x] M0：同步租户 ProductCore 商品/SKU 搜索、购物车、创建即时零售单
- [x] M0：多种结算方式枚举（现金、静态二维码、微信/支付宝预留）
- [x] M0：电子小票 HTML 生成
- [ ] M1：创建订单后微信/支付宝扫码支付（零售店模式）
- [ ] M2：大小票模板设计、`receipt_templates` 管理
- [ ] M3：打印机管理、云打印机对接（预留接口）

#### 5.2 销售订单

- [x] M0：门店销售单模型（与 POS 即时零售分离）
- [ ] M1：手动创建、订货后提货流程
- [ ] M2：送货上门 / 发快递履约、地址与物流单
- [ ] M3：与销售单联动的采购需求（`need_procurement`）

#### 5.3 服务工单（维修/预约）

- [x] M0：服务单模型（维修、预约时间、设备信息）
- [ ] M1：工单状态机、工程师派工
- [ ] M2：与 [STORE_COLLECTION.md](./STORE_COLLECTION.md) 收款单联动
- [ ] M3：可选配件 SKU 出库

#### 5.4 门店库存

- [x] M0：门店 × SKU 库存表（OSMS 库存子集）
- [ ] M1：入库/出库/盘点流水
- [ ] M2：与平台级 IMS/WMS 同步（平台库存系统开发后）
- [ ] M3：收银/销售单占用与扣减

#### 5.5 门店采购

- [x] M0：门店采购单模型
- [ ] M1：销售订单驱动采购、门店备货采购
- [ ] M2：供应商从 **SupplyCore** 选取（HTTP 集成）
- [ ] M3：到货入库更新门店库存

#### 5.6 监控管理

- [x] M0：监控设备档案（流地址/回放地址预留）
- [ ] M1：实时预览（对接 NVR/云平台 SDK）
- [ ] M2：录像查阅、时间轴检索

**交付标准（M1）：** UserCore 应用中心可进入门店管理；可维护门店档案；收银台可搜索 ProductCore SKU 并完成结算；各模块列表 API 可用。

**原 Phase 5 门店订货摘要（并入销售订单）：**
- 订单来源 `source=store`，关联门店 ID、导购员
- 可支持「门店自提 / 总部代发 / 快递」等履约方式
- 库存策略：门店仓 vs 中心仓（待 IMS 统一）

---

## 4. 统一订单模型设计

### 4.1 核心思路

借鉴 mall 的 `OmsOrder`，但扩展 **渠道维度** 和 **平台原始 ID**，避免后期返工。

```
orders                  -- 订单主表（统一）
order_items             -- 订单商品行
order_addresses         -- 收货地址快照
order_payments          -- 支付信息
order_shipments         -- 发货单（支持多包裹）
order_shipment_items    -- 包裹明细
order_status_logs       -- 状态流水
order_platform_ext      -- 平台扩展（JSON 或键值）
```

### 4.2 关键字段建议

**orders 主表：**

| 字段 | 说明 |
|------|------|
| order_no | 内部统一单号 |
| order_type | `retail` / `store` / `service` |
| source_channel | `douyin` / `taobao` / `xhs` / `channels` / `xianyu` / `wx_mall` / `store` / `manual` |
| platform_order_id | 平台原始订单号（唯一索引 + channel 联合） |
| shop_id | 店铺/门店 ID |
| status | 统一状态枚举 |
| buyer_name / buyer_phone | 买家信息 |
| total_amount / pay_amount / freight_amount | 金额 |
| pay_status / pay_time | 支付 |
| remark / seller_remark | 备注 |
| raw_payload | JSON，存平台原始报文（排查问题用） |

**order_items：**

| 字段 | 说明 |
|------|------|
| sku_id | ProductCore 中央 SKU（可空，未映射时告警） |
| platform_sku_id / platform_item_id | 平台 SKU/商品 ID |
| product_name / sku_specs | 快照 |
| quantity / price / total_amount | 数量金额 |

### 4.3 统一状态机

```
                    ┌──────────┐
                    │ 待付款    │
                    └────┬─────┘
                         │ 支付成功
                    ┌────▼─────┐
         ┌──────────│ 待发货    │──────────┐
         │          └────┬─────┘          │
         │ 取消/超时      │ 发货           │ 部分发货
         │          ┌────▼─────┐     ┌────▼─────┐
         │          │ 已发货    │────►│ 部分发货  │
         │          └────┬─────┘     └──────────┘
         │               │ 确认收货
         │          ┌────▼─────┐
         │          │ 已完成    │
         │          └──────────┘
         │
    ┌────▼─────┐
    │ 已关闭    │
    └──────────┘

售后分支：申请中 → 退货中 → 已退款 / 已拒绝
```

各平台状态码通过 Adapter **映射** 到统一状态，不要反过来。

---

## 5. 渠道集成层（Integration Hub）

### 5.1 Adapter 模式

每个平台一个实现，接口统一：

```go
type ChannelAdapter interface {
    // 商品
    ListProducts(ctx context.Context, shopID string) ([]PlatformProduct, error)
    PublishProduct(ctx context.Context, skuID int64) error

    // 订单
    PullOrders(ctx context.Context, since time.Time) ([]PlatformOrder, error)
    HandleOrderWebhook(ctx context.Context, body []byte) (*PlatformOrder, error)

    // 履约
    ShipOrder(ctx context.Context, req ShipRequest) error  // 回传物流
    SyncOrderStatus(ctx context.Context, platformOrderID string) (OrderStatus, error)
}
```

新增渠道 = 新增 `adapter/douyin`、`adapter/taobao` 包，**不动 OMS 核心**。

### 5.2 同步策略

| 方式 | 适用 |
|------|------|
| **Webhook 推送** | 抖店、微信类，实时性好 |
| **定时轮询** | 淘宝、闲鱼等，按更新时间增量拉 |
| **手动导入** | 过渡方案，CSV/Excel |

建议：**Webhook 为主 + 定时补偿同步**（防止漏单）。

### 5.3 商品映射表（ProductCore 侧）

```
platform_shop          -- 各平台店铺授权
platform_listing       -- SPU 在各平台的上架记录
platform_sku_mapping   -- 中央 sku_id ↔ 平台 sku_id
```

订单接入时，通过 `platform_sku_mapping` 解析为中央 SKU；映射缺失时 **告警 + 人工绑定**。

---

## 6. 发货与物流

### 6.1 发货流程

```
待发货订单 → 创建 shipment → 选快递/打单 → 写入 delivery_sn
    → 扣减库存 → 发布 OrderShipped 事件 → 各平台 Adapter 回传
    → 订阅物流轨迹 → 更新订单/通知客服
```

### 6.2 物流单号回传

- 各平台 API 字段不同（快递公司编码、单号、发货时间），在 Adapter 内转换
- 支持 **多次发货**（分包裹），平台若只支持整单发货需在 Adapter 层做策略
- 失败重试：写入 `outbox` 表 + 定时重推，避免漏回传

### 6.3 推荐第三方

- **电子面单**：快递100、菜鸟、各快递公司直连
- **轨迹查询**：快递100 订阅推送

---

## 7. 库存策略（IMS）

商品库与订单衔接时，库存是必答题：

| 策略 | 说明 |
|------|------|
| **中心仓库存** | 初期简单：一个 SKU 一个可售数 |
| **渠道独占库存** | 抖店 100、淘宝 50，避免超卖 |
| **下单占用** | 待付款占用，超时释放；待发货锁定 |
| **发货扣减** | 实际出库时扣减 |

建议 Phase 1 商品库阶段 **只记录 stock 字段**，Phase 2 做 OMS 时引入 **占用/扣减流水表**。

---

## 8. 项目结构建议（Go Monorepo）

```
ProductCore/                    # 可升级为 biz-platform  monorepo
├── cmd/
│   ├── product-api/            # 商品服务入口
│   ├── order-api/              # 订单服务入口（Phase 2）
│   └── integration-worker/     # 渠道同步 Worker（Phase 2+）
├── internal/
│   ├── product/                # 商品领域
│   ├── order/                  # 订单领域
│   ├── fulfillment/            # 发货领域
│   ├── inventory/              # 库存领域
│   └── integration/
│       ├── adapter/
│       │   ├── douyin/
│       │   ├── taobao/
│       │   └── wxmall/
│       └── webhook/
├── pkg/                        # 公共库
├── web/                        # 管理前端
├── docs/
└── deploy/
```

**Admin 前端：** 现阶段可在一个 Vue 项目里加「订单管理」菜单；模块多了以后用微前端或路由分包即可。

---

## 9. 技术栈延续建议

| 层级 | 建议 |
|------|------|
| 后端 | 继续 **Go + Gin + GORM**，与 ProductCore 一致 |
| 数据库 | **PostgreSQL**（JSON 字段存平台扩展、raw_payload） |
| 缓存/锁 | Redis（库存锁、幂等、Webhook 去重） |
| 队列 | **Redis Stream**（初期）→ RabbitMQ/NATS（订单量上来后） |
| 任务调度 | asynq / robfig/cron（拉单、回传重试） |
| 前端 | 继续 Vue 3 + Element Plus，订单列表/发货台复用现有布局 |

---

## 10. 现在就要预留的设计

在 Phase 1 做商品库时，建议 **表结构和 API 预留**，避免 Phase 2 大改：

1. **SKU 表** 保留 `sku_code` 全局唯一，作为各平台映射键
2. **platform_shop / platform_sku_mapping** 表先建好
3. **商品变更事件** 先定义结构体（即使暂不发 MQ）：`ProductCreated`, `SkuStockChanged`
4. **订单号规则** 提前定好：`PC{yyyyMMdd}{seq}` 与平台单号分离
5. **统一错误码 / 日志 trace_id**，便于跨模块排查

---

## 11. 风险与建议

| 风险 | 应对 |
|------|------|
| 各平台 API 差异大、变更频繁 | Adapter 隔离 + raw_payload 保留原始数据 |
| SKU 映射不全导致订单对不上 | 映射缺失告警、后台「绑定 SKU」工具 |
| 超卖 | 下单占用库存 + 渠道库存隔离 |
| 范围膨胀 | 严格按 Phase 交付，维修/门店可 Phase 5 再做 |
| mall 老系统并行 | ProductCore 新库；订单可逐步从 mall 迁移或双写过渡 |

---

## 12. 推荐优先级总结

```
现在     ──► ProductCore 商品库 + SKU + 平台映射表
下一步   ──► OMS 核心 + 1 个渠道 + 后台订单列表
再下一步 ──► 发货 + 物流回传
然后     ──► 抖店/淘宝/小红书/视频号/闲鱼 逐个接入
最后     ──► StoreCore 门店管理（OSMS：收银/销售/服务/库存/采购/监控）
```

**一句话：** 商品库是「货」的中枢，订单系统是「单」的中枢，Integration 是「渠道」的翻译层，**StoreCore（OSMS）** 是「线下门店」的中枢；四者通过 **中央 SKU** 和 **事件** 连接，分阶段建设、模块边界清晰，后续扩展不会推倒重来。

---

## 13. StoreCore（OSMS）架构速览

```
UserCore (IAM) ──JWT──► StoreCore (:8094 / :5179)
                              │
         ┌────────────────────┼────────────────────┐
         ▼                    ▼                    ▼
   ProductCore           SupplyCore            平台 IMS/WMS
   商品/SKU 底库          供应商/VMS            （后续，库存超集）
         │                    │
         └──── sku_id ────────┴──── supplier_id
                              │
                    StoreCore 六大模块
         收银台 │ 销售订单 │ 服务工单 │ 库存 │ 采购 │ 监控
```

| 项目 | 值 |
|------|-----|
| 代码仓库 | `/home/asialeaf/projects/StoreCore` |
| Go module | `storecore` |
| Docker 镜像 | `storecore-api`、`storecore-web` |
| UserCore 应用码 | `storecore`（权限 `store:read` / `store:write`） |
| 平台编排 | `/home/asialeaf/projects/deploy` |

相关文档：[ONLINE_OFFLINE.md](./ONLINE_OFFLINE.md)、[STORE_COLLECTION.md](./STORE_COLLECTION.md)。
