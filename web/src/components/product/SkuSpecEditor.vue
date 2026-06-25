<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { Plus, Delete, View } from '@element-plus/icons-vue'
import PictureCardUpload from './PictureCardUpload.vue'
import ErpFormRow from './ErpFormRow.vue'
import { ElImageViewer, ElMessage, ElMessageBox } from 'element-plus'
import type { SkuItem, SkuSpec, SkuSpecValue } from '../../types/product'
import { duplicateSpecNameIndexes, duplicateSpecValueIndexes, emptySpecValue } from '../../types/product'
import { isValidSkuCode, nextSkuCode, sanitizeSkuCode } from '../../utils/skuCode'
import { MEDIA_UPLOAD_RULES } from '../../utils/uploadValidate'
import type { UploadContext } from '../../api/upload'

const props = defineProps<{
  modelValue: SkuSpec[]
  skus: SkuItem[]
  /** 父组件保存校验失败时递增，触发内联红框 */
  validateTick?: number
  /** 仅展示规格，不可编辑（列表页 SKU 管理） */
  specsReadonly?: boolean
  /** 允许增删 SKU 行（列表页 SKU 管理） */
  manageMode?: boolean
  uploadContext?: UploadContext
}>()

const emit = defineEmits<{
  'update:modelValue': [value: SkuSpec[]]
  'update:skus': [value: SkuItem[]]
}>()

function cloneSpecs(list: SkuSpec[]): SkuSpec[] {
  return JSON.parse(JSON.stringify(list))
}

function cloneSkus(list: SkuItem[]): SkuItem[] {
  return JSON.parse(JSON.stringify(list))
}

const localSpecs = ref<SkuSpec[]>(cloneSpecs(props.modelValue))
const localSkus = ref<SkuItem[]>(cloneSkus(props.skus))
/** 已生成/已保存的规格组合，编码不可再改 */
const lockedComboKeys = ref<Set<string>>(new Set())
let syncing = false

function comboKeyForSku(sku: SkuItem): string {
  const names = localSpecs.value
    .filter((s) => s.name.trim() && specValueTexts(s.values).length > 0)
    .map((s) => s.name.trim())
  return comboKey(names.map((n) => sku.specs[n] || ''))
}

function syncLockedKeys(skus: SkuItem[]) {
  const set = new Set<string>()
  for (const s of skus) {
    if (sanitizeSkuCode(s.skuCode)) {
      set.add(comboKeyForSku(s))
    }
  }
  lockedComboKeys.value = set
}

function isSkuCodeLocked(row: SkuItem): boolean {
  return lockedComboKeys.value.has(comboKeyForSku(row))
}

function lockSkuRow(row: SkuItem) {
  if (!sanitizeSkuCode(row.skuCode)) return
  lockedComboKeys.value = new Set([...lockedComboKeys.value, comboKeyForSku(row)])
}

function usedSkuCodes(): Set<string> {
  return new Set(localSkus.value.map((s) => sanitizeSkuCode(s.skuCode)).filter(Boolean))
}

function generateSkuCodeForRow(row: SkuItem) {
  if (isSkuCodeLocked(row)) return
  const used = usedSkuCodes()
  row.skuCode = nextSkuCode(used)
  lockSkuRow(row)
}

function generateAllSkuCodes() {
  const used = usedSkuCodes()
  let generated = 0
  const nextLocked = new Set(lockedComboKeys.value)
  localSkus.value = localSkus.value.map((sku) => {
    if (!skuMatchesFilter(sku) || nextLocked.has(comboKeyForSku(sku))) return sku
    const code = nextSkuCode(used)
    nextLocked.add(comboKeyForSku({ ...sku, skuCode: code }))
    generated++
    return { ...sku, skuCode: code }
  })
  lockedComboKeys.value = nextLocked
  if (generated) ElMessage.success(`已生成 ${generated} 个规格编码`)
  else ElMessage.info('当前筛选下没有待生成的规格')
}

function onSkuCodeInput(row: SkuItem, val: string) {
  row.skuCode = sanitizeSkuCode(val)
}

function onSkuCodeBlur(row: SkuItem) {
  if (isValidSkuCode(row.skuCode)) {
    lockSkuRow(row)
  }
}

syncLockedKeys(localSkus.value)

function getSpecPic(si: number, vi: number) {
  const pic = localSpecs.value[si]?.values[vi]?.pic
  return pic ? [pic] : []
}

