# VMS / PMS 设计说明

> 在 ProductCore（PIM）之上建设 **供应商管理（VMS）** 与 **采购管理（PMS）**。  
> 与 [ROADMAP.md](./ROADMAP.md)、[SUPPLY_CHAIN.md](./SUPPLY_CHAIN.md) 对齐；IAM Phase C 暂停，供应链 Phase P2/P3 优先。

---

## 1. 业务目标（你的需求映射）

| 需求 | 设计落点 |
|------|----------|
| 商品区分 **自营 / 代发** | SKU 履约属性 + 可选 SPU 默认 |
| SKU 关联多个供应商 | `sku_supplier_offers`（供货报价） |
| 下单时知道「哪些供应商可选、哪些可代发」 | 按 `sku_id` 查询供货方案 API |
| 同一 SKU 不同供应商：拿货价、发货地 | 报价行上的 `supply_price`、`ship_from_*` |
| 拿货时对比价格与发货地 | 供货方案列表 + 排序/筛选 UI |
| 采购单全生命周期 | `purchase_orders` 状态机 |
| 采购金额、打款账号、是否打款 | `purchase_payments` |
| 物流：已发货/运输中、单号、预计到货 | PO 行/发货子表 `purchase_shipments` |
| 附件：供货商销售单、付款截图等 | `purchase_attachments` |

**原则：** 订单将来走 OMS；现阶段先把 **货从哪来、怎么买、怎么跟** 做扎实。SKU 仍是全局主键（`tenant_id` 隔离）。

---

## 2. 自营 vs 代发

### 2.1 定义

| 模式 | 英文 | 含义 |
|------|------|------|
| **自营** | `self` | 自有库存，从 **自有仓库/地址** 发货 |
| **代发** | `dropship` | 不占用（或弱占用）自有库存，由 **供应商直发** 给买家 |
| **混合** | `mixed` | 同一 SPU 下不同 SKU 或不同订单行可不同模式 |

### 2.2 字段建议

**SPU（products）— 默认，可被 SKU 覆盖**

| 字段 | 类型 | 说明 |
|------|------|------|
| `default_fulfillment_mode` | enum | `self` / `dropship` / `mixed` |

**SKU（product_skus）— 订单匹配以 SKU 为准**

| 字段 | 类型 | 说明 |
|------|------|------|
| `fulfillment_mode` | enum | `inherit`（跟 SPU）/ `self` / `dropship` |
| `primary_supplier_id` | FK | 默认主供（代发场景常用） |

**何时用哪个：** 销售订单行绑定 `sku_id` 后，解析有效履约模式 → 查 `sku_supplier_offers` 得到可选供应商与代发能力。

---

## 3. VMS — 供应商管理

### 3.1 供应商主档 `suppliers`

| 字段 | 说明 |
|------|------|
| tenant_id | 租户 |
| code | 供应商编码，租户内唯一 |
| name | 名称 |
| short_name | 简称 |
| status | 启用/停用 |
| contact_name / phone / email | 主联系人 |
| remark | 备注 |
| default_payment_terms | 账期说明（可选） |
| bank_name / bank_account / account_name | 默认收款信息（PO 可覆盖） |

### 3.2 供应商发货地址 `supplier_addresses`

一个供应商可有多个发货仓/档口：

| 字段 | 说明 |
|------|------|
| supplier_id | |
| label | 如「深圳仓」「义乌档口」 |
| contact_name / phone | |
| province / city / district / address | 结构化地址 |
| is_default | 默认发货地 |
| status | |

### 3.3 SKU 供货报价 `sku_supplier_offers`（核心）

**一个 SKU + 一个供应商 = 一条报价**（同一 SKU 可多条，对应不同发货地或不同价目）。

| 字段 | 说明 |
|------|------|
| tenant_id | |
| sku_id | ProductCore SKU |
| supplier_id | |
| supplier_sku_code | 对方货号/编码 |
| supply_price | **拿货价**（采购单价） |
| currency | 默认 CNY |
| min_order_qty | 起订量 |
| lead_time_days | 交期（天） |
| ship_from_address_id | 发货地址 → `supplier_addresses` |
| supports_dropship | **是否支持代发**（1/0） |
| supports_self_stock | 是否供货到仓（采购入库用，1/0） |
| is_primary | 是否主供 |
| priority | 排序（比价默认顺序） |
| status | 启用/停用 |
| remark | |

**唯一约束：** `(tenant_id, sku_id, supplier_id, ship_from_address_id)` 或简化为 `(tenant_id, sku_id, supplier_id)` 若一供一价。

### 3.4 查询：下单/拿货时「有哪些可选」

```
GET /api/v1/admin/skus/{skuId}/supply-options
```

响应示例（概念）：

