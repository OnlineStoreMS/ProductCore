<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  listMaterialCategories,
  listMaterials,
  materialAssetUrl,
  materialListSrc,
  type MaterialCategory,
  type MaterialItem,
} from '../../api/material'
import { uploadImageFromUrlAdapt, type UploadContext } from '../../api/upload'
import type { UploadValidateRules } from '../../utils/uploadValidate'
import { adaptParamsFromRules } from '../../utils/materialAdapt'
import MaterialCropStep from './MaterialCropStep.vue'

const props = defineProps<{
  modelValue: boolean
  max?: number
  rules?: UploadValidateRules
  uploadContext?: UploadContext
}>()

const emit = defineEmits<{
  'update:modelValue': [boolean]
  confirm: [urls: string[]]
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const loading = ref(false)
const applying = ref(false)
const categories = ref<MaterialCategory[]>([])
const selectedCatId = ref(0)
const keyword = ref('')
const page = ref(1)
const total = ref(0)
const materials = ref<MaterialItem[]>([])
const selected = ref<MaterialItem[]>([])
const phase = ref<'pick' | 'crop'>('pick')
const cropQueue = ref<MaterialItem[]>([])
const cropIndex = ref(0)
const cropResults = ref<Map<number, { x: number; y: number; w: number; h: number } | null>>(new Map())

const slots = computed(() => Math.max(1, props.max ?? 1))
const needCrop = computed(() => !!(props.rules?.aspectRatio && props.rules.kind === 'image'))
const aspectLabel = computed(() => props.rules?.aspectRatio?.label || '')
const currentCrop = computed(() => cropQueue.value[cropIndex.value] || null)

const flatCats = computed(() => {
  const out: { id: number; name: string; depth: number }[] = [{ id: 0, name: '全部', depth: 0 }]
  const walk = (nodes: MaterialCategory[], depth: number) => {
    for (const n of nodes || []) {
      out.push({ id: n.id, name: n.name, depth })
      if (n.children?.length) walk(n.children, depth + 1)
    }
  }
  walk(categories.value, 0)
  return out
})

async function loadCategories() {
  try {
    categories.value = await listMaterialCategories()
  } catch (e) {
    ElMessage.error((e as Error).message || '加载分类失败')
  }
}

async function loadMaterials() {
  loading.value = true
  try {
    const res = await listMaterials({
      categoryId: selectedCatId.value || undefined,
      includeChildren: selectedCatId.value ? true : undefined,
      mediaType: 'image',
      keyword: keyword.value.trim() || undefined,
      page: page.value,
      pageSize: 24,
    })
    materials.value = (res.list || []).filter((m) => m.mediaType === 'image')
    total.value = res.total || 0
  } catch (e) {
    ElMessage.error((e as Error).message || '加载素材失败')
  } finally {
    loading.value = false
  }
}

function isSelected(id: number) {
  return selected.value.some((m) => m.id === id)
}

function toggle(m: MaterialItem) {
  const i = selected.value.findIndex((x) => x.id === m.id)
  if (i >= 0) {
    selected.value.splice(i, 1)
    return
  }
  if (selected.value.length >= slots.value) {
    ElMessage.warning(`最多选择 ${slots.value} 张`)
    return
  }
  selected.value.push(m)
}

function thumb(m: MaterialItem) {
  return materialListSrc(m)
}

function startApply(mode: 'center' | 'manual') {
  if (!selected.value.length) {
    ElMessage.warning('请先选择素材')
    return
  }
  if (needCrop.value && mode === 'manual') {
    cropQueue.value = [...selected.value]
    cropIndex.value = 0
    cropResults.value = new Map()
    phase.value = 'crop'
    return
  }
  void applyAll(null)
}

async function applyAll(forcedCrop: { x: number; y: number; w: number; h: number } | null | 'per-item') {
  applying.value = true
  const urls: string[] = []
  try {
    for (const m of selected.value) {
      let crop: { x: number; y: number; w: number; h: number } | undefined
      if (forcedCrop === 'per-item') {
        const c = cropResults.value.get(m.id)
        if (c) crop = c
      } else if (forcedCrop) {
        crop = forcedCrop
      }
      const params = adaptParamsFromRules(materialAssetUrl(m.url), props.rules, crop)
      // 主图场景给更大边长
      if (props.rules?.aspectRatio?.label === '1:1' && (props.max ?? 1) > 1) {
        params.maxEdge = 1200
      }
      const url = await uploadImageFromUrlAdapt(params, props.uploadContext)
      urls.push(url)
    }
    ElMessage.success(`已添加 ${urls.length} 张（已压缩适配电商尺寸）`)
    emit('confirm', urls)
    visible.value = false
  } catch (e) {
    ElMessage.error((e as Error).message || '导入失败')
  } finally {
    applying.value = false
  }
}

function onCropConfirm(crop: { x: number; y: number; w: number; h: number }) {
  const cur = currentCrop.value
  if (!cur) return
  cropResults.value.set(cur.id, crop)
  if (cropIndex.value < cropQueue.value.length - 1) {
    cropIndex.value += 1
    return
  }
  void applyAll('per-item')
}

function onCropCenter() {
  const cur = currentCrop.value
  if (!cur) return
  cropResults.value.set(cur.id, null) // null = backend center
  if (cropIndex.value < cropQueue.value.length - 1) {
    cropIndex.value += 1
    return
  }
  void applyAll('per-item')
}

function onCropCancel() {
  phase.value = 'pick'
}

watch(visible, async (v) => {
  if (!v) {
    phase.value = 'pick'
    selected.value = []
    applying.value = false
    return
  }
  phase.value = 'pick'
  selected.value = []
  page.value = 1
  await loadCategories()
  await loadMaterials()
})

watch([selectedCatId, page], () => {
  if (visible.value && phase.value === 'pick') void loadMaterials()
})

const pager = reactive({
  onSearch() {
    page.value = 1
    void loadMaterials()
  },
})
</script>

<template>
  <el-dialog
    v-model="visible"
    title="从素材中心选取"
    width="860px"
    top="6vh"
    destroy-on-close
    :close-on-click-modal="!applying"
    append-to-body
  >
    <template v-if="phase === 'pick'">
      <p class="hint">
        仅图片；已选 {{ selected.length }}/{{ slots }}。
        <template v-if="needCrop">将按 {{ aspectLabel }} 裁剪并压缩为电商清晰尺寸（原图过大时自动缩小）。</template>
        <template v-else>详情图不强制比例，将限制宽高并压缩体积。</template>
      </p>
      <div class="toolbar">
        <el-select v-model="selectedCatId" style="width: 200px" filterable>
          <el-option
            v-for="c in flatCats"
            :key="c.id"
            :label="'　'.repeat(c.depth) + c.name"
            :value="c.id"
          />
        </el-select>
        <el-input
          v-model="keyword"
          clearable
          placeholder="搜索标题"
          style="width: 200px"
          @keyup.enter="pager.onSearch"
        />
        <el-button type="primary" @click="pager.onSearch">查询</el-button>
      </div>

      <div v-loading="loading" class="grid">
        <div
          v-for="m in materials"
          :key="m.id"
          class="card"
          :class="{ on: isSelected(m.id) }"
          @click="toggle(m)"
        >
          <el-image :src="thumb(m)" fit="cover" class="thumb" lazy />
          <div class="name" :title="m.title">{{ m.title || m.fileName || '未命名' }}</div>
        </div>
        <el-empty v-if="!loading && !materials.length" description="暂无图片素材" />
      </div>

      <div class="pager">
        <el-pagination
          v-model:current-page="page"
          :page-size="24"
          :total="total"
          layout="total, prev, pager, next"
          small
        />
      </div>
    </template>

    <MaterialCropStep
      v-else-if="currentCrop"
      :src="thumb(currentCrop)"
      :aspect-w="rules?.aspectRatio?.w || 1"
      :aspect-h="rules?.aspectRatio?.h || 1"
      :title="`裁剪 ${cropIndex + 1}/${cropQueue.length}：${currentCrop.title || ''}`"
      @confirm="onCropConfirm"
      @skip-center="onCropCenter"
      @cancel="onCropCancel"
    />

    <template #footer>
      <template v-if="phase === 'pick'">
        <el-button :disabled="applying" @click="visible = false">取消</el-button>
        <el-button
          v-if="needCrop"
          type="primary"
          plain
          :loading="applying"
          :disabled="!selected.length"
          @click="startApply('manual')"
        >
          逐张裁剪添加
        </el-button>
        <el-button type="primary" :loading="applying" :disabled="!selected.length" @click="startApply('center')">
          {{ needCrop ? '居中裁剪并添加' : '压缩并添加' }}
        </el-button>
      </template>
    </template>
  </el-dialog>
</template>

<style scoped>
.hint {
  margin: 0 0 10px;
  font-size: 13px;
  color: #909399;
  line-height: 1.5;
}
.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 10px;
  min-height: 240px;
  max-height: 48vh;
  overflow: auto;
}
.card {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  background: #fff;
}
.card.on {
  border-color: #409eff;
  box-shadow: 0 0 0 1px #409eff inset;
}
.thumb {
  width: 100%;
  aspect-ratio: 1;
  display: block;
  background: #f5f7fa;
}
.name {
  padding: 6px 8px;
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.pager {
  margin-top: 10px;
  display: flex;
  justify-content: flex-end;
}
</style>