function setSpecPic(si: number, vi: number, pics: string[]) {
  const item = localSpecs.value[si]?.values[vi]
  if (!item) return
  item.pic = pics[0] || undefined
  touchSpecs()
}

function touchSpecs() {
  localSpecs.value = cloneSpecs(localSpecs.value)
}

function addSpec() {
  localSpecs.value = [...localSpecs.value, { name: '', values: [emptySpecValue()] }]
}

function removeSpec(index: number) {
  localSpecs.value = localSpecs.value.filter((_, i) => i !== index)
}

function addSpecValue(specIndex: number) {
  localSpecs.value = localSpecs.value.map((s, i) =>
    i === specIndex ? { ...s, values: [...s.values, emptySpecValue()] } : s,
  )
}

function removeSpecValue(specIndex: number, valueIndex: number) {
  localSpecs.value = localSpecs.value.map((s, i) =>
    i === specIndex ? { ...s, values: s.values.filter((_, vi) => vi !== valueIndex) } : s,
  )
}

function specValueTexts(values: SkuSpecValue[]) {
  return values.map((v) => v.value.trim()).filter(Boolean)
}

function onSpecValueBlur(specIndex: number, valueIndex: number) {
  const spec = localSpecs.value[specIndex]
  const val = spec?.values[valueIndex]?.value.trim()
  if (!spec || !val) return
  const duplicated = spec.values.some((v, i) => i !== valueIndex && v.value.trim() === val)
  if (duplicated) {
    ElMessage.warning(`规格「${spec.name.trim() || '规格'}」存在重复规格值「${val}」`)
  }
}

function comboKey(values: string[]) {
  return values.join('\0')
}

function findExistingSku(
  combo: string[],
  specNames: string[],
  existing: SkuItem[],
): SkuItem | undefined {
  const key = comboKey(combo)
  for (const sku of existing) {
    const vals = specNames.map((name) => sku.specs[name] || '')
    if (comboKey(vals) === key) return sku
  }
  for (const sku of existing) {
    if (comboKey(Object.values(sku.specs)) === key) return sku
  }
  return undefined
}

function applySpecPics(skus: SkuItem[]): SkuItem[] {
  const first = localSpecs.value[0]
  if (!first?.name) return skus
  return skus.map((sku) => {
    const val = sku.specs[first.name]
    if (!val) return sku
    const matched = first.values.find((v) => v.value === val)
    return matched?.pic ? { ...sku, pic: matched.pic } : sku
  })
}

function regenerateSkus(preserveFrom?: SkuItem[]) {
  const baseline = preserveFrom ?? localSkus.value
  const readySpecs = localSpecs.value.filter(
    (s) => s.name.trim() && specValueTexts(s.values).length > 0,
  )

  if (readySpecs.length === 0) {
    commitSkus([])
    return
  }

  const specNames = readySpecs.map((s) => s.name.trim())
  const valueArrays = readySpecs.map((s) => specValueTexts(s.values))
  const combinations = cartesian(valueArrays)

  const nextSkus = combinations.map((combo) => {
    const specRecord: Record<string, string> = {}
    combo.forEach((val, i) => {
      const name = specNames[i]
      if (name) specRecord[name] = val
    })
    const existing = findExistingSku(combo, specNames, baseline)
    if (existing) {
      return { ...existing, specs: specRecord }
    }
    const firstVal = combo[0]
    const matched = readySpecs[0]?.values.find((v) => v.value === firstVal)
    return {
      skuCode: '',
      specs: specRecord,
      price: 0,
      costPrice: 0,
      marketPrice: 0,
      stock: 0,
      weight: 0,
      pic: matched?.pic,
    } satisfies SkuItem
  })

  commitSkus(applySpecPics(nextSkus))
}

function commitSkus(list: SkuItem[]) {
  syncing = true
  localSkus.value = list
  syncLockedKeys(list)
  emit('update:skus', cloneSkus(list))
  pruneFilters()
  nextTick(() => {
    syncing = false
  })
}

function cartesian<T>(arrays: T[][]): T[][] {
  return arrays.reduce<T[][]>(
    (acc, curr) => acc.flatMap((a) => curr.map((c) => [...a, c])),
    [[]],
  )
}

watch(
  localSpecs,
  () => {
    if (syncing || props.specsReadonly) return
    syncing = true
    emit('update:modelValue', cloneSpecs(localSpecs.value))
    regenerateSkus()
    nextTick(() => {
      syncing = false
    })
  },
  { deep: true },
)

