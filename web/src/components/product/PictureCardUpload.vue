<script setup lang="ts">
import { computed, ref } from 'vue'
import { Plus, VideoCamera, View, Picture } from '@element-plus/icons-vue'
import { ElImageViewer, ElMessage } from 'element-plus'
import type { UploadFile, UploadInstance, UploadRequestOptions, UploadUserFile } from 'element-plus'
import { uploadImage, uploadImagesBatch, uploadVideo } from '../../api/upload'
import type { UploadContext } from '../../api/upload'
import type { UploadValidateRules } from '../../utils/uploadValidate'
import { acceptFromRules, validateUploadFile } from '../../utils/uploadValidate'

const props = withDefaults(
  defineProps<{
    modelValue: string[]
    max?: number
    accept?: string
    uploadLabel?: string
    mode?: 'image' | 'video'
    disabled?: boolean
    size?: 'default' | 'small' | 'inline'
    sortable?: boolean
    rules?: UploadValidateRules
    uploadContext?: UploadContext
  }>(),
  {
    max: 10,
    uploadLabel: '上传图片',
    mode: 'image',
    disabled: false,
    size: 'default',
    sortable: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const list = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const sizeClass = computed(() => ({
  small: props.size === 'small',
  inline: props.size === 'inline',
}))

const canSort = computed(() => props.sortable && props.mode === 'image' && !props.disabled && props.max > 1)

const canSelectMultiple = computed(
  () => props.mode === 'image' && props.max > 1 && list.value.length < props.max,
)

const useBatchUpload = computed(() => canSelectMultiple.value)

const uploadRef = ref<UploadInstance>()
const batchUploading = ref(false)
let batchCollectTimer: ReturnType<typeof setTimeout> | null = null

function fileIdentity(file: File): string {
  return `${file.name}:${file.size}:${file.lastModified}`
}

function dedupeFiles(files: File[]): File[] {
  const seen = new Set<string>()
  const out: File[] = []
  for (const file of files) {
    const key = fileIdentity(file)
    if (seen.has(key)) continue
    seen.add(key)
    out.push(file)
  }
  return out
}

async function uploadOneFile(file: File): Promise<string | null> {
  if (effectiveRules.value) {
    const err = await validateUploadFile(file, effectiveRules.value)
    if (err) {
      ElMessage.warning(`${file.name}: ${err}`)
      return null
    }
  }
  return props.mode === 'video' ? await uploadVideo(file, props.uploadContext) : await uploadImage(file, props.uploadContext)
}

function handleUpload(options: UploadRequestOptions) {
  const file = options.file as File
  return uploadOneFile(file)
    .then((url) => {
      if (!url) {
        options.onError?.(new Error('upload failed') as never)
        return
      }
      if (list.value.length >= props.max) {
        ElMessage.warning(`最多上传 ${props.max} ${props.mode === 'video' ? '个' : '张'}`)
        options.onError?.(new Error('upload limit exceeded') as never)
        return
      }
      if (list.value.includes(url)) {
        options.onSuccess?.({} as never)
        return
      }
      list.value = [...list.value, url]
      ElMessage.success('上传成功')
      options.onSuccess?.({} as never)
    })
    .catch((e) => {
      ElMessage.error((e as Error).message || '上传失败')
      options.onError?.(e as never)
      throw e
    })
}

function handleFileChange(_uploadFile: UploadFile, uploadFiles: UploadUserFile[]) {
  if (!useBatchUpload.value || batchUploading.value) return
  if (batchCollectTimer) clearTimeout(batchCollectTimer)
  batchCollectTimer = setTimeout(() => {
    batchCollectTimer = null
    void processBatchUpload([...uploadFiles])
  }, 30)
}

async function processBatchUpload(uploadFiles: UploadUserFile[]) {
  const slots = props.max - list.value.length
  if (slots <= 0) {
    ElMessage.warning(`最多上传 ${props.max} 张`)
    uploadRef.value?.clearFiles()
    return
  }

  const rawFiles: File[] = []
  for (const item of uploadFiles) {
    if (item.raw instanceof File) rawFiles.push(item.raw)
  }
  // 须在读取 raw 之后再 clear，否则 el-upload 会清空 raw 引用
  uploadRef.value?.clearFiles()

  const files = dedupeFiles(rawFiles).slice(0, slots)
  if (!files.length) {
    ElMessage.warning('未读取到有效图片，请重试')
    return
  }

  if (rawFiles.length > slots) {
    ElMessage.warning(`最多还能上传 ${slots} 张，已按选取顺序截取`)
  }

  batchUploading.value = true
  try {
    const validFiles: File[] = []
    for (const file of files) {
      if (effectiveRules.value) {
        const err = await validateUploadFile(file, effectiveRules.value)
        if (err) {
          ElMessage.warning(`${file.name}: ${err}`)
          continue
        }
      }
      validFiles.push(file)
    }
    if (!validFiles.length) return

    let urls: string[] = []
    let failed: { filename: string; message: string }[] = []
    try {
      const result = await uploadImagesBatch(validFiles, props.uploadContext)
      urls = result.urls ?? []
      failed = result.failed ?? []
    } catch {
      // 兼容未重启后端、批量接口不可用时的降级
      for (const file of validFiles) {
        try {
          urls.push(await uploadImage(file, props.uploadContext))
        } catch (err) {
          failed.push({ filename: file.name, message: (err as Error).message || '上传失败' })
        }
      }
    }
    for (const item of failed) {
      ElMessage.warning(`${item.filename}: ${item.message}`)
    }
    const next = [...list.value]
    for (const url of urls) {
      if (next.length >= props.max) break
      if (!next.includes(url)) next.push(url)
    }
    list.value = next

    if (urls.length > 0) {
      ElMessage.success(urls.length === 1 ? '上传成功' : `已成功上传 ${urls.length} 张图片`)
    }
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    batchUploading.value = false
  }
}

const effectiveAccept = computed(() => {
  if (props.rules) return acceptFromRules(props.rules)
  if (props.mode === 'video') return 'video/mp4,.mp4'
  if (props.accept) return props.accept
  return 'image/jpeg,image/png,image/jpg'
})

const effectiveRules = computed(() => {
  if (props.rules) return props.rules
  if (props.mode === 'video') {
    return { kind: 'video' as const, maxSizeMB: 200, acceptExt: ['mp4'] }
  }
  return undefined
})

const dragFrom = ref<number | null>(null)
const dragOverIndex = ref<number | null>(null)
const videoPreviewUrl = ref('')
const imageViewerVisible = ref(false)
const imageViewerIndex = ref(0)

const videoPreviewVisible = computed({
  get: () => !!videoPreviewUrl.value,
  set: (open: boolean) => {
    if (!open) videoPreviewUrl.value = ''
  },
})

function removeAt(index: number) {
  const next = [...list.value]
  next.splice(index, 1)
  list.value = next
}

function openVideoPreview(url: string) {
  videoPreviewUrl.value = url
}

function openImagePreview(index: number) {
  imageViewerIndex.value = index
  imageViewerVisible.value = true
}

function closeImagePreview() {
  imageViewerVisible.value = false
}

function onDragStart(index: number) {
  if (!canSort.value) return
  dragFrom.value = index
}

function onDragOver(index: number, e: DragEvent) {
  if (!canSort.value || dragFrom.value === null) return
  e.preventDefault()
  dragOverIndex.value = index
}

function onDrop(index: number) {
  if (!canSort.value || dragFrom.value === null) {
    resetDrag()
    return
  }
  const from = dragFrom.value
  if (from !== index) {
    const next = [...list.value]
    const [moved] = next.splice(from, 1)
    next.splice(index, 0, moved)
    list.value = next
  }
  resetDrag()
}

function resetDrag() {
  dragFrom.value = null
  dragOverIndex.value = null
}
</script>

<template>
  <div class="picture-card-upload" :class="sizeClass">
    <div
      v-for="(url, i) in list"
      :key="`${url}-${i}`"
      class="picture-card filled"
      :class="{
        sortable: canSort,
        'drag-over': dragOverIndex === i && dragFrom !== null && dragFrom !== i,
      }"
      :draggable="canSort"
      @dragstart="onDragStart(i)"
      @dragover="onDragOver(i, $event)"
      @drop="onDrop(i)"
      @dragend="resetDrag"
    >
      <div v-if="mode === 'video'" class="preview-media-wrap">
        <video :src="url" class="preview" muted preload="metadata" />
        <button
          type="button"
          class="preview-eye-btn"
          title="预览"
          @click.stop="openVideoPreview(url)"
        >
          <el-icon><View /></el-icon>
        </button>
      </div>
      <div v-else class="preview-media-wrap">
        <el-image :src="url" fit="cover" class="preview preview-image" />
        <button
          type="button"
          class="preview-eye-btn"
          title="预览"
          @click.stop="openImagePreview(i)"
        >
          <el-icon><View /></el-icon>
        </button>
      </div>
      <button v-if="!disabled" type="button" class="remove-btn" @click.stop="removeAt(i)">×</button>
    </div>
    <el-upload
      v-if="list.length < max && !disabled"
      ref="uploadRef"
      :show-file-list="false"
      :auto-upload="!useBatchUpload"
      :http-request="useBatchUpload ? undefined : handleUpload"
      :accept="effectiveAccept"
      :multiple="canSelectMultiple"
      :disabled="batchUploading"
      class="upload-wrap"
      @change="handleFileChange"
    >
      <div class="picture-card add" :title="uploadLabel">
        <el-icon class="upload-icon">
          <VideoCamera v-if="mode === 'video'" />
          <Picture v-else-if="size === 'inline'" />
          <Plus v-else />
        </el-icon>
        <span class="upload-text">{{ uploadLabel }}</span>
      </div>
    </el-upload>

    <el-image-viewer
      v-if="imageViewerVisible"
      :url-list="list"
      :initial-index="imageViewerIndex"
      teleported
      @close="closeImagePreview"
    />

    <el-dialog
      v-model="videoPreviewVisible"
      title="视频预览"
      width="720px"
      destroy-on-close
      append-to-body
      class="video-preview-dialog"
      @closed="videoPreviewUrl = ''"
    >
      <video
        v-if="videoPreviewUrl"
        :src="videoPreviewUrl"
        class="video-preview-player"
        controls
        autoplay
      />
    </el-dialog>
  </div>
</template>

<style scoped>
.picture-card-upload {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.picture-card {
  position: relative;
  width: 88px;
  height: 88px;
  border: 1px dashed #d9d9d9;
  border-radius: 4px;
  background: #fafafa;
  overflow: hidden;
  box-sizing: border-box;
}

.picture-card-upload.small .picture-card {
  width: 52px;
  height: 52px;
}

.picture-card-upload.inline .picture-card {
  width: 32px;
  height: 32px;
}

.picture-card-upload.inline {
  gap: 0;
}

.upload-wrap {
  width: 88px;
  height: 88px;
  line-height: normal;
}

.upload-wrap :deep(.el-upload) {
  position: relative;
  width: 100%;
  height: 100%;
  display: block;
  outline: none;
}

.upload-wrap :deep(.el-upload__input) {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  cursor: pointer;
  z-index: 1;
}

.picture-card-upload.small .upload-wrap {
  width: 52px;
  height: 52px;
}

.picture-card-upload.inline .upload-wrap {
  width: 32px;
  height: 32px;
}

.picture-card.add {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  width: 100%;
  height: 100%;
  cursor: pointer;
  color: rgba(0, 0, 0, 0.45);
  transition: border-color 0.2s, color 0.2s;
  box-sizing: border-box;
  border: 1px dashed #d9d9d9;
  border-radius: 4px;
  background: #fafafa;
  pointer-events: none;
}

.picture-card-upload.small .picture-card.add {
  gap: 0;
}

.picture-card.add:hover {
  border-color: #ff7700;
  color: #ff7700;
}

.upload-wrap:hover .picture-card.add {
  border-color: #ff7700;
  color: #ff7700;
}

.picture-card.filled {
  border-style: solid;
  border-color: #e8e8e8;
}

.picture-card.sortable {
  cursor: grab;
}

.picture-card.sortable:active {
  cursor: grabbing;
}

.picture-card.drag-over {
  border-color: #ff7700;
  box-shadow: 0 0 0 2px rgba(255, 119, 0, 0.25);
}

.preview {
  width: 100%;
  height: 100%;
  display: block;
}

.preview-image {
  pointer-events: none;
}

.preview-image :deep(.el-image__inner) {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-media-wrap {
  position: relative;
  width: 100%;
  height: 100%;
}

.preview-media-wrap .preview {
  object-fit: cover;
}

.preview-eye-btn {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 24px;
  height: 24px;
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

.preview-eye-btn .el-icon {
  font-size: 14px;
}

.picture-card.filled:hover .preview-eye-btn {
  opacity: 1;
}

.preview-eye-btn:hover {
  background: rgba(0, 0, 0, 0.7);
}

.picture-card-upload.small .preview-eye-btn {
  width: 20px;
  height: 20px;
}

.picture-card-upload.small .preview-eye-btn .el-icon {
  font-size: 12px;
}

.picture-card-upload.inline .preview-eye-btn {
  width: 16px;
  height: 16px;
}

.picture-card-upload.inline .preview-eye-btn .el-icon {
  font-size: 10px;
}

.video-preview-player {
  display: block;
  width: 100%;
  max-height: 70vh;
  background: #000;
  border-radius: 4px;
}

.upload-icon {
  font-size: 20px;
}

.picture-card-upload.small .upload-icon {
  font-size: 16px;
}

.picture-card-upload.inline .upload-icon {
  font-size: 14px;
}

.picture-card-upload.inline .picture-card.add {
  gap: 0;
  background: #fff;
}

.upload-text {
  font-size: 13px;
  line-height: 1.2;
  text-align: center;
  padding: 0 4px;
}

.picture-card-upload.small .upload-text {
  display: none;
}

.picture-card-upload.inline .upload-text {
  display: none;
}

.remove-btn {
  position: absolute;
  top: 2px;
  right: 2px;
  width: 18px;
  height: 18px;
  border: none;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  font-size: 12px;
  line-height: 1;
  cursor: pointer;
  padding: 0;
  z-index: 1;
}

.picture-card-upload.inline .remove-btn {
  width: 14px;
  height: 14px;
  font-size: 10px;
  top: 0;
  right: 0;
}
</style>
