<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Upload, Download, View } from '@element-plus/icons-vue'
import { fetchProductSkus, updateProductSkus } from '../../api/product'
import { uploadImageFromUrl } from '../../api/upload'
import type { ProductSkus, SkuItem, SkuSpec } from '../../types/product'
import { isValidSkuCode, sanitizeSkuCode } from '../../utils/skuCode'
import {
  applySkuFileImport,
  downloadSkuExport,
  parseSkuFile,
  type SkuExportFormat,
} from '../../utils/skuCsvImport'
import { colWidthFromTexts } from '../../utils/tableColWidth'

const props = defineProps<{
  modelValue: boolean
  productId?: number
  productName?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  saved: [product: ProductSkus]
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const loading = ref(false)
const saving = ref(false)
const product = ref<ProductSkus | null>(null)
const skus = ref<SkuItem[]>([])
const keyword = ref('')
const batchPrice = ref<number>()
const batchStock = ref<number>()
const csvInputRef = ref<HTMLInputElement>()
const importing = ref(false)

const materialCode = computed(() => product.value?.materialCode?.trim() || '')

const displayTitle = computed(() => props.productName || product.value?.name || '')

const specColumns = computed(() => {
  const names = new Set<string>()
  for (const spec of product.value?.skuSpecs || []) {
    const name = spec.name?.trim()
    if (name) names.add(name)
  }
  for (const sku of skus.value) {
    for (const key of Object.keys(sku.specs || {})) {
      if (key.trim()) names.add(key.trim())
    }
  }
  return [...names]
})

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

const specColumnWidths = computed(() => {
  const specs = product.value?.skuSpecs || []
  const widths: Record<string, number> = {}
  for (const name of specColumns.value) {
    const cells = skus.value.map((sku) => formatSpecDisplay(specs, name, sku.specs[name]))
    widths[name] = colWidthFromTexts(name, cells)
  }
  return widths
})

const skuCodeColWidth = computed(() =>
  colWidthFromTexts('编码', skus.value.map((s) => s.skuCode || '-'), 72, 140),
)

const dialogWidth = computed(() => {
  const specSum = specColumns.value.reduce(
    (sum, name) => sum + (specColumnWidths.value[name] || 80),
    0,
  )
  const base = 64 + skuCodeColWidth.value + 120 + 100 + 48
  return `${Math.min(Math.max(720, base + specSum), 960)}px`
})

function formatSpecLabel(specs: SkuSpec[], sku: SkuItem): string {
  if (specColumns.value.length) {
    const parts = specColumns.value.map((name) =>
      formatSpecDisplay(specs, name, sku.specs[name]),
    )
    return parts.filter((p) => p !== '-').join(' / ') || '默认规格'
  }
  const fallback = Object.values(sku.specs || {}).filter(Boolean)
  return fallback.length ? fallback.join(' / ') : '默认规格'
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

const filteredSkus = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  if (!q) return skus.value
  const specs = product.value?.skuSpecs || []
  return skus.value.filter((sku) => {
    if (sku.skuCode?.toLowerCase().includes(q)) return true
    return formatSpecLabel(specs, sku).toLowerCase().includes(q)
  })
})

function validateSkus(): string | null {
  if (!skus.value.length) return null
  const codes = new Set<string>()
  for (const sku of skus.value) {
    const code = sanitizeSkuCode(sku.skuCode || '')
    if (!code) return '存在未填写规格编码的 SKU'
    if (!isValidSkuCode(code)) return '规格编码格式不正确'
    if (codes.has(code)) return '规格编码不能重复'
    codes.add(code)
  }
  return null
}