watch(
  () => props.modelValue,
  (v) => {
    if (syncing) return
    if (specsSignature(v) === specsSignature(localSpecs.value)) return
    localSpecs.value = cloneSpecs(v)
    if (!props.specsReadonly) {
      regenerateSkus(cloneSkus(props.skus))
    }
  },
  { deep: true },
)

watch(
  () => props.skus,
  (v) => {
    if (syncing) return
    if (v.length === 0 && localSkus.value.length === 0) return
    if (skusSignature(v) === skusSignature(localSkus.value)) return
    localSkus.value = cloneSkus(v)
    syncLockedKeys(localSkus.value)
  },
  { deep: true },
)

watch(
  localSkus,
  (v) => {
    if (syncing) return
    syncing = true
    emit('update:skus', cloneSkus(v))
    nextTick(() => {
      syncing = false
    })
  },
  { deep: true },
)

function specsSignature(list: SkuSpec[]) {
  return JSON.stringify(
    list.map((s) => ({
      name: s.name,
      values: s.values.map((v) => ({ value: v.value, remark: v.remark, pic: v.pic })),
    })),
  )
}

function skusSignature(list: SkuItem[]) {
  return JSON.stringify(list.map((s) => ({ specs: s.specs, skuCode: s.skuCode })))
}

const filterBySpec = reactive<Record<string, string>>({})
const batchMarket = ref<number>()
const batchPrice = ref<number>()
const batchStock = ref<number>()
const batchWeight = ref<number>()

const skuPicViewerUrl = ref('')

function openSkuPicPreview(url: string) {
  skuPicViewerUrl.value = url
}

function closeSkuPicPreview() {
  skuPicViewerUrl.value = ''
}

function getFilterOptions(specName: string) {
  const values = new Set<string>()
  for (const sku of localSkus.value) {
    const v = sku.specs[specName]
    if (v) values.add(v)
  }
  return [...values].map((v) => ({
    label: formatSpecDisplay(specName, v),
    value: v,
  }))
}

function skuMatchesFilter(sku: SkuItem) {
  for (const specName of tableSpecColumns.value) {
    const fv = filterBySpec[specName]
    if (fv && sku.specs[specName] !== fv) return false
  }
  return true
}

function pruneFilters() {
  const names = new Set(tableSpecColumns.value)
  for (const key of Object.keys(filterBySpec)) {
    if (!names.has(key)) {
      delete filterBySpec[key]
      continue
    }
    const fv = filterBySpec[key]
    if (fv && !localSkus.value.some((sku) => sku.specs[key] === fv)) {
      delete filterBySpec[key]
    }
  }
}

const filteredSkus = computed(() => localSkus.value.filter(skuMatchesFilter))

function updateSpecName(specIndex: number, name: string) {
  localSpecs.value = localSpecs.value.map((s, i) => (i === specIndex ? { ...s, name } : s))
}

function updateSpecValue(specIndex: number, valueIndex: number, value: string) {
  localSpecs.value = localSpecs.value.map((s, i) =>
    i === specIndex
      ? { ...s, values: s.values.map((v, vi) => (vi === valueIndex ? { ...v, value } : v)) }
      : s,
  )
}

const duplicateSpecNameDups = ref<Set<number>>(new Set())
const duplicateSpecValueDups = ref<Map<number, Set<number>>>(new Map())

function recomputeDuplicateErrors() {
  duplicateSpecNameDups.value = duplicateSpecNameIndexes(localSpecs.value)
  duplicateSpecValueDups.value = duplicateSpecValueIndexes(localSpecs.value)
}

watch(localSpecs, recomputeDuplicateErrors, { deep: true, immediate: true })

watch(
  () => props.validateTick,
  () => {
    if (props.validateTick) recomputeDuplicateErrors()
  },
)

function isDuplicateSpecName(index: number): boolean {
  return duplicateSpecNameDups.value.has(index)
}

function isDuplicateSpecValue(specIndex: number, valueIndex: number): boolean {
  return duplicateSpecValueDups.value.get(specIndex)?.has(valueIndex) ?? false
}

function ensureAllSkuCodesSilent() {
  const used = usedSkuCodes()
  localSkus.value = localSkus.value.map((sku) => {
    const code = sanitizeSkuCode(sku.skuCode)
    if (code) {
      used.add(code)
      return { ...sku, skuCode: code }
    }
    const generated = nextSkuCode(used)
    return { ...sku, skuCode: generated }
  })
}

