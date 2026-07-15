# WarehouseCore（仓储中心）设计说明

> 独立应用 **WarehouseCore** 作为 OSMS 平台库存管理中心（WMS + 初期 IMS）。  
> 与 [ROADMAP.md](./ROADMAP.md)、[SUPPLY_CHAIN.md](./SUPPLY_CHAIN.md) 对齐。

---

## 1. 决策

| 项 | 选择 |
|----|------|
| 商品关系 | **完全独立**：仓配父SKU/库存SKU 不强制关联 ProductCore；预留 `pim_spu_id` / `pim_sku_id` |
| 落地形态 | 独立仓库 `/home/asialeaf/projects/WarehouseCore` |
| Go module | `warehousecore` |
| API / Web | `:8095` / `:5180` |
| DB / MinIO | `warehousecore` / bucket `warehousecore` |
| UserCore | app code `warehousecore`，权限 `warehouse:read` / `warehouse:write` |
| 展示名 | 仓储中心 |

**原则：**

1. 仓配 SKU 编码租户内唯一，条码以 `sku_code` 为准
2. 库存只通过单据过账变动，禁止直接改结存
3. `stock_movements` 只追加不删改
4. 与 ProductCore / StoreCore / SupplyCore **弱耦合**
5. 一期不做批次 lot、渠道占用、拣货打包

---

## 2. 平台位置

```
UserCore (IAM) ──JWT──► WarehouseCore (:8095 / :5180)
                              │
         ┌────────────────────┼────────────────────┐
         ▼                    ▼                    ▼
   ProductCore           SupplyCore            StoreCore
   （后期映射）           （采购入库 GRN）        （调拨到店）
```

| 中心 | 职责边界 |
|------|----------|
| ProductCore | 销售/PIM 底库（本阶段不依赖） |
| WarehouseCore | 仓配主档 + 中心仓库存账 + 仓内单据 |
| StoreCore | 门店仓子集（`store_inventories`） |
| SupplyCore | 采购到货入库（二期对接） |

---

## 3. 菜单结构