async function loadProduct() {
  if (!props.productId) return
  loading.value = true
  product.value = null
  skus.value = []
  keyword.value = ''
  batchPrice.value = undefined
  batchStock.value = undefined
  try {
    const p = await fetchProductSkus(props.productId)
    product.value = p
    skus.value = JSON.parse(JSON.stringify(p.skus || []))
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
    visible.value = false
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  const err = validateSkus()
  if (err) {
    ElMessage.warning(err)
    return
  }
  if (!props.productId) return
  saving.value = true
  try {
    const payload = skus.value.map((s) => ({
      ...s,
      skuCode: sanitizeSkuCode(s.skuCode),
    }))
    const updated = await updateProductSkus(
      props.productId,
      payload,
      product.value?.skuSpecs,
    )
    ElMessage.success('已保存')
    emit('saved', updated)
    visible.value = false
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  } finally {
    saving.value = false
  }
}

function applyBatch() {
  const targets = filteredSkus.value
  if (!targets.length) return
  const hasPrice = batchPrice.value !== undefined && batchPrice.value !== null
  const hasStock = batchStock.value !== undefined && batchStock.value !== null
  if (!hasPrice && !hasStock) {
    ElMessage.info('请填写要批量修改的销售价或库存')
    return
  }
  const targetSet = new Set(targets)
  skus.value = skus.value.map((sku) => {
    if (!targetSet.has(sku)) return sku
    return {
      ...sku,
      ...(hasPrice ? { price: batchPrice.value! } : {}),
      ...(hasStock ? { stock: batchStock.value! } : {}),
    }
  })
  ElMessage.success(`已更新 ${targets.length} 条`)
}

function triggerCsvImport() {
  if (!props.productId) return
  if (!materialCode.value) {
    ElMessage.warning('当前商品缺少资料编码，无法导入')
    return
  }
  csvInputRef.value?.click()
}

function exportSku(format: SkuExportFormat | string) {
  const fmt = (format === 'csv' ? 'csv' : 'xlsx') as SkuExportFormat
  if (!props.productId || !skus.value.length) {
    ElMessage.warning('暂无 SKU 可导出')
    return
  }
  if (!materialCode.value) {
    ElMessage.warning('当前商品缺少资料编码，无法导出')
    return
  }
  try {
    downloadSkuExport(
      {
        productId: props.productId,
        materialCode: materialCode.value,
        skus: skus.value,
        specColumns: specColumns.value,
        resolvePic: resolveSkuPic,
      },
      fmt,
    )
  } catch (e) {
    ElMessage.error((e as Error).message || '导出失败')
    return
  }
  ElMessage.success(`已导出 ${skus.value.length} 条 SKU（${fmt === 'xlsx' ? 'Excel' : 'CSV'}）`)
}

async function onFileSelected(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || !props.productId) return
  importing.value = true
  try {
    const rows = await parseSkuFile(file)
    const result = await applySkuFileImport(
      rows,
      materialCode.value,
      async (sourceUrl) =>
        uploadImageFromUrl(sourceUrl, {
          scope: 'product',
          productId: props.productId,
          resource: 'spec',
        }),
    )
    skus.value = result.skus
    if (product.value) {
      product.value.skuSpecs = JSON.parse(JSON.stringify(result.skuSpecs))
    }
    const parts = [`${result.imported} 条 SKU`]
    if (result.picUpdated) parts.push(`图片 ${result.picUpdated} 张`)
    ElMessage.success(`已导入并重建：${parts.join('、')}（规格编码已重新生成，请确认后保存）`)
  } catch (err) {
    ElMessage.error((err as Error).message || '导入失败')
  } finally {
    importing.value = false
  }
}

watch(
  () => [props.modelValue, props.productId] as const,
  ([open, id]) => {
    if (open && id) loadProduct()
  },
)
</script>