function flushSpecsToParent() {
  emit('update:modelValue', cloneSpecs(localSpecs.value))
}

defineExpose({ flushSpecsToParent, recomputeDuplicateErrors, ensureAllSkuCodesSilent })

const skuTableKey = computed(() => specsSignature(localSpecs.value))

const tableSpecColumns = computed(() =>
  localSpecs.value
    .filter((s) => s.name.trim() && specValueTexts(s.values).length > 0)
    .map((s) => s.name.trim()),
)

function getSpecRemark(specName: string, value?: string) {
  if (!value) return ''
  const spec = localSpecs.value.find((s) => s.name.trim() === specName)
  return spec?.values.find((v) => v.value === value)?.remark?.trim() || ''
}

function formatSpecDisplay(specName: string, value?: string) {
  if (!value) return ''
  const remark = getSpecRemark(specName, value)
  return remark ? `${value}(${remark})` : value
}

function applyBatch() {
  const targets = localSkus.value.filter(skuMatchesFilter)
  if (!targets.length) return

  localSkus.value = localSkus.value.map((sku) => {
    if (!skuMatchesFilter(sku)) return sku
    return {
      ...sku,
      ...(batchMarket.value !== undefined && batchMarket.value !== null
        ? { marketPrice: batchMarket.value }
        : {}),
      ...(batchPrice.value !== undefined && batchPrice.value !== null ? { price: batchPrice.value } : {}),
      ...(batchStock.value !== undefined && batchStock.value !== null ? { stock: batchStock.value } : {}),
      ...(batchWeight.value !== undefined && batchWeight.value !== null ? { weight: batchWeight.value } : {}),
    }
  })
  ElMessage.success('批量填充完成')
}

const readonlySpecs = computed(() =>
  localSpecs.value.filter((s) => s.name.trim() || specValueTexts(s.values).length > 0),
)

function getReadySpecs() {
  return localSpecs.value.filter((s) => s.name.trim() && specValueTexts(s.values).length > 0)
}

function skuComboKeyFromRecord(specRecord: Record<string, string>): string {
  const names = getReadySpecs().map((s) => s.name.trim())
  return comboKey(names.map((n) => specRecord[n] || ''))
}

function addSkuInManageMode() {
  const readySpecs = getReadySpecs()
  if (readySpecs.length === 0) {
    if (localSkus.value.length > 0) {
      ElMessage.warning('暂无规格维度，仅支持一条 SKU')
      return
    }
    localSkus.value.push({
      skuCode: '',
      specs: {},
      price: 0,
      costPrice: 0,
      marketPrice: 0,
      stock: 0,
      weight: 0,
    })
    ElMessage.success('已添加 SKU')
    return
  }
  const specNames = readySpecs.map((s) => s.name.trim())
  const valueArrays = readySpecs.map((s) => specValueTexts(s.values))
  const combinations = cartesian(valueArrays)
  const existingKeys = new Set(localSkus.value.map((s) => skuComboKeyFromRecord(s.specs || {})))
  for (const combo of combinations) {
    const specRecord: Record<string, string> = {}
    combo.forEach((val, i) => {
      if (specNames[i]) specRecord[specNames[i]] = val
    })
    const key = comboKey(combo)
    if (!existingKeys.has(key)) {
      const firstVal = combo[0]
      const matched = readySpecs[0]?.values.find((v) => v.value === firstVal)
      localSkus.value.push({
        skuCode: '',
        specs: specRecord,
        price: 0,
        costPrice: 0,
        marketPrice: 0,
        stock: 0,
        weight: 0,
        pic: matched?.pic,
      })
      ElMessage.success('已添加 SKU')
      return
    }
  }
  ElMessage.info('所有规格组合已存在')
}

function generateAllCombosInManageMode() {
  const readySpecs = getReadySpecs()
  if (!readySpecs.length) {
    ElMessage.warning('请先配置商品规格')
    return
  }
  regenerateSkus(localSkus.value)
  ElMessage.success('已根据规格生成 SKU 列表')
}

async function removeSkuRow(row: SkuItem) {
  try {
    await ElMessageBox.confirm('确定删除该 SKU？', '删除确认', { type: 'warning' })
    const idx = localSkus.value.indexOf(row)
    if (idx >= 0) localSkus.value.splice(idx, 1)
  } catch {
    /* cancelled */
  }
}
</script>

