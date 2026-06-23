# 公域 × 私域联动与身份归并

> **公域**：抖店/淘宝/618 大促等平台订单（流量在活动、券在后端）。  
> **私域**：微信小程序门店/俱乐部（会员、服务、收款、推送）。  
>
> 目标：顾客 **在公域只下一单**，通过 **一次极轻的私域动作**，把平台订单与微信会员 **在系统内关联**，全域可见、可运营。

---

## 1. 核心结论（先说清楚）

| 问题 | 答案 |
|------|------|
| 公域下一单，能否自动和微信打通？ | **不能 100% 自动**。平台不会把买家的微信 openid 给你。 |
| 能否在系统里看到关联？ | **能**。一次 **绑定动作** 后，Commerce 订单 ↔ Store 会员 永久关联。 |
| 用户要操作几次？ | **公域下单 1 次** + **私域绑定 1 次**（扫码/授权手机号，通常 10 秒）。 |
| 618 大促场景 | 公域薅券下单 → 包裹/短信引导进小程序 → 授权手机号 → 自动匹配近期平台订单 → 关联成功。 |

**设计原则：公域负责成交，私域负责关系；系统用「统一会员 ID」把两边订单串起来。**

---

## 2. 顶层：统一会员（Customer Hub）

在 ProductCore / StoreHub / CommerceHub 之上，增加 **逻辑上的 Customer Hub**（可先落在 StoreHub，不必单独部署）：

```
                    ┌─────────────────────────┐
                    │   member_id  统一会员     │
                    │   （内部主键，全局唯一）   │
                    └────────────┬────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         ▼                       ▼                       ▼
  member_channel_bindings   commerce_orders          store_orders
  （身份绑定）               （公域平台单）            （私域单/收款单）
```

### 2.1 一张会员，多种身份

```
member_profiles
  ├── member_id
  ├── display_name
  ├── primary_phone         -- 归并主键（最重要）
  └── created_at

member_channel_bindings
  ├── member_id
  ├── channel_type          -- wx_mini | wx_union | phone | douyin | taobao | xhs | ...
  ├── channel_user_id       -- openid / unionid / 平台 buyer_id（若有）
  ├── verified              -- 是否用户确认
  ├── bound_at
  └── meta                  -- JSON
```

**微信互通靠：**
- 小程序 **openid**（本应用内）
- **unionid**（同一微信开放平台下，小程序 + 服务号 + 开放平台应用）
- **手机号**（用户授权后，与公域收货手机号匹配）

**公域平台身份靠：**
- Integration 拉单里的 **收货手机号**、**买家昵称**（辅助）
- 部分平台提供的 **buyer_id**（可存 binding，但用户不可见）

---

## 3. 公域订单如何挂上 member_id

CommerceHub 的 `orders` 增加：

```
orders
  ├── ... 原有平台字段 ...
  ├── member_id             -- 归并后填入（可空→已绑定）
  ├── link_status           -- unlinked | auto_matched | user_confirmed | manual
  └── linked_at
```

### 3.1 绑定方式（按推荐顺序）

| 方式 | 触发 | 置信度 | 618 适用 |
|------|------|--------|----------|
| **手机号自动匹配** | 小程序授权手机号 = 公域收货手机号 | 高（建议用户点确认） | ★★★ |
| **包裹卡扫码绑定** | 出库包裹内 QR → 小程序输入/确认订单 | 很高 | ★★★ |
| **订单号 + 手机验证** | 小程序手动填平台单号 + 短信验证码 | 很高 | ★★ |
| **发货短信链接** | 短信 URL Link 带 `?from=ship&order=xxx` | 高 | ★★★ |
| **客服人工绑定** | 后台搜平台单号绑 member | 人工 | ★ |
| **平台 buyer_id** | API 若有且用户曾在私域授权同一平台 | 中 | 视平台 |

**推荐主路径（618）：**

```
公域 618 下单（抖店/淘宝，用券）
    → CommerceHub 拉单入库，link_status=unlinked
    → 仓库发货，包裹放「扫码绑定会员·领私域权益」卡
    → 用户扫小程序码 → 授权手机号
    → 系统：phone 匹配最近 30 天 unlinked 订单
    → 展示「是否关联以下订单？」→ 用户确认
    → member_id 写入 orders，link_status=user_confirmed
    → 私域：积分/延保/门店服务预约 生效
```

用户感知：**我只在公域买了一次，扫个码就多了会员权益** — 不是让他再下一单。

---

## 4. 「一次下单」在系统里的含义

```
顾客视角：
  公域下单 ×1  +  私域扫码/授权 ×1（可选但强烈建议引导）

系统视角：
  commerce_order (SO)  ──link──►  member_id  ◄──bind──  wx openid + phone
                                      │
                                      ├── 历史公域订单（同 phone 批量归并）
                                      ├── 未来公域订单（自动匹配）
                                      └── 私域 store_order / collect_order / service_order
```

