<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElImageViewer, ElMessage } from 'element-plus'
import { retouchProductImage } from '../../api/ai'
import type { UploadContext } from '../../api/upload'
import ImageFileMeta from './ImageFileMeta.vue'

export interface BatchRetouchSource {
  key: string
  group: string
  label: string
  url: string
  uploadContext?: UploadContext
}

interface BatchRow extends BatchRetouchSource {
  selected: boolean
  status: 'idle' | 'running' | 'ok' | 'error'
  resultUrl: string
  error: string
}

const props = defineProps<{
  modelValue: boolean
  items: BatchRetouchSource[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  replace: [rows: { key: string; url: string }[]]
}>()

const prompt = ref('')
const busy = ref(false)
const progress = ref('')
const previewUrl = ref('')
const rows = ref<BatchRow[]>([])
let runId = 0

watch(
  () => props.modelValue,
  (open) => {
    runId += 1
    busy.value = false
    progress.value = ''
    previewUrl.value = ''
    if (!open) return
    prompt.value = ''
    rows.value = props.items.map((item) => ({
      ...item,
      selected: true,
      status: 'idle',
      resultUrl: '',
      error: '',
    }))
  },
)

const selectedCount = computed(() => rows.value.filter((row) => row.selected).length)
const allChecked = computed(() => rows.value.length > 0 && rows.value.every((row) => row.selected))
const someChecked = computed(() => selectedCount.value > 0 && !allChecked.value)
const replaceCount = computed(() => rows.value.filter((row) => row.selected && row.resultUrl).length)

function errorText(e: unknown): string {
  const err = e as { response?: { data?: { message?: string } }; message?: string }
  return err.response?.data?.message || err.message || '修图失败'
}

function toggleAll(checked: boolean | string | number) {
  const on = checked === true
  for (const row of rows.value) {
    if (row.status === 'running') continue
    row.selected = on
  }
}

async function run() {
  const text = prompt.value.trim()
  if (!text) {
    ElMessage.warning('请输入提示词')
    return
  }
  const targets = rows.value.filter((row) => row.selected)
  if (!targets.length) {
    ElMessage.warning('请至少勾选一张图片')
    return
  }
  const id = ++runId
  busy.value = true
  let done = 0
  try {
    for (const row of rows.value) {
      if (id !== runId) return
      if (!row.selected) continue
      row.status = 'running'
      row.error = ''
      row.resultUrl = ''
      progress.value = `${done + 1}/${targets.length}`
      try {
        const data = await retouchProductImage({
          imageUrl: row.url,
          prompt: text,
          scope: row.uploadContext?.scope,
          resource: row.uploadContext?.resource,
          productId: row.uploadContext?.productId,
          skuId: row.uploadContext?.skuId,
        })
        if (id !== runId) return
        row.resultUrl = data.url
        row.status = 'ok'
      } catch (e) {
        if (id !== runId) return
        row.status = 'error'
        row.error = errorText(e)
      }
      done += 1
    }
  } finally {
    if (id === runId) {
      busy.value = false
      progress.value = ''
    }
  }
  if (id !== runId) return
  const ok = rows.value.filter((row) => row.status === 'ok').length
  const failed = rows.value.filter((row) => row.status === 'error').length
  if (failed) ElMessage.warning(`完成 ${ok} 张，失败 ${failed} 张`)
  else ElMessage.success(`已完成 ${ok} 张`)
}

function replaceSelected() {
  const ready = rows.value.filter((row) => row.selected && row.resultUrl)
  if (!ready.length) {
    ElMessage.warning('没有可替换的结果')
    return
  }
  emit(
    'replace',
    ready.map((row) => ({ key: row.key, url: row.resultUrl })),
  )
  emit('update:modelValue', false)
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    title="批量 AI 修图"
    width="980px"
    append-to-body
    destroy-on-close
    :close-on-click-modal="!busy"
    :close-on-press-escape="!busy"
    class="ai-retouch-dialog"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="ai-retouch">
      <div class="preview">
        <div class="list-bar">
          <el-checkbox :model-value="allChecked" :indeterminate="someChecked" :disabled="busy" @change="toggleAll">
            全选
          </el-checkbox>
          <span>已勾选 {{ selectedCount }} / {{ rows.length }}</span>
        </div>
        <div v-for="row in rows" :key="row.key" class="batch-row">
          <div class="row-head">
            <el-checkbox v-model="row.selected" :disabled="row.status === 'running'">
              {{ row.group }} · {{ row.label }}
            </el-checkbox>
            <span v-if="row.status === 'running'" class="state">修图中</span>
            <span v-else-if="row.status === 'ok'" class="state ok">已出图</span>
            <span v-else-if="row.status === 'error'" class="state err">{{ row.error }}</span>
          </div>
          <div class="pair">
            <div class="pane">
              <div class="pane-title">原图</div>
              <img class="zoomable" :src="row.url" alt="原图" title="点击预览" @click="previewUrl = row.url" />
              <ImageFileMeta :url="row.url" />
            </div>
            <div class="pane">
              <div class="pane-title">结果</div>
              <img v-if="row.resultUrl" class="zoomable" :src="row.resultUrl" alt="修图结果" title="点击预览" @click="previewUrl = row.resultUrl" />
              <div v-else class="empty">{{ row.selected ? '修图后显示在这里' : '未勾选，不处理' }}</div>
              <ImageFileMeta v-if="row.resultUrl" :url="row.resultUrl" />
            </div>
          </div>
        </div>
      </div>
      <aside class="tool">
        <div class="tool-title">AI 修图</div>
        <el-input
          v-model="prompt"
          type="textarea"
          :rows="8"
          maxlength="2000"
          show-word-limit
          placeholder="描述要改成什么样，例如把背景换成纯白、去掉杂物、让商品更清晰"
          :disabled="busy"
        />
        <el-button type="primary" :loading="busy" @click="run">{{ busy ? `修图 ${progress}` : '修图' }}</el-button>
        <el-button :disabled="!replaceCount || busy" @click="replaceSelected">替换当前图</el-button>
        <p class="hint">只替换已勾选并且已经出结果的图片。没勾选、没出结果的保持原图。替换后，点页面上的保存才会写进这件商品。</p>
      </aside>
    </div>
    <el-image-viewer v-if="previewUrl" :url-list="[previewUrl]" teleported :z-index="4000" @close="previewUrl = ''" />
  </el-dialog>
</template>

<style scoped>
.ai-retouch {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.preview {
  flex: 1;
  min-width: 0;
  max-height: 70vh;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.list-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  z-index: 1;
  background: #fff;
  padding-bottom: 4px;
  color: #909399;
  font-size: 13px;
}

.batch-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.row-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.state {
  font-size: 12px;
  color: #909399;
}

.state.ok {
  color: #67c23a;
}

.state.err {
  color: #f56c6c;
  max-width: 240px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pair {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.pane {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 8px;
  display: flex;
  flex-direction: column;
  min-height: 220px;
}

.pane-title {
  font-size: 13px;
  color: #606266;
  margin-bottom: 8px;
}

.pane img {
  width: 100%;
  height: 180px;
  object-fit: contain;
  background: #fff;
  border-radius: 4px;
}

.zoomable {
  cursor: zoom-in;
}

.empty {
  height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
  background: #fff;
  border-radius: 4px;
}

.tool {
  width: 260px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  position: sticky;
  top: 0;
}

.tool-title {
  font-weight: 600;
}

.tool :deep(.el-button) {
  margin-left: 0;
}

.hint {
  margin: 0;
  font-size: 12px;
  line-height: 1.5;
  color: #909399;
}
</style>