<template>
  <div class="sku-editor">
    <ErpFormRow
      label="商品规格"
      :hint="specsReadonly ? '规格维度请在商品编辑页修改；下方可管理各规格组合的价格与库存' : undefined"
    >
      <template v-if="specsReadonly">
        <div v-if="readonlySpecs.length" class="spec-list">
          <div v-for="(spec, si) in readonlySpecs" :key="si" class="spec-block">
            <div class="spec-block-head">
              <div class="spec-name-block">
                <div class="spec-name-row">
                  <span class="field-label">规格名</span>
                  <span class="spec-readonly-name">{{ spec.name || '-' }}</span>
                  <p v-if="si === 0" class="spec-note">
                    注:仅第一组规格支持添加规格图片;若要上传图片,所有规格值都需要上传
                  </p>
                </div>
              </div>
            </div>
            <div class="spec-value-row">
              <span class="field-label">规格值</span>
              <div class="spec-value-area">
                <div class="spec-value-grid">
                  <div
                    v-for="(sv, vi) in spec.values.filter((v) => v.value.trim() || v.remark?.trim() || v.pic)"
                    :key="vi"
                    class="spec-value-cell spec-value-cell--readonly"
                  >
                    <span class="spec-readonly-value">{{ sv.value || '-' }}</span>
                    <span v-if="sv.remark?.trim()" class="spec-readonly-remark">{{ sv.remark }}</span>
                    <el-image
                      v-if="si === 0 && sv.pic"
                      :src="sv.pic"
                      class="spec-readonly-pic"
                      fit="cover"
                      :preview-src-list="[sv.pic]"
                      preview-teleported
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <p v-else class="spec-readonly-empty">暂未配置商品规格，可直接添加 SKU 或前往商品编辑页配置</p>
      </template>
      <el-form v-else class="spec-editor-form" @submit.prevent>
      <div class="spec-list">
        <div v-for="(spec, si) in localSpecs" :key="si" class="spec-block">
          <div class="spec-block-head">
            <div class="spec-name-block">
              <div class="spec-name-row">
                <span class="field-label">规格名</span>
                <el-form-item
                  class="spec-name-form-item"
                  :class="{ 'spec-name-form-item--duplicate': isDuplicateSpecName(si) }"
                  :error="isDuplicateSpecName(si) ? '规格名重复' : undefined"
                  :show-message="isDuplicateSpecName(si)"
                >
                  <div
                    class="spec-name-input-shell"
                    :class="{ 'spec-name-input-shell--duplicate': isDuplicateSpecName(si) }"
                  >
                    <el-input
                      :model-value="spec.name"
                      placeholder="请输入"
                      class="spec-name-input"
                      @update:model-value="updateSpecName(si, $event)"
                    />
                  </div>
                </el-form-item>
                <p v-if="si === 0" class="spec-note">
                  注:仅第一组规格支持添加规格图片;若要上传图片,所有规格值都需要上传
                </p>
              </div>
            </div>
            <button
              v-if="localSpecs.length > 1"
              type="button"
              class="spec-delete-link"
              @click="removeSpec(si)"
            >
              删除
            </button>
          </div>
          <div class="spec-value-row">
            <span class="field-label">规格值</span>
            <div class="spec-value-area">
              <div class="spec-value-grid">
                <div v-for="(sv, vi) in spec.values" :key="vi" class="spec-value-cell">
                  <div
                    class="spec-value-input-shell"
                    :class="{ 'spec-value-input-shell--duplicate': isDuplicateSpecValue(si, vi) }"
                  >
                    <el-input
                      :model-value="sv.value"
                      placeholder="请输入"
                      @update:model-value="updateSpecValue(si, vi, $event)"
                      @blur="onSpecValueBlur(si, vi)"
                    />
                  </div>
                  <el-input
                    v-model="sv.remark"
                    placeholder="规格备注"
                  />
                  <div class="spec-value-actions">
                    <PictureCardUpload
                      v-if="si === 0"
                      :model-value="getSpecPic(si, vi)"
                      :max="1"
                      size="inline"
                      upload-label="上传规格图"
                      :rules="MEDIA_UPLOAD_RULES.skuSpec"
                      :upload-context="uploadContext"
                      @update:model-value="setSpecPic(si, vi, $event)"
                    />
                    <button
                      v-if="spec.values.length > 1"
                      type="button"
                      class="value-delete-btn"
                      @click="removeSpecValue(si, vi)"
                    >
                      <el-icon><Delete /></el-icon>
                    </button>
                  </div>
                </div>
                <button
                  type="button"
                  class="add-value-btn"
                  title="添加规格值"
                  @click="addSpecValue(si)"
                >
                  <el-icon><Plus /></el-icon>
                </button>
              </div>
            </div>
          </div>
        </div>
        <button type="button" class="add-spec-btn" @click="addSpec">
          <el-icon><Plus /></el-icon>
          添加规格项
        </button>
      </div>
      </el-form>
    </ErpFormRow>

    <ErpFormRow v-if="manageMode || localSkus.length" label="规格信息">
      <div class="sku-panel">
        <div class="batch-fill-bar">
          <el-select
            v-for="specName in tableSpecColumns"
            :key="specName"
            v-model="filterBySpec[specName]"
            clearable
            :placeholder="`全部${specName}`"
            class="bf-select"
          >
            <el-option
              v-for="opt in getFilterOptions(specName)"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
          <el-input v-model.number="batchMarket" placeholder="市场价" class="bf-input">
            <template #suffix>元</template>
          </el-input>
          <el-input v-model.number="batchPrice" placeholder="销售价" class="bf-input">
            <template #suffix>元</template>
          </el-input>
          <el-input v-model.number="batchStock" placeholder="库存" class="bf-input">
            <template #suffix>件</template>
          </el-input>
          <el-button class="bf-btn" @click="generateAllSkuCodes">批量生成编码</el-button>
          <el-input v-model.number="batchWeight" placeholder="重量" class="bf-input weight" />
          <el-button type="primary" class="bf-btn" @click="applyBatch">批量填充</el-button>
          <template v-if="manageMode">
            <el-button class="bf-btn bf-btn-outline" @click="addSkuInManageMode">添加 SKU</el-button>
            <el-button
              v-if="getReadySpecs().length"
              class="bf-btn bf-btn-outline"
              @click="generateAllCombosInManageMode"
            >
              生成全部规格
            </el-button>
          </template>
        </div>

        <el-table
          v-if="localSkus.length"
          :key="skuTableKey"
          :data="filteredSkus"
          border
          class="sku-table"
          :row-key="(row: SkuItem) => JSON.stringify(row.specs)"
        >
          <el-table-column
            v-for="specName in tableSpecColumns"
            :key="specName"
            :label="specName"
            min-width="240"
            show-overflow-tooltip
          >
            <template #default="{ row }">
              <span class="spec-cell-text">{{
                formatSpecDisplay(specName, row.specs[specName])
              }}</span>
            </template>
          </el-table-column>
          <el-table-column label="预览图" width="80" align="center">
            <template #default="{ row }">
              <div v-if="row.pic" class="sku-preview">
                <el-image :src="row.pic" fit="cover" class="sku-thumb" />
                <button
                  type="button"
                  class="preview-eye-btn"
                  title="预览"
                  @click.stop="openSkuPicPreview(row.pic!)"
                >
                  <el-icon><View /></el-icon>
                </button>
              </div>
              <span v-else class="sku-preview-empty">-</span>
            </template>
          </el-table-column>
          <el-table-column label="市场价" width="108">
            <template #default="{ row }">
              <el-input v-model.number="row.marketPrice" placeholder="请输入">
                <template #suffix>元</template>
              </el-input>
            </template>
          </el-table-column>
          <el-table-column width="110">
            <template #header><span class="required-col">销售价</span></template>
            <template #default="{ row }">
              <el-input v-model.number="row.price" placeholder="请输入">
                <template #suffix>元</template>
              </el-input>
            </template>
          </el-table-column>
          <el-table-column width="100">
            <template #header><span class="required-col">库存</span></template>
            <template #default="{ row }">
              <el-input v-model.number="row.stock" placeholder="请输入">
                <template #suffix>件</template>
              </el-input>
            </template>
          </el-table-column>
          <el-table-column label="规格编码" min-width="200">
            <template #default="{ row }">
              <div class="sku-code-cell">
                <el-input
                  :model-value="row.skuCode"
                  placeholder="点击生成"
                  maxlength="64"
                  :disabled="isSkuCodeLocked(row)"
                  @update:model-value="onSkuCodeInput(row, $event)"
                  @blur="onSkuCodeBlur(row)"
                />
                <el-button
                  v-if="!isSkuCodeLocked(row)"
                  link
                  type="primary"
                  class="gen-code-btn"
                  @click="generateSkuCodeForRow(row)"
                >
                  生成
                </el-button>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="重量(g)" width="96">
            <template #default="{ row }">
              <el-input v-model.number="row.weight" placeholder="请输入" />
            </template>
          </el-table-column>
          <el-table-column v-if="manageMode" label="操作" width="72" fixed="right" align="center">
            <template #default="{ row }">
              <button type="button" class="sku-row-delete" @click="removeSkuRow(row)">删除</button>
            </template>
          </el-table-column>
        </el-table>
        <el-empty
          v-else-if="manageMode"
          description="暂无 SKU，可点击「添加 SKU」或「生成全部规格」"
          :image-size="72"
        />
      </div>
    </ErpFormRow>

    <el-image-viewer
      v-if="skuPicViewerUrl"
      :url-list="[skuPicViewerUrl]"
      teleported
      @close="closeSkuPicPreview"
    />
  </div>
