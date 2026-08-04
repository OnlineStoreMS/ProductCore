export interface Brand {
  id: number
  name: string
  logo?: string
  firstLetter: string
  sort: number
  showStatus: 0 | 1
  productCount: number
}

export interface Category {
  id: number
  parentId: number
  name: string
  level: number
  productCount: number
  sort: number
  showStatus: 0 | 1
  children?: Category[]
}

export interface ProductGroup {
  id: number
  parentId: number
  name: string
  description: string
  level: number
  productCount: number
  sort: number
  createTime: string
  children?: ProductGroup[]
}

export interface ProductKeyword {
  id: number
  name: string
  description: string
  sort: number
  productCount: number
  createTime: string
}

/** 规格值：文本 + 备注 + 图片（第一组规格） */
export interface SkuSpecValue {
  value: string
  remark?: string
  pic?: string
}

export interface SkuSpec {
  name: string
  values: SkuSpecValue[]
}

export interface SkuItem {
  id?: number
  skuCode: string
  specs: Record<string, string>
  price: number
  costPrice: number
  stock: number
  pic?: string
  marketPrice?: number
  weight?: number
}

/** SKU 管理轻量视图（不含详情、媒体等大字段） */
export interface ProductSkus {
  id: number
  name: string
  materialCode: string
  skuCount: number
  price: number
  originalPrice: number
  stock: number
  skuSpecs: SkuSpec[]
  skus: SkuItem[]
}

/** SPU 扩展媒体（3:4 主图、多比例视频、素材图、详情图） */
export interface ProductMedia {
  pics34?: string[]
  videos?: {
    ratio11?: string
    ratio34?: string
    ratio169?: string
    ratio916?: string
  }
  materials?: {
    white?: string
    transparent?: string
    guide34?: string
    long?: string
  }
  detailPics?: string[]
}

export interface ListedShop {
  shopId: number
  shopName: string
  platformTypeName?: string
  platformTypeLogo?: string
  sourceChannel?: string
}

export interface Product {
  id: number
  name: string
  subTitle: string
  materialCode: string
  source: string
  productSn: string
  brandId: number
  brandName: string
  categoryId: number
  categoryName: string
  groupIds: number[]
  keywordIds: number[]
  pic: string
  albumPics: string[]
  productVideo?: string
  media?: ProductMedia
  price: number
  originalPrice: number
  stock: number
  unit: string
  weight: number
  publishStatus: 0 | 1
  isDraft?: 0 | 1
  hasEditDraft?: boolean
  draftSavedAt?: string
  verifyStatus: 0 | 1
  sort: number
  sale: number
  skuCount?: number
  listedShops?: ListedShop[]
  listedShopCount?: number
  description: string
  detailHtml: string
  skuSpecs: SkuSpec[]
  skus: SkuItem[]
  createTime: string
  updateTime: string
  deleteTime?: string
}

export type ProductForm = Omit<Product, 'id' | 'brandName' | 'categoryName' | 'createTime' | 'updateTime' | 'sale'> & {
  id?: number
}

export function emptySpecValue(): SkuSpecValue {
  return { value: '' }
}

/** 编辑页无 SKU / 无已存规格时的默认空框（仅 UI，不入库） */
export function defaultSkuSpecEditorState(): SkuSpec[] {
  return [{ name: '规格', values: [emptySpecValue()] }]
}

function isBlankSpecValue(v: SkuSpecValue): boolean {
  return !v.value.trim() && !v.remark?.trim() && !v.pic
}

/** 压缩规格结构：去掉空白项，无有效内容时返回 []（用于持久化） */
export function compactSkuSpecs(specs: SkuSpec[]): SkuSpec[] {
  return specs
    .map((s) => ({
      name: s.name.trim(),
      values: (s.values || []).filter((v) => !isBlankSpecValue(v)),
    }))
    .filter((s) => s.name && s.values.length > 0)
}

/** 校验同一规格项下是否存在重复规格值，返回错误文案或 null */
export function validateSkuSpecsNoDuplicateValues(specs: SkuSpec[]): string | null {
  for (const spec of specs) {
    const name = spec.name.trim() || '规格'
    const seen = new Set<string>()
    for (const v of spec.values || []) {
      const val = v.value.trim()
      if (!val) continue
      if (seen.has(val)) {
        return `规格「${name}」存在重复规格值「${val}」`
      }
      seen.add(val)
    }
  }
  return null
}

/** 返回存在重复规格名的规格项下标（用于红框提示） */
export function duplicateSpecNameIndexes(specs: SkuSpec[]): Set<number> {
  const byName = new Map<string, number[]>()
  specs.forEach((spec, index) => {
    const name = spec.name.trim()
    if (!name) return
    const list = byName.get(name) ?? []
    list.push(index)
    byName.set(name, list)
  })
  const duplicated = new Set<number>()
  for (const indexes of byName.values()) {
    if (indexes.length > 1) {
      indexes.forEach((i) => duplicated.add(i))
    }
  }
  return duplicated
}

/** 返回各规格项内重复规格值的下标（specIndex -> valueIndexes） */
export function duplicateSpecValueIndexes(specs: SkuSpec[]): Map<number, Set<number>> {
  const result = new Map<number, Set<number>>()
  specs.forEach((spec, specIndex) => {
    const seen = new Map<string, number>()
    const dups = new Set<number>()
    spec.values.forEach((v, vi) => {
      const val = v.value.trim()
      if (!val) return
      if (seen.has(val)) {
        dups.add(seen.get(val)!)
        dups.add(vi)
      } else {
        seen.set(val, vi)
      }
    })
    if (dups.size > 0) result.set(specIndex, dups)
  })
  return result
}

/** 校验是否存在重复规格名，返回错误文案或 null */
export function validateSkuSpecsNoDuplicateNames(specs: SkuSpec[]): string | null {
  const byName = new Map<string, number>()
  for (const spec of specs) {
    const name = spec.name.trim()
    if (!name) continue
    if (byName.has(name)) {
      return `规格名「${name}」重复，请修改`
    }
    byName.set(name, 1)
  }
  return null
}

/** 加载编辑页：库中无规格时展示默认空框，有则还原已存规格 */
export function normalizeSkuSpecsForEditor(specs: SkuSpec[] | undefined): SkuSpec[] {
  if (!specs?.length) return defaultSkuSpecEditorState()
  const normalized = specs.map((s) => ({
    name: s.name,
    values: (s.values || []).map((v) =>
      typeof v === 'string' ? { value: v } : { value: v.value || '', remark: v.remark, pic: v.pic },
    ),
  }))
  const compact = compactSkuSpecs(normalized)
  return compact.length > 0 ? compact : defaultSkuSpecEditorState()
}

/** @deprecated 使用 normalizeSkuSpecsForEditor */
export function normalizeSkuSpecs(specs: SkuSpec[] | undefined): SkuSpec[] {
  return normalizeSkuSpecsForEditor(specs)
}

/** SPU 展示重量：取销售价最低 SKU 的重量；若无则回退 SPU 已存 weight */
export function spuWeightFromSkus(skus: SkuItem[] | undefined, fallback = 0): number {
  if (!skus?.length) return fallback > 0 ? fallback : 0
  let minPrice = 0
  let weight = 0
  for (const sku of skus) {
    const price = sku.price || 0
    if (price <= 0) continue
    const skuWeight = sku.weight || 0
    if (minPrice === 0 || price < minPrice) {
      minPrice = price
      weight = skuWeight
    } else if (price === minPrice && weight <= 0 && skuWeight > 0) {
      weight = skuWeight
    }
  }
  if (weight > 0) return weight
  return fallback > 0 ? fallback : 0
}