绑定 **一次** 后，同一手机号的新公域订单可 **自动** `auto_matched`（ configurable：是否仍需用户点确认）。

---

## 5. 618 大促联动玩法（产品层）

| 环节 | 公域 | 私域 | 系统 |
|------|------|------|------|
| 引流 | 平台活动页、大促券 | — | — |
| 成交 | 平台下单支付 | — | CommerceHub 落单 |
| 履约 | 平台物流 | — | WMS 发货 |
| 转私域 | 包裹卡 / 发货短信 | 小程序「绑定订单领权益」 | 订单↔会员关联 |
| 留存 | — | 会员积分、俱乐部、门店服务 | member 360 视图 |
| 复购 | 下次大促仍可在公域买 | 私域推送活动、预约到店 | 全域订单时间线 |

**权益示例（促进绑定）：**
- 绑定送俱乐部积分 / 延保
- 线下打包、调试预约优先
- 大促私域专属券（下次用，不跟平台券冲突）

公域继续用平台券做性价比；私域用 **服务和关系** 补位，不是硬导流再下一单。

---

## 6. 系统内如何「看到关联」

### 6.1 会员 360 视图（Store Admin / 统一后台）

```
会员：张三  138****1234  [微信已绑定]

【身份】
  小程序 openid  ✓   unionid  ✓   手机号  ✓
  抖店 buyer     —   淘宝 buyer  —（有则显示）

【公域订单】
  SO-618-001  抖店  ¥299  已发货  [618大促]  ← 已关联
  SO-517-008  淘宝  ¥150  已完成

【私域订单】
  COL-xxx  收款单  自行车打包  ¥140  已支付
  SVC-xxx  维修单  进行中

【时间线】
  06-18  抖店下单 SO-618-001
  06-20  包裹签收，扫码绑定会员
  06-22  小程序预约门店调试
```

### 6.2 Commerce 订单详情

```
平台订单 SO-618-001（抖店）
  关联会员：张三 (member_id: M10086)  [查看私域档案]
  绑定方式：包裹扫码 + 手机号确认
  绑定时间：2026-06-20 15:30
```

### 6.3 追溯

公域订单仍走 Commerce `trace_id`；绑定后 trace 时间线可展示 **私域后续服务**（可选扩展 `cross_domain_events`）。

---

## 7. 数据模型补充

```
order_member_links          -- 订单与会员关联审计（可选，与 orders.member_id 冗余）
  ├── order_id
  ├── member_id
  ├── link_method           -- phone_auto | qr_scan | manual | cs
  ├── link_confidence
  ├── confirmed_by_user
  └── linked_at

member_merge_logs           -- 账号合并日志（两 member 合并为一）
```

**合并规则：** 两个 member 若同一 phone + 同一 unionid 验证，合并为一个，历史订单全部迁移 `member_id`。

---

## 8. 隐私与确认

- **自动匹配手机号** 建议仍 **弹窗让用户确认**「是否关联您在抖店的订单 xxx」，避免错绑。
- 只展示 **脱敏** 信息（商品名、金额、尾号）。
- 绑定协议说明用途：售后、门店服务、会员权益。

---

## 9. 与各模块关系

```
ProductCore     -- 不变，货统一
CommerceHub     -- 公域订单 source；member_id 可空→已绑
StoreHub        -- 私域会员主地；小程序绑定入口；member 360
Integration     -- 拉单带收货 phone；不负责 openid
Customer Hub    -- member + bindings + 归并逻辑（逻辑模块）
```

**线上用户账号和微信互通** = `member_channel_bindings` 里 **phone + openid + unionid** 聚到同一 `member_id`，公域订单通过 **phone / 扫码** 挂到该 member。

---

## 10. 分阶段实施

| 阶段 | 内容 |
|------|------|
| **C1** | member + phone + 小程序 openid 注册 |
| **C2** | commerce orders 增加 member_id、link_status |
| **C3** | 小程序「绑定公域订单」页 + 手机号匹配 + 用户确认 |
| **C4** | 发货短信 / 包裹 QR、618 权益包 |
| **C5** | 会员 360、自动归并新公域单、服务号触达 |
| **C6** | 平台 buyer_id 绑定（若 API 支持） |

可与 StoreHub 收款单、CommerceHub 平台订单 **并行**；C1～C3 即可验证 618 场景 MVP。

---

## 11. 一句话总结

> **公域 618 下一单，私域扫码授权一次** — 用手机号 + 用户确认把平台订单挂到 `member_id`，微信 openid 与公域买家在同一系统内关联；之后公域买、私域服务，**一张会员档案、全域订单可见**。

参见 [ONLINE_OFFLINE.md](./ONLINE_OFFLINE.md)、[STORE_COLLECTION.md](./STORE_COLLECTION.md)。