</template>

<style scoped>
.sku-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 15px;
}

.spec-list {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.spec-editor-form {
  width: 100%;
}

.spec-name-input-shell {
  width: 100%;
}

.spec-block {
  background: #f0f0f0;
  border-radius: 4px;
  padding: 12px;
}

.spec-block-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.spec-name-block {
  flex: 1;
  min-width: 0;
}

.spec-name-row {
  display: flex;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.spec-value-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-top: 10px;
}

.field-label {
  flex-shrink: 0;
  width: 64px;
  font-size: 14px;
  color: rgba(0, 0, 0, 0.85);
  line-height: 32px;
  white-space: nowrap;
  text-align: right;
}

.spec-name-input {
  width: 100%;
}

.spec-name-form-item {
  width: 220px;
  flex-shrink: 0;
  margin-bottom: 0;
}

.spec-name-form-item :deep(.el-form-item__content) {
  line-height: normal;
}

.spec-name-form-item :deep(.el-form-item__error) {
  position: static;
  padding-top: 2px;
  font-size: 12px;
}

.spec-delete-link {
  flex-shrink: 0;
  border: none;
  background: transparent;
  padding: 0;
  font-size: 15px;
  line-height: 34px;
  color: #ff7700;
  cursor: pointer;
}

.spec-delete-link:hover {
  color: #ff8c1a;
}

.spec-note {
  margin: 0;
  font-size: 13px;
  color: rgba(0, 0, 0, 0.45);
  line-height: 32px;
  flex: 1;
  min-width: 200px;
}

.spec-value-area {
  flex: 1;
  min-width: 0;
}

.spec-value-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(260px, 1fr));
  gap: 10px 16px;
}

