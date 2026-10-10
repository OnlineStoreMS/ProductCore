<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElImageViewer, ElMessage } from 'element-plus'
import { retouchProductImage } from '../../api/ai'
import type { UploadContext } from '../../api/upload'
import ImageFileMeta from './ImageFileMeta.vue'

const props = defineProps<{
  modelValue: boolean
  source: string
  uploadContext?: UploadContext
  canAdd: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  replace: [url: string]
  add: [url: string]
}>()

const prompt = ref('')
const resultUrl = ref('')
const busy = ref(false)
const previewUrl = ref('')

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    prompt.value = ''
    resultUrl.value = ''
    busy.value = false
    previewUrl.value = ''
  },
)

function errorText(e: unknown): string {
  const err = e as { response?: { data?: { message?: string } }; message?: string }
  return err.response?.data?.message || err.message || '修图失败'
}

async function run() {
  const text = prompt.value.trim()
  if (!text) {
    ElMessage.warning('请输入提示词')
    return
  }
  if (!props.source) {
    ElMessage.warning('没有原图')
    return
  }
  busy.value = true
  try {
    const data = await retouchProductImage({
      imageUrl: props.source,
      prompt: text,
      scope: props.uploadContext?.scope,
      resource: props.uploadContext?.resource,
      productId: props.uploadContext?.productId,
      skuId: props.uploadContext?.skuId,
    })
    resultUrl.value = data.url
  } catch (e) {
    ElMessage.error(errorText(e))
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    title="查看图片"
    width="960px"
    append-to-body
    destroy-on-close
    class="ai-retouch-dialog"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="ai-retouch">
      <div class="preview">
        <div class="pane">
          <div class="pane-title">原图</div>
          <img v-if="source" class="zoomable" :src="source" alt="原图" title="点击预览" @click="previewUrl = source" />
          <ImageFileMeta v-if="source" :url="source" />
        </div>
        <div class="pane">
          <div class="pane-title">结果</div>
          <img v-if="resultUrl" class="zoomable" :src="resultUrl" alt="修图结果" title="点击预览" @click="previewUrl = resultUrl" />
          <div v-else class="empty">修图后显示在这里</div>
          <ImageFileMeta v-if="resultUrl" :url="resultUrl" />
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
        <el-button type="primary" :loading="busy" @click="run">修图</el-button>
        <el-button :disabled="!resultUrl || busy" @click="emit('replace', resultUrl)">替换当前图</el-button>
        <el-button :disabled="!resultUrl || !canAdd || busy" @click="emit('add', resultUrl)">添加保存</el-button>
        <p class="hint">替换或添加后，点页面上的保存才会写进这件商品。</p>
      </aside>
    </div>
    <el-image-viewer v-if="previewUrl" :url-list="[previewUrl]" teleported :z-index="4000" @close="previewUrl = ''" />
  </el-dialog>
</template>

<style scoped>
.ai-retouch {
  display: flex;
  gap: 16px;
  min-height: 420px;
}

.preview {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  min-width: 0;
}

.pane {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 8px;
  display: flex;
  flex-direction: column;
  min-height: 360px;
}

.pane-title {
  font-size: 13px;
  color: #606266;
  margin-bottom: 8px;
}

.pane img {
  width: 100%;
  flex: 1;
  object-fit: contain;
  background: #fff;
  border-radius: 4px;
}

.zoomable {
  cursor: zoom-in;
}

.empty {
  flex: 1;
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
