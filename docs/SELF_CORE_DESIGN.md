# SelfCore — 自营中心设计（SupplyCore 方向反转）

> 独立应用 **SelfCore**（自营中心）：管理自营履约订单与分销（批发）业务。  
> 业务对称于 [VMS_PMS_DESIGN.md](./VMS_PMS_DESIGN.md) / SupplyCore，但资金与货物流向相反。

| 项 | 值 |
|----|-----|
| 目录 | `/home/asialeaf/projects/SelfCore` |
| API | `:8103` |
| Web | `:5187` |
| UserCore app | `selfcore`（`self:read` / `self:write`） |
| DB / MinIO | `selfcore` |

---

## 1. 与 SupplyCore 对照

| SupplyCore（向供应商买） | SelfCore（向分销商卖） |
|--------------------------|------------------------|
| 供应商 `suppliers` | 分销商 `distributors` |
| 拿货价 `supply_price`（我方付） | 批发价 `wholesale_price`（对方付） |
| 采购单 `purchase_orders` | 分销订单 `dist_orders` |
| 我方付款 `purchase_payments` | 分销商收款 `dist_receipts` |
| 供应商发货给我 `purchase_shipments` | 我方发货给分销商/终端 `dist_shipments` |
| 代发 / 采购入库 | **自营订单**（承接本店/电商自营） / **分销订单**（承接分销商） |

**原则：** OMS（OrderCore）仍是销售订单汇聚；SelfCore 负责「货卖给谁、批发价、分销跟单与收款」。

---

## 2. 两大订单承接

### 2.1 自营订单（self）

- 来源：OrderCore 中 `alloc_type=self_ship` 等自营履约订单（只读聚合 + 跟单备注）
- 用途：自营中心工作台查看、补录物流、对账入口
- MVP：列表骨架 + 外链 OrderCore 详情；深度履约仍在 OrderCore / ShippingCore

### 2.2 分销订单（distributor）

- 类似供应商订单，方向反转：我方是卖方
- 状态机（对齐采购）：`draft` → `confirmed` → `shipped` → `completed` / `cancelled`
- 收款状态：`unpaid` / `partial` / `paid`（由收款记录汇总）
- 履约类型：
  - `wholesale`：批发到分销商仓
  - `dropship`：代发到终端买家（分销商下单，我方直发）

---

## 3. 领域模型（MVP M1–M3）

### M1 · 分销商（类 VMS）

- `distributors`：编码、名称、联系人、账期、结算周期、收款账户（我方收款信息可挂租户，分销商侧记付款人）
- `distributor_addresses`：收货地址（对方仓 / 退货地址）
- `distributor_categories`：类别

### M1 · SKU 批发价

- `sku_distributor_prices`：SKU × 分销商 → `wholesale_price`、起订量、交期、是否支持代发

### M2 · 分销订单

- `dist_orders` / `dist_order_items`
- 从批发价带入单价；可关联 `ref_so_id` / `ref_order_no`（电商/分销渠道单）

### M3 · 物流 / 收款 / 附件

- `dist_shipments` / `dist_shipment_items`
- `dist_receipts`（分销商付款给我方）
- `dist_attachments`（销售单、付款截图等）

---

## 4. 菜单结构（Web）

```
工作台
自营订单（骨架，对接 OrderCore）
分销订单
  全部订单 / 代发订单 / 批发订单
分销商
  分销商信息 / SKU 批发价
```

---

## 5. 里程碑

| 阶段 | 内容 |
|------|------|
| M0 | 脚手架、UserCore 注册、端口/部署约定 |
| M1 | 分销商 CRUD、地址、批发价 |
| M2 | 分销订单 CRUD + 状态机 |
| M3 | 物流、收款、附件 |
| M4 | 自营订单聚合、与 OrderCore/ShippingCore 联动、结算合并 |

---

## 6. 非目标（本期不做）

- 分销商门户 / 小程序下单
- 多级分销佣金
- 采购入库/退回类逆向镜像（分销退货可后续单独立项）