.spec-value-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.spec-value-input-shell {
  flex: 1;
  min-width: 0;
}

.spec-value-input-shell :deep(.el-input) {
  width: 100%;
}

.spec-value-cell :deep(.el-input) {
  flex: 1;
  min-width: 0;
}

.spec-value-cell :deep(.el-input__wrapper) {
  min-height: 32px;
  padding-top: 0;
  padding-bottom: 0;
}

.spec-value-cell :deep(.el-input__inner) {
  font-size: 14px;
  height: 32px;
  line-height: 32px;
}

.spec-value-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.value-delete-btn {
  border: none;
  background: transparent;
  padding: 4px;
  color: rgba(0, 0, 0, 0.45);
  cursor: pointer;
  display: flex;
  align-items: center;
  font-size: 16px;
  line-height: 1;
}

.value-delete-btn:hover {
  color: #ff7700;
}

.add-value-btn {
  width: 32px;
  height: 32px;
  border: 1px dashed #d9d9d9;
  border-radius: 4px;
  background: #fff;
  color: rgba(0, 0, 0, 0.45);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  align-self: center;
  font-size: 16px;
}

.add-value-btn:hover {
  border-color: #ff7700;
  color: #ff7700;
}

.add-spec-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  height: 28px;
  padding: 0 10px;
  border: 1px dashed #d9d9d9;
  border-radius: 4px;
  background: #fff;
  font-size: 13px;
  color: #ff7700;
  cursor: pointer;
  align-self: flex-start;
}

.add-spec-btn:hover {
  border-color: #ff7700;
  background: #fff7e6;
}

.spec-readonly-name {
  width: 220px;
  flex-shrink: 0;
  font-size: 14px;
  line-height: 32px;
  color: rgba(0, 0, 0, 0.85);
}

.spec-readonly-empty {
  margin: 0;
  font-size: 13px;
  color: rgba(0, 0, 0, 0.45);
  line-height: 1.5;
}

.spec-value-cell--readonly {
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
  padding: 6px 10px;
  min-height: 32px;
}