> 商品管理对齐 [普源云ERP](https://erp.allroot.com/erp/main/index)：**商品信息**、**商品类别**、**包装规格**；跨境申报/店铺SKU/1688 等不在仓配一期范围。

```
仓储中心
├── 商品管理
│   ├── 商品信息（父SKU + 库存SKU）
│   ├── 商品类别
│   ├── 包装规格
│   ├── 组合品 / 组装品
│   └── 条码打印
├── 库存情况
│   ├── 库存查询
│   ├── 库存汇总账
│   ├── 库存明细表
│   └── 滞销查询
├── 仓库货位
│   ├── 仓库设置
│   └── 库位管理
├── 仓库盘点
│   ├── 仓库盘点单
│   └── 盘点明细表
├── 仓库调拨
│   └── 仓库调拨单
└── 其他出入库
    ├── 其他入库单
    └── 其他出库单
```

---

## 4. 数据模型

### 4.1 仓配主档

**`inv_categories`** — 仓配分类（本中心自建）

| 字段 | 说明 |
|------|------|
| tenant_id | |
| code / name | 编码、商品类别名称 |
| alias_cn / alias_en | 中文品名、英文品名 |
| parent_id | 上级分类 |
| sort / status | |

**`inv_pack_specs`** — 包装规格

| 字段 | 说明 |
|------|------|
| name | 包装规格名称，租户内唯一 |
| cost / weight_g | 成本价、重量(g) |
| remark / status | |

**`inv_pack_spec_skus`** — 包装规格绑定库存SKU（数量范围）

**`inv_products`** — 父SKU / 主SKU

| 字段 | 说明 |
|------|------|
| tenant_id | |
| parent_sku | 父SKU 编码，租户内唯一 |
| name | 商品名称 |
| category_id | 仓配分类 |
| pack_spec_id | 外包装规格 |
| developed_at | 开发日期 |
| default_warehouse_id | 默认发货仓库 |
| score_factor | 分值系数 |
| remark / pic / album_pics | |
| status | 启用/停用 |
| pim_spu_id | 可空，后期映射 |

**`inv_skus`** — 库存SKU

| 字段 | 说明 |
|------|------|
| parent_id | 所属父SKU |
| sku_code | 库存SKU，租户内唯一（条码主码） |
| pic / status | SKU 图片、状态 |
| product_type | 结构：`normal` / `combo` / `assembly` |
| goods_kind | 商品类型：`normal`普通 / `packaging`包材 / `accessory`配件 / `gift`赠品 |
| pick_name | 配货名称 |
| style1 / style2 / style3 | 款式 |
| weight_g | 重量（克） |
| last_purchase_price / min_purchase_price / retail_price | |
| description / upc / asin / supplier_item_no | |
| pim_sku_id | 可空，后期映射 |

父商品另含：物流报关（申报名/重量/价值/原产国/海关编码）、采购（渠道/采购员/最低价/MOQ）、**多供应商**（挂在父商品，供应商主数据来自 SupplyCore VMS）、包装尺寸与成本、销售属性与多档售价、商品主图。

侧栏「商品」对齐普源：商品管理（信息/类别/包装规格）、商品明细（组装/组合/店铺SKU）、条码打印、其它（商品费用设置 / 重量检测 / 利润试算）。

**`inv_bom_headers` / `inv_bom_items`** — 组合/组装 BOM

| 类型 | 含义 | 库存行为 |
|------|------|----------|
| combo | 虚拟捆绑 | 出库展开扣子件 |
| assembly | 加工成品 | 按成品出入库；加工领料二期 |

### 4.2 仓库货位

**`warehouses`**：code、name、type(`central`/`return`/`transit`)、地址、联系人、status、is_default

**`warehouse_locations`**：warehouse_id、code、zone/aisle/shelf/bin、status；无库位时用虚拟库位 `DEFAULT`

### 4.3 库存账

**`inv_balances`**：tenant × warehouse × location × inv_sku → `on_hand`

**`stock_movements`**：只追加；qty、direction、doc_type、doc_no、balance_after、ref_doc_type/ref_doc_id

变动类型：`other_in` / `other_out` / `transfer_in` / `transfer_out` / `stocktake_gain` / `stocktake_loss` / `purchase_in`（预留）/ `sale_out`（预留）

### 4.4 单据

| 表 | 状态机 |
|----|--------|
| `other_inbound_orders` / items | draft → posted / cancelled |
| `other_outbound_orders` / items | draft → posted / cancelled |
| `stocktake_orders` / items | draft → counting → review → posted / cancelled |
| `transfer_orders` / items | draft → in_transit → received / cancelled |

---

## 5. 分阶段里程碑

| 阶段 | 内容 | 交付标准 |
|------|------|----------|
| **M0** | 空应用 + UserCore/deploy + 登录 | 应用中心可打开仓储中心 |
| **M1** | 分类、父SKU/库存SKU、BOM、仓库/库位、条码打印 | 可维护仓配商品与仓位 |
| **M2** | balances/movements、其他入/出、四类库存查询 | 手工出入库后账实可查 |
| **M3** | 盘点单/明细、调拨单过账 | 仓内调拨与盘点闭环 |
| **M4** | PIM 映射 API、采购入库预留、调拨到店预留 | 跨中心接口就绪 |

---

## 6. 工程信息

| 项目 | 值 |
|------|-----|
| 代码仓库 | `/home/asialeaf/projects/WarehouseCore` |
| Go module | `warehousecore` |
| Docker 镜像 | `warehousecore-api`、`warehousecore-web` |
| UserCore 应用码 | `warehousecore` |
| 权限 | `warehouse:read` / `warehouse:write` |
| 平台编排 | `/home/asialeaf/projects/deploy` |