```json
{
  "skuId": 1001,
  "fulfillmentMode": "dropship",
  "offers": [
    {
      "offerId": 1,
      "supplierId": 10,
      "supplierName": "深圳XX商行",
      "supplyPrice": 28.50,
      "supportsDropship": true,
      "shipFrom": { "label": "深圳仓", "city": "深圳市", "address": "..." },
      "leadTimeDays": 2,
      "isPrimary": true
    },
    {
      "offerId": 2,
      "supplierId": 11,
      "supplierName": "义乌YY",
      "supplyPrice": 26.00,
      "supportsDropship": true,
      "shipFrom": { "label": "义乌档口", "city": "金华市", "address": "..." },
      "leadTimeDays": 3,
      "isPrimary": false
    }
  ]
}
```

前端：**比价表格** — 按价格、发货地、交期排序；代发场景只展示 `supports_dropship=true` 的行。

---

## 4. PMS — 采购管理

### 4.1 采购单主表 `purchase_orders`

| 字段 | 说明 |
|------|------|
| tenant_id | |
| po_no | 采购单号，租户内唯一 |
| supplier_id | |
| status | 见 4.2 状态机 |
| total_amount | 采购总额（汇总） |
| currency | |
| expected_arrival_date | 整单预计到货（可选） |
| warehouse_id | 自营入库目标仓（非代发 PO 必填） |
| fulfillment_type | `stock_in`（采购入仓）/ `dropship`（代发直邮，可关联销售单） |
| ref_trace_id | 关联追溯链（将来 OMS） |
| ref_so_id | 关联销售单（代发采购时常用） |
| buyer_id / buyer_name | 采购员 |
| remark | |
| ordered_at | 下单给供应商时间 |
| completed_at | |

### 4.2 采购单状态机

```
草稿(draft)
  → 待审核(pending_approval)     [可选，小团队可跳过]
  → 已下单(ordered)              已发给供应商
  → 待付款(pending_payment)      [可与 ordered 合并视流程而定]
  → 已付款(paid)
  → 部分发货(partial_shipped)
  → 运输中(in_transit)           至少一行有物流且在途
  → 部分到货(partial_received)   自营入仓场景
  → 已完成(completed)
  → 已取消(cancelled)
```

**说明：**

- **代发 PO**：重点跟 `partial_shipped` → `in_transit` → `completed`（客户签收），不一定走入库
- **入仓 PO**：`partial_received` → 入库单 GRN（Phase P3 与 WMS 衔接）

### 4.3 采购明细 `purchase_order_items`

| 字段 | 说明 |
|------|------|
| po_id | |
| sku_id | |
| offer_id | 可选，指向 `sku_supplier_offers` 快照来源 |
| supplier_sku_code | 快照 |
| qty | 采购数量 |
| unit_price | 成交单价（快照，可与 offer 不同） |
| line_amount | qty × unit_price |
| expected_arrival_date | 行级预计到货 |
| received_qty | 已到货数量（入仓） |
| remark | |

### 4.4 采购发货/物流 `purchase_shipments`

支持 **一张 PO 多次发货、多个包裹**：

| 字段 | 说明 |
|------|------|
| po_id | |
| shipment_no | 内部发货批次号 |
| status | `pending` / `shipped` / `in_transit` / `delivered` / `exception` |
| carrier_code / carrier_name | 快递公司 |
| tracking_no | **物流单号** |
| shipped_at | 发货时间 |
| expected_arrival_date | **预计到货** |
| delivered_at | 实际签收/到货 |
| ship_from_address_id | 发货地快照 |
| receiver_name / phone / address | 收货方（代发=客户地址；入仓=仓库地址） |
| remark | |

明细行关联：`purchase_shipment_items`（shipment_id, po_item_id, qty）

### 4.5 付款 `purchase_payments`

| 字段 | 说明 |
|------|------|
| po_id | |
| pay_amount | **付款金额** |
| pay_method | 银行转账 / 支付宝 / 微信 / 其他 |
| pay_account | **打款账号**（我方付出账户或对方收款账户，字段可拆 payer/payee） |
| payee_account | 供应商收款账号 |
| payee_name | 收款户名 |
| pay_status | `unpaid` / `partial` / `paid` |
| paid_at | 打款时间 |
| remark | |

PO 主表可冗余 `pay_status` 汇总，明细以 payments 为准。

### 4.6 附件 `purchase_attachments`

| 字段 | 说明 |
|------|------|
| po_id | |
| payment_id | 可选，关联某次付款 |
| file_type | `supplier_sales_order`（供货商销售单）/ `payment_screenshot`（付款截图）/ `contract` / `other` |
| file_name | |
| file_url | OSS/MinIO |
| uploaded_by | |
| remark | |