.spec-readonly-value {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.85);
}

.spec-readonly-remark {
  font-size: 13px;
  color: rgba(0, 0, 0, 0.45);
}

.spec-readonly-pic {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  flex-shrink: 0;
}

.bf-btn-outline {
  --el-button-bg-color: #fff;
  --el-button-border-color: #ff7700;
  --el-button-text-color: #ff7700;
  --el-button-hover-bg-color: #fff7e6;
  --el-button-hover-border-color: #ff7700;
  --el-button-hover-text-color: #ff7700;
}

.sku-row-delete {
  border: none;
  background: transparent;
  padding: 0;
  font-size: 14px;
  color: #ff4d4f;
  cursor: pointer;
}

.sku-row-delete:hover {
  color: #ff7875;
}

.sku-panel {
  width: 100%;
  overflow-x: auto;
}

.batch-fill-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
  background: #f0f0f0;
  border-radius: 4px;
  padding: 10px 12px;
}

.bf-select {
  width: 148px;
}

.bf-input {
  width: 100px;
}

.bf-input.weight {
  width: 88px;
}

.sku-code-cell {
  display: flex;
  align-items: center;
  gap: 4px;
}

.sku-code-cell :deep(.el-input) {
  flex: 1;
  min-width: 0;
}

.gen-code-btn {
  flex-shrink: 0;
  padding: 0 4px;
  font-size: 14px;
}

.bf-btn {
  --el-button-bg-color: #ff7700;
  --el-button-border-color: #ff7700;
  --el-button-hover-bg-color: #ff8c1a;
  --el-button-hover-border-color: #ff8c1a;
  --el-button-active-bg-color: #e66a00;
  --el-button-active-border-color: #e66a00;
}

.required-col::before {
  content: '*';
  color: #ff4d4f;
  margin-right: 1px;
}

.sku-table :deep(.el-table__cell) {
  padding: 6px 8px;
}

.sku-table :deep(.el-table__body tr) {
  height: auto;
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

.sku-table :deep(.el-input__suffix) {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.45);
}

.sku-table :deep(.el-table__header th) {
  background: #fafafa;
  padding: 8px 0;
  font-size: 14px;
  font-weight: 500;
  color: rgba(0, 0, 0, 0.85);
  white-space: nowrap;
}

.sku-table :deep(.el-table__body td) {
  font-size: 14px;
  vertical-align: middle;
}

.spec-cell-text {
  display: block;
  line-height: 1.4;
  font-size: 14px;
  color: rgba(0, 0, 0, 0.85);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sku-preview {
  position: relative;
  width: 48px;
  height: 48px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  background: #fafafa;
  border: 1px solid #e8e8e8;
  overflow: hidden;
}

.sku-preview .preview-eye-btn {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 22px;
  height: 22px;
  border: none;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s, background 0.2s;
  z-index: 1;
  padding: 0;
}

.sku-preview:hover .preview-eye-btn {
  opacity: 1;
}

.sku-preview .preview-eye-btn .el-icon {
  font-size: 12px;
}

.sku-preview .preview-eye-btn:hover {
  background: rgba(0, 0, 0, 0.7);
}

.sku-thumb {
  width: 48px;
  height: 48px;
}

.sku-thumb :deep(.el-image__inner) {
  width: 48px;
  height: 48px;
  object-fit: cover;
  display: block;
  pointer-events: none;
}

.sku-preview-empty {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.25);
  line-height: 1;
}
</style>

<style>
/* 与 Element Plus 表单校验一致：1px 内描边，inset 避免被 overflow 裁切 */
.sku-editor .spec-name-input-shell--duplicate .el-input__wrapper,
.sku-editor .spec-name-input-shell--duplicate .el-input__wrapper:hover,
.sku-editor .spec-name-input-shell--duplicate .el-input__wrapper.is-focus,
.sku-editor .spec-name-form-item--duplicate.is-error .el-input__wrapper,
.sku-editor .spec-name-form-item--duplicate.is-error .el-input__wrapper:hover,
.sku-editor .spec-name-form-item--duplicate.is-error .el-input__wrapper.is-focus,
.sku-editor .spec-value-input-shell--duplicate .el-input__wrapper,
.sku-editor .spec-value-input-shell--duplicate .el-input__wrapper:hover,
.sku-editor .spec-value-input-shell--duplicate .el-input__wrapper.is-focus {
  box-shadow: 0 0 0 1px var(--el-color-danger) inset !important;
}
</style>
