<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { View } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { fetchGroups, fetchProduct } from '../../api/product'
import type { Product, ProductGroup, SkuItem, SkuSpec } from '../../types/product'
import { spuWeightFromSkus } from '../../types/product'
import { colWidthFromTexts, priceColWidth } from '../../utils/tableColWidth'

const props = defineProps<{
  modelValue: boolean
  productId?: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const loading = ref(false)
const product = ref<Product | null>(null)
const groups = ref<ProductGroup[]>([])

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const groupNames = computed(() => {
  if (!product.value?.groupIds?.length) return '-'
  const map = new Map(groups.value.map((g) => [g.id, g.name]))
  return product.value.groupIds.map((id) => map.get(id) || `#${id}`).join('、') || '-'
})

const displayWeight = computed(() => {
  if (!product.value) return null
  const w = spuWeightFromSkus(product.value.skus, product.value.weight)
  return w > 0 ? w : null
})

const albumPics = computed(() => {
  if (!product.value) return []
  const pics = [product.value.pic, ...(product.value.albumPics || [])].filter(Boolean)
  return [...new Set(pics)]
})

const specColumns = computed(() => {
  const names = new Set<string>()
  for (const spec of product.value?.skuSpecs || []) {
    const name = spec.name?.trim()
    if (name) names.add(name)
  }
  for (const sku of product.value?.skus || []) {
    for (const key of Object.keys(sku.specs || {})) {
      if (key.trim()) names.add(key.trim())
    }
  }
  return [...names]
})

const specColumnWidths = computed(() => {
  const specs = product.value?.skuSpecs || []
  const skus = product.value?.skus || []
  const widths: Record<string, number> = {}
  for (const name of specColumns.value) {
    const cells = skus.map((sku) => formatSpecDisplay(specs, name, sku.specs[name]))
    widths[name] = colWidthFromTexts(name, cells)
  }
  return widths
})

const skuCodeColWidth = computed(() => {
  const skus = product.value?.skus || []
  return colWidthFromTexts('规格编码', skus.map((s) => s.skuCode || '-'), 88, 160)
})

const salePriceColWidth = computed(() =>
  priceColWidth('销售价', (product.value?.skus || []).map((s) => s.price)),
)

const marketPriceColWidth = computed(() =>
  priceColWidth('市场价', (product.value?.skus || []).map((s) => s.marketPrice)),
)

const costPriceColWidth = computed(() =>
  priceColWidth('成本价', (product.value?.skus || []).map((s) => s.costPrice)),
)

function getSpecRemark(specs: SkuSpec[], specName: string, value?: string) {
  if (!value) return ''
  const spec = specs.find((s) => s.name.trim() === specName)
  return spec?.values.find((v) => v.value === value)?.remark?.trim() || ''
}

function formatSpecDisplay(specs: SkuSpec[], specName: string, value?: string) {
  if (!value) return '-'
  const remark = getSpecRemark(specs, specName, value)
  return remark ? `${value}(${remark})` : value
}

function formatSkuSpecsText(sku: SkuItem, specs: SkuSpec[]) {
  const parts = specColumns.value.map((name) => formatSpecDisplay(specs, name, sku.specs[name]))
  return parts.filter((p) => p !== '-').join(' / ') || '-'
}

function resolveSkuPic(sku: SkuItem): string {
  if (sku.pic?.trim()) return sku.pic.trim()
  const specs = product.value?.skuSpecs || []
  const first = specs[0]
  if (!first?.name) return ''
  const val = sku.specs[first.name]
  return first.values.find((v) => v.value === val)?.pic?.trim() || ''
}

const skuPreviewPic = ref('')
const skuPreviewVisible = ref(false)

function openSkuPicPreview(url: string) {
  skuPreviewPic.value = url
  skuPreviewVisible.value = true
}

function skuRowClassName({ row }: { row: SkuItem }) {
  return (row.stock ?? 0) === 0 ? 'sku-row-zero-stock' : ''
}

async function loadDetail() {
  if (!props.productId) return
  loading.value = true
  product.value = null
  try {
    const [p, g] = await Promise.all([
      fetchProduct(props.productId),
      groups.value.length ? Promise.resolve(groups.value) : fetchGroups(),
    ])
    product.value = p
    if (!groups.value.length) groups.value = g
  } catch (e) {
    ElMessage.error((e as Error).message || '加载商品详情失败')
    visible.value = false
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.modelValue, props.productId] as const,
  ([open, id]) => {
    if (open && id) loadDetail()
  },
)
</script>

<template>
  <el-drawer
    v-model="visible"
    :title="product?.name || '商品详情'"
    size="860px"
    destroy-on-close
    class="product-detail-drawer"
  >
    <div v-loading="loading" class="drawer-body">
      <template v-if="product">
        <section class="detail-section">
          <h4 class="section-title">基础信息</h4>
          <div class="basic-head">
            <el-image
              v-if="product.pic"
              :src="product.pic"
              class="main-pic"
              fit="cover"
              :preview-src-list="albumPics"
              preview-teleported
              hide-on-click-modal
            />
            <div v-else class="main-pic main-pic-empty">无图</div>
            <el-descriptions :column="2" border size="small" class="basic-desc">
              <el-descriptions-item label="商品标题" :span="2">{{ product.name }}</el-descriptions-item>
              <el-descriptions-item label="货号">{{ product.productSn || '-' }}</el-descriptions-item>
              <el-descriptions-item label="资料编码">{{ product.materialCode || '-' }}</el-descriptions-item>
              <el-descriptions-item label="商品ID">{{ product.id }}</el-descriptions-item>
              <el-descriptions-item label="商品来源">{{ product.source || '-' }}</el-descriptions-item>
              <el-descriptions-item label="导购短标题" :span="2">{{ product.subTitle || '-' }}</el-descriptions-item>
              <el-descriptions-item label="品牌">{{ product.brandName || '-' }}</el-descriptions-item>
              <el-descriptions-item label="分类">{{ product.categoryName || '-' }}</el-descriptions-item>
              <el-descriptions-item label="分组" :span="2">{{ groupNames }}</el-descriptions-item>
              <el-descriptions-item label="单位">{{ product.unit || '-' }}</el-descriptions-item>
              <el-descriptions-item label="排序">{{ product.sort ?? 0 }}</el-descriptions-item>
              <el-descriptions-item label="上架状态">
                <el-tag :type="product.publishStatus ? 'success' : 'info'" size="small">
                  {{ product.publishStatus ? '已上架' : '已下架' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="销量">{{ product.sale ?? 0 }}</el-descriptions-item>
              <el-descriptions-item label="SPU 销售价">¥{{ product.price }}</el-descriptions-item>
              <el-descriptions-item label="SPU 市场价">¥{{ product.originalPrice || '-' }}</el-descriptions-item>
              <el-descriptions-item label="SPU 库存">{{ product.stock }}</el-descriptions-item>
              <el-descriptions-item label="重量(g)">{{ displayWeight ?? '-' }}</el-descriptions-item>
              <el-descriptions-item v-if="product.description" label="商品描述" :span="2">
                {{ product.description }}
              </el-descriptions-item>
            </el-descriptions>
          </div>
        </section>

        <section v-if="albumPics.length > 1" class="detail-section">
          <h4 class="section-title">主图相册</h4>
          <div class="album-list">
            <el-image
              v-for="(url, i) in albumPics"
              :key="url + i"
              :src="url"
              class="album-thumb"
              fit="cover"
              :preview-src-list="albumPics"
              :initial-index="i"
              preview-teleported
              hide-on-click-modal
            />
          </div>
        </section>

        <section v-if="product.skuSpecs?.length" class="detail-section">
          <h4 class="section-title">规格定义</h4>
          <div v-for="(spec, si) in product.skuSpecs" :key="si" class="spec-block">
            <div class="spec-name">{{ spec.name }}</div>
            <div class="spec-values">
              <div v-for="(sv, vi) in spec.values" :key="vi" class="spec-value-item">
                <el-image
                  v-if="si === 0 && sv.pic"
                  :src="sv.pic"
                  class="spec-value-pic"
                  fit="cover"
                  :preview-src-list="[sv.pic]"
                  preview-teleported
                />
                <span class="spec-value-text">
                  {{ sv.value }}
                  <span v-if="sv.remark" class="spec-value-remark">({{ sv.remark }})</span>
                </span>
              </div>
            </div>
          </div>
        </section>

        <section class="detail-section">
          <h4 class="section-title">
            SKU 明细
            <el-tag size="small" type="info">{{ product.skus?.length || 0 }} 条</el-tag>
          </h4>
          <div v-if="product.skus?.length" class="sku-table-wrap">
            <el-table
              :data="product.skus"
              border
              stripe
              size="small"
              class="sku-table"
              :row-class-name="skuRowClassName"
            >
            <el-table-column label="预览" width="56" align="center" class-name="col-sku-preview">
              <template #default="{ row }">
                <div v-if="resolveSkuPic(row)" class="sku-preview">
                  <el-image :src="resolveSkuPic(row)" fit="cover" class="sku-pic" />
                  <button
                    type="button"
                    class="preview-eye-btn"
                    title="预览"
                    @click.stop="openSkuPicPreview(resolveSkuPic(row))"
                  >
                    <el-icon><View /></el-icon>
                  </button>
                </div>
                <span v-else class="sku-pic-empty">-</span>
              </template>
            </el-table-column>
            <el-table-column
              v-for="name in specColumns"
              :key="name"
              :label="name"
              :min-width="specColumnWidths[name]"
              show-overflow-tooltip
            >
              <template #default="{ row }">
                {{ formatSpecDisplay(product!.skuSpecs, name, row.specs[name]) }}
              </template>
            </el-table-column>
            <el-table-column v-if="!specColumns.length" label="规格" :min-width="120" show-overflow-tooltip>
              <template #default="{ row }">
                {{ formatSkuSpecsText(row, product!.skuSpecs) }}
              </template>
            </el-table-column>
            <el-table-column
              prop="skuCode"
              label="规格编码"
              :min-width="skuCodeColWidth"
              show-overflow-tooltip
            />
            <el-table-column
              label="销售价"
              :width="salePriceColWidth"
              align="right"
              class-name="col-price"
            >
              <template #default="{ row }">¥{{ row.price }}</template>
            </el-table-column>
            <el-table-column
              label="市场价"
              :width="marketPriceColWidth"
              align="right"
              class-name="col-price"
            >
              <template #default="{ row }">{{ row.marketPrice ? `¥${row.marketPrice}` : '-' }}</template>
            </el-table-column>
            <el-table-column
              label="成本价"
              :width="costPriceColWidth"
              align="right"
              class-name="col-price"
            >
              <template #default="{ row }">{{ row.costPrice ? `¥${row.costPrice}` : '-' }}</template>
            </el-table-column>
            <el-table-column label="库存" :min-width="64" align="center" class-name="col-stock">
              <template #default="{ row }">
                <span :class="{ 'sku-stock-zero': (row.stock ?? 0) === 0 }">{{ row.stock ?? 0 }}</span>
              </template>
            </el-table-column>
            <el-table-column label="重量(g)" :min-width="72" align="center">
              <template #default="{ row }">{{ row.weight ?? '-' }}</template>
            </el-table-column>
            </el-table>
          </div>
          <el-empty v-else description="暂无 SKU" :image-size="72" />
        </section>
      </template>
    </div>
    <el-image-viewer
      v-if="skuPreviewVisible"
      :url-list="[skuPreviewPic]"
      teleported
      hide-on-click-modal
      @close="skuPreviewVisible = false"
    />
  </el-drawer>
</template>

<style scoped>
.drawer-body {
  min-height: 200px;
  padding-bottom: 24px;
}

.detail-section + .detail-section {
  margin-top: 24px;
}

.section-title {
  margin: 0 0 12px;
  font-size: 15px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  gap: 8px;
}

.basic-head {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.main-pic {
  width: 96px;
  height: 96px;
  border-radius: 6px;
  flex-shrink: 0;
}

.main-pic-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  color: #909399;
  font-size: 13px;
}

.basic-desc {
  flex: 1;
  min-width: 0;
}

.album-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.album-thumb {
  width: 72px;
  height: 72px;
  border-radius: 4px;
}

.spec-block {
  background: #fafafa;
  border-radius: 4px;
  padding: 10px 12px;
}

.spec-block + .spec-block {
  margin-top: 8px;
}

.spec-name {
  font-weight: 500;
  margin-bottom: 8px;
  color: rgba(0, 0, 0, 0.85);
}

.spec-values {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.spec-value-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  font-size: 13px;
}

.spec-value-pic {
  width: 28px;
  height: 28px;
  border-radius: 2px;
  flex-shrink: 0;
}

.spec-value-remark {
  color: #909399;
}

.sku-table-wrap {
  width: 100%;
  overflow-x: auto;
  border-radius: 4px;
}

.sku-table {
  width: 100%;
}

.sku-table :deep(.el-table__body),
.sku-table :deep(.el-table__header) {
  table-layout: auto;
}

.sku-table :deep(.el-table__cell .cell) {
  white-space: nowrap;
}

.sku-table :deep(.col-sku-preview .cell) {
  overflow: visible;
  text-overflow: clip;
  line-height: 0;
}

.sku-table :deep(.col-price .cell) {
  padding-left: 4px;
  padding-right: 6px;
  font-size: 13px;
}

.sku-table :deep(.sku-row-zero-stock > td.el-table__cell) {
  color: var(--el-color-danger);
}

.sku-table :deep(.sku-row-zero-stock:hover > td.el-table__cell) {
  color: var(--el-color-danger);
}

.sku-stock-zero {
  color: var(--el-color-danger);
  font-weight: 500;
}

.sku-preview {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 40px;
  vertical-align: middle;
}

.sku-preview .preview-eye-btn {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.15s;
}

.sku-preview:hover .preview-eye-btn {
  opacity: 1;
}

.sku-preview .preview-eye-btn .el-icon {
  font-size: 14px;
}

.sku-pic {
  display: block;
  width: 40px;
  height: 40px;
  border-radius: 4px;
}

.sku-pic :deep(.el-image__inner) {
  display: block;
  width: 40px;
  height: 40px;
  object-fit: cover;
}

.sku-pic-empty {
  color: #c0c4cc;
  line-height: 40px;
}
</style>

<style>
/* 抽屉固定适中宽度，小屏不溢出 */
.product-detail-drawer.el-drawer {
  max-width: calc(100vw - 32px);
}
</style>