上传走现有 ProductCore 存储（MinIO），与商品图同一套 `storage` 配置。

---

## 5. 与 ProductCore 现有模块的关系

```
ProductCore (现有)
  products / product_skus     ← 增加 fulfillment 字段
  brands / categories         ← 不变

VMS (新增 internal/vendor)
  suppliers
  supplier_addresses
  sku_supplier_offers

PMS (新增 internal/procurement)
  purchase_orders / items
  purchase_shipments / items
  purchase_payments
  purchase_attachments
```

- **API 前缀建议：** `/api/v1/admin/vendors/*`、`/api/v1/admin/purchase-orders/*`
- **鉴权：** 沿用 UserCore JWT + `tenant_id`（与商品一致）
- **权限点（UserCore 后续可加）：** `vendor:manage`、`purchase:manage`、`purchase:approve`

---

## 6. 管理后台菜单（Vue）

在现有 Sidebar 增加 **供应链** 分组：

| 菜单 | 页面 |
|------|------|
| 供应商管理 | 列表 / 新建 / 编辑 / 发货地址 |
| SKU 供货关系 | 可从商品编辑页 Tab「供货方案」进入，或独立列表按 SKU 搜 |
| 采购单列表 | 筛选：状态、供应商、是否已付款、日期 |
| 采购单详情 | 明细、物流、付款、附件、状态流转 |
| （后期）采购入库 | 对接 WMS GRN |

**商品编辑页增强：**

- Tab **供货方案**：维护该 SKU 的多供应商报价、代发开关、发货地
- 字段 **履约方式**：自营 / 代发 / 继承 SPU

---

## 7. 分步实施建议（可交付里程碑）

与 [SUPPLY_CHAIN.md §8](./SUPPLY_CHAIN.md) 的 P2/P3 对齐：

### 里程碑 M1 — VMS 基础（约 1~2 周）

- [ ] 表：`suppliers`、`supplier_addresses`、`sku_supplier_offers`
- [ ] SKU/SPU `fulfillment_mode` 迁移
- [ ] API：供应商 CRUD、地址 CRUD、SKU 供货 CRUD
- [ ] API：`GET /skus/{id}/supply-options`
- [ ] 前端：供应商管理、商品编辑「供货方案」Tab

**交付：** 可为任意 SKU 配置多供应商、价格、发货地、是否代发；可查询比价列表。

### 里程碑 M2 — PMS 采购单核心（约 2 周）

- [ ] 表：`purchase_orders`、`purchase_order_items`
- [ ] 状态机：草稿 → 已下单 → 已付款 → 已完成/取消
- [ ] 从 SKU 供货方案 **一键生成 PO 行**（带入默认价、供应商）
- [ ] 前端：采购单列表、新建、详情、编辑明细

**交付：** 可完整录入采购单与金额，跟单到「已下单/已付款」。

### 里程碑 M3 — 物流 + 附件 + 付款（约 1~2 周）

- [ ] 表：`purchase_shipments`、`purchase_payments`、`purchase_attachments`
- [ ] 上传：供货商销售单、付款截图
- [ ] 物流单号、预计到货、运输中状态
- [ ] 前端：PO 详情 Tab（物流 / 付款 / 附件）

**交付：** 满足你描述的采购跟单与凭证留存。

### 里程碑 M4 — 与订单衔接（OMS 阶段）

- [x] 销售订单行选 SKU → 调 `supply-options` → 生成代发 PO（自发货行跳过采购）
- [x] `ref_so_id` / `trace_id` 贯通（SupplyCore 独立实现）

**交付：** SupplyCore 销售单 + 寻源 + 代发 PO；外部 OMS 可调用 `POST /sourcing/dropship-purchase-order`。

## 8. 暂不做的（避免范围膨胀）

- 批次库存 `inventory_lots`、入库单 GRN（WMS P3，M2 之后）
- 采购审批流多级、供应商账期对账
- 自动向供应商下 API 单
- IAM 审计日志（Phase C 暂停）

---

## 9. 技术约定

- **Go 模块路径：** 先在 ProductCore monorepo 内 `internal/vendor`、`internal/procurement`，不急于拆进程
- **编号规则：** `PO{yyyyMMdd}{4位序}`、`SUP{code}` 供应商编码手工或自动生成
- **金额：** `decimal(12,2)`，行金额后端重算防篡改
- **附件：** 复用 `storage`；路径如 `procurement/{po_id}/attachments/`

---

## 10. 一句话

**SKU 是锚点；供应商报价回答「从谁买、多少钱、从哪发、能否代发」；采购单回答「买了什么、付没付钱、货在哪、凭证在哪」。** 先 M1→M2→M3，再接 OMS 订单选供。