<template>
  <el-dialog
    v-model="visible"
    :width="dialogWidth"
    destroy-on-close
    class="sku-quick-dialog"
    :close-on-click-modal="false"
    :show-close="true"
  >
    <template #header>
      <div class="dialog-header">
        <span class="dialog-header-label">SKU列表</span>
        <span v-if="displayTitle" class="dialog-header-divider">·</span>
        <span v-if="displayTitle" class="dialog-header-title" :title="displayTitle">
          {{ displayTitle }}
        </span>
        <span v-if="materialCode" class="dialog-header-meta">资料编码 {{ materialCode }}</span>
      </div>
    </template>

    <input
      ref="csvInputRef"
      type="file"
      accept=".csv,.xlsx,.xls"
      class="csv-input-hidden"
      @change="onFileSelected"
    />

    <div v-loading="loading" class="dialog-body">
      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="搜索规格 / 编码"
          clearable
          :prefix-icon="Search"
          class="search-input"
        />
        <el-dropdown
          split-button
          :disabled="!skus.length || !productId || !materialCode"
          @click="exportSku('xlsx')"
          @command="exportSku"
        >
          <template #default>
            <el-icon class="el-icon--left"><Download /></el-icon>
            导出 SKU
          </template>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="xlsx">导出 Excel (.xlsx)</el-dropdown-item>
              <el-dropdown-item command="csv">导出 CSV (.csv)</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button
          :icon="Upload"
          :loading="importing"
          :disabled="!productId || !materialCode"
          @click="triggerCsvImport"
        >
          导入 SKU
        </el-button>
        <span class="sku-count">共 {{ skus.length }} 条</span>
      </div>

      <div v-if="skus.length" class="batch-bar">
        <el-input v-model.number="batchPrice" placeholder="销售价" class="batch-input">
          <template #suffix>元</template>
        </el-input>
        <el-input v-model.number="batchStock" placeholder="库存" class="batch-input">
          <template #suffix>件</template>
        </el-input>
        <el-button type="primary" class="batch-btn" @click="applyBatch">
          批量填充{{ keyword.trim() ? '（当前筛选）' : '' }}
        </el-button>
      </div>

      <div v-if="skus.length" class="sku-table-wrap">
        <el-table
          :data="filteredSkus"
          border
          size="small"
          class="sku-table"
          max-height="420"
        >
          <el-table-column label="图片" width="64" align="center" class-name="col-sku-preview">
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
          <el-table-column
            v-if="!specColumns.length"
            label="规格"
            :min-width="120"
            show-overflow-tooltip
          >
            <template #default="{ row }">
              {{ formatSpecLabel(product!.skuSpecs, row) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="skuCode"
            label="编码"
            :min-width="skuCodeColWidth"
            show-overflow-tooltip
          />
          <el-table-column label="销售价" width="120">
            <template #header><span class="col-required">销售价</span></template>
            <template #default="{ row }">
              <el-input v-model.number="row.price" placeholder="请输入">
                <template #suffix>元</template>
              </el-input>
            </template>
          </el-table-column>
          <el-table-column label="库存" width="100">
            <template #header><span class="col-required">库存</span></template>
            <template #default="{ row }">
              <el-input v-model.number="row.stock" placeholder="请输入">
                <template #suffix>件</template>
              </el-input>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <el-empty v-else-if="!loading" description="暂无 SKU" :image-size="64" />

      <p v-if="productId" class="foot-hint">
        修改规格、编码等请
        <router-link :to="`/products/${productId}/edit`" target="_blank">前往商品编辑</router-link>
      </p>
    </div>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" class="save-btn" :loading="saving" @click="handleSave">保存</el-button>
    </template>
    <el-image-viewer
      v-if="skuPreviewVisible"
      :url-list="[skuPreviewPic]"
      teleported
      hide-on-click-modal
      @close="skuPreviewVisible = false"
    />
  </el-dialog>
</template>

<style scoped>
.dialog-header {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  padding-right: 24px;
}

.dialog-header-label {
  flex-shrink: 0;
  font-size: 16px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
}

.dialog-header-divider {
  flex-shrink: 0;
  color: #c0c4cc;
  font-weight: 400;
}

.dialog-header-title {
  font-size: 15px;
  font-weight: 400;
  color: rgba(0, 0, 0, 0.65);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dialog-header-meta {
  flex-shrink: 0;
  margin-left: auto;
  font-size: 12px;
  color: #909399;
}

.csv-input-hidden {
  display: none;
}

.dialog-body {
  min-height: 120px;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.search-input {
  flex: 1;
  max-width: 240px;
}

.sku-count {
  font-size: 13px;
  color: #909399;
}

.batch-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  padding: 8px 10px;
  background: #fafafa;
  border-radius: 4px;
}

.batch-input {
  width: 108px;
}

.batch-btn {
  --el-button-bg-color: #ff7700;
  --el-button-border-color: #ff7700;
  --el-button-hover-bg-color: #ff8c1a;
  --el-button-hover-border-color: #ff8c1a;
}

.sku-table-wrap {
  overflow-x: auto;
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
  vertical-align: middle;
}

.sku-pic :deep(.el-image__inner) {
  display: block;
  width: 40px;
  height: 40px;
  object-fit: cover;
}

.sku-pic-empty {
  color: #c0c4cc;
  font-size: 13px;
  line-height: 40px;
}

.col-required::before {
  content: '*';
  color: #ff4d4f;
  margin-right: 2px;
}

.sku-table :deep(.el-table__cell) {
  padding: 6px 8px;
}

.sku-table :deep(.el-input__wrapper) {
  min-height: 32px;
  padding: 0 6px;
  box-shadow: 0 0 0 1px #d9d9d9 inset;
}

.sku-table :deep(.el-input__inner) {
  font-size: 14px;
  height: 32px;
  line-height: 32px;
}

.foot-hint {
  margin: 12px 0 0;
  font-size: 12px;
  color: #909399;
}

.foot-hint a {
  color: #ff7700;
  text-decoration: none;
}

.foot-hint a:hover {
  text-decoration: underline;
}

.save-btn {
  --el-button-bg-color: #ff7700;
  --el-button-border-color: #ff7700;
  --el-button-hover-bg-color: #ff8c1a;
  --el-button-hover-border-color: #ff8c1a;
}
</style>

<style>
.sku-quick-dialog.el-dialog {
  max-width: calc(100vw - 32px);
}
</style>
