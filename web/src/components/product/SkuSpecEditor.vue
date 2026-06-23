<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'
import type { SkuItem, SkuSpec } from '../../types/product'

const props = defineProps<{
  modelValue: SkuSpec[]
  skus: SkuItem[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: SkuSpec[]]
  'update:skus': [value: SkuItem[]]
}>()

const specs = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const skuList = computed({
  get: () => props.skus,
  set: (v) => emit('update:skus', v),
})

function addSpec() {
  specs.value = [...specs.value, { name: '', values: [''] }]
}

function removeSpec(index: number) {
  const next = specs.value.filter((_, i) => i !== index)
  specs.value = next
  regenerateSkus()
}

function addSpecValue(specIndex: number) {
  const next = specs.value.map((s, i) =>
    i === specIndex ? { ...s, values: [...s.values, ''] } : s,
  )
  specs.value = next
}

function removeSpecValue(specIndex: number, valueIndex: number) {
  const next = specs.value.map((s, i) =>
    i === specIndex ? { ...s, values: s.values.filter((_, vi) => vi !== valueIndex) } : s,
  )
  specs.value = next
  regenerateSkus()
}

function cartesian<T>(arrays: T[][]): T[][] {
  return arrays.reduce<T[][]>(
    (acc, curr) => acc.flatMap((a) => curr.map((c) => [...a, c])),
    [[]],
  )
}

function regenerateSkus() {
  const validSpecs = specs.value.filter((s) => s.name && s.values.some((v) => v.trim()))
  if (validSpecs.length === 0) {
    skuList.value = []
    return
  }

  const specNames = validSpecs.map((s) => s.name)
  const valueArrays = validSpecs.map((s) => s.values.filter((v) => v.trim()))

  if (valueArrays.some((arr) => arr.length === 0)) return

  const combinations = cartesian(valueArrays)
  const existingMap = new Map(
    skuList.value.map((sku) => [JSON.stringify(sku.specs), sku]),
  )

  skuList.value = combinations.map((combo) => {
    const specRecord: Record<string, string> = {}
    combo.forEach((val, i) => {
      const name = specNames[i]
      if (name) specRecord[name] = val
    })
    const key = JSON.stringify(specRecord)
    const existing = existingMap.get(key)
    if (existing) return existing
    const code = combo.join('-').replace(/\s/g, '')
    return {
      skuCode: code,
      specs: specRecord,
      price: 0,
      costPrice: 0,
      stock: 0,
    } satisfies SkuItem
  })
}

const specColumns = computed(() =>
  specs.value.filter((s) => s.name).map((s) => s.name),
)

watch(
  () => props.modelValue,
  () => regenerateSkus(),
  { deep: true },
)

const batchPrice = ref<number>()
const batchStock = ref<number>()

function applyBatch(field: 'price' | 'stock') {
  const val = field === 'price' ? batchPrice.value : batchStock.value
  if (val === undefined) return
  skuList.value = skuList.value.map((sku) => ({ ...sku, [field]: val }))
}
</script>

<template>
  <div class="sku-editor">
    <div class="spec-section">
      <div class="section-header">
        <h4>销售规格</h4>
        <el-button type="primary" link :icon="Plus" @click="addSpec">添加规格项</el-button>
      </div>
      <p class="hint">参照淘宝/抖店：先定义规格维度（如颜色、尺码），系统自动生成 SKU 矩阵</p>

      <div v-for="(spec, si) in specs" :key="si" class="spec-row">
        <el-input v-model="spec.name" placeholder="规格名，如：颜色" style="width: 120px" @blur="regenerateSkus" />
        <div class="spec-values">
          <el-tag
            v-for="(_, vi) in spec.values"
            :key="vi"
            closable
            class="value-tag"
            @close="removeSpecValue(si, vi)"
          >
            <el-input
              v-model="spec.values[vi]"
              size="small"
              placeholder="规格值"
              style="width: 80px"
              @blur="regenerateSkus"
            />
          </el-tag>
          <el-button size="small" :icon="Plus" circle @click="addSpecValue(si)" />
        </div>
        <el-button type="danger" link :icon="Delete" @click="removeSpec(si)">删除</el-button>
      </div>
    </div>

    <div v-if="skuList.length" class="sku-table-section">
      <div class="section-header">
        <h4>SKU 列表 <el-tag size="small" type="info">{{ skuList.length }} 个 SKU</el-tag></h4>
        <div class="batch-actions">
          <el-input-number v-model="batchPrice" :min="0" :precision="2" placeholder="批量价格" size="small" controls-position="right" />
          <el-button size="small" @click="applyBatch('price')">应用价格</el-button>
          <el-input-number v-model="batchStock" :min="0" placeholder="批量库存" size="small" controls-position="right" />
          <el-button size="small" @click="applyBatch('stock')">应用库存</el-button>
        </div>
      </div>

      <el-table :data="skuList" border stripe size="small" max-height="400">
        <el-table-column v-for="col in specColumns" :key="col" :label="col" width="100">
          <template #default="{ row }">{{ row.specs[col] }}</template>
        </el-table-column>
        <el-table-column label="SKU 编码" min-width="140">
          <template #default="{ row }">
            <el-input v-model="row.skuCode" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="SKU 图片" width="80">
          <template #default="{ row }">
            <el-upload
              class="sku-uploader"
              action="#"
              :show-file-list="false"
              :auto-upload="false"
            >
              <img v-if="row.pic" :src="row.pic" class="sku-thumb" />
              <el-icon v-else class="upload-icon"><Plus /></el-icon>
            </el-upload>
          </template>
        </el-table-column>
        <el-table-column label="销售价" width="120">
          <template #default="{ row }">
            <el-input-number v-model="row.price" :min="0" :precision="2" size="small" controls-position="right" />
          </template>
        </el-table-column>
        <el-table-column label="成本价" width="120">
          <template #default="{ row }">
            <el-input-number v-model="row.costPrice" :min="0" :precision="2" size="small" controls-position="right" />
          </template>
        </el-table-column>
        <el-table-column label="库存" width="100">
          <template #default="{ row }">
            <el-input-number v-model="row.stock" :min="0" size="small" controls-position="right" />
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-empty v-else-if="specs.length" description="请完善规格名和规格值，系统将自动生成 SKU" />
  </div>
</template>

<style scoped>
.sku-editor {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.section-header h4 {
  margin: 0;
  font-size: 15px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.hint {
  color: #909399;
  font-size: 13px;
  margin: 0 0 16px;
}

.spec-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
  padding: 12px;
  background: #fafafa;
  border-radius: 6px;
}

.spec-values {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.value-tag {
  height: auto;
  padding: 4px 8px;
}

.batch-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.sku-uploader {
  width: 48px;
  height: 48px;
  border: 1px dashed #d9d9d9;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.sku-uploader:hover {
  border-color: #409eff;
}

.sku-thumb {
  width: 48px;
  height: 48px;
  object-fit: cover;
  border-radius: 4px;
}

.upload-icon {
  font-size: 18px;
  color: #8c939d;
}
</style>
