<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchProduct, suggestListingTitle } from '../../api/product'
import PictureCardUpload from './PictureCardUpload.vue'
import { MEDIA_UPLOAD_RULES } from '../../utils/uploadValidate'

const props = defineProps<{
  active: boolean
  productId?: number
  productName?: string
}>()

const title = defineModel<string>('title', { default: '' })
const images = defineModel<string[]>('images', { default: () => [] })
const ready = defineModel<boolean>('ready', { default: false })

const generating = ref(false)
const loadingImages = ref(false)

const titleUnits = computed(() => {
  let n = 0
  for (const ch of title.value.trim()) {
    const code = ch.codePointAt(0) || 0
    n += code > 127 ? 2 : 1
  }
  return n
})

const uploadContext = computed(() => ({
  scope: 'product' as const,
  productId: props.productId,
  resource: 'main',
}))

watch(
  () => [props.active, props.productId] as const,
  ([active, id]) => {
    if (!active || !id) return
    title.value = ''
    ready.value = false
    void loadImages(id)
  },
)

function productMainImages(pic?: string, album?: string[]) {
  const seen = new Set<string>()
  const out: string[] = []
  for (const raw of [pic, ...(album || [])]) {
    const url = (raw || '').trim()
    if (!url || seen.has(url)) continue
    seen.add(url)
    out.push(url)
    if (out.length >= 10) break
  }
  return out
}

async function loadImages(id: number) {
  loadingImages.value = true
  images.value = []
  try {
    const product = await fetchProduct(id)
    if (props.productId !== id) return
    images.value = productMainImages(product.pic, product.albumPics)
    ready.value = true
  } catch (e) {
    images.value = []
    ready.value = false
    ElMessage.error((e as Error).message || '读取主图失败')
  } finally {
    loadingImages.value = false
  }
}

async function generateTitle() {
  if (!props.productId) return
  generating.value = true
  try {
    const data = await suggestListingTitle(props.productId)
    title.value = data.title || ''
    if (!title.value) ElMessage.warning('没有生成标题')
  } catch (e) {
    ElMessage.error((e as Error).message || '生成标题失败')
  } finally {
    generating.value = false
  }
}

function shuffleImages() {
  const list = [...images.value]
  if (list.length < 2) {
    ElMessage.warning('至少两张主图才能随机')
    return
  }
  const first = list[0]
  for (let i = list.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    const tmp = list[i]
    list[i] = list[j]
    list[j] = tmp
  }
  if (list[0] === first) {
    const moved = list.shift()
    if (moved) list.push(moved)
  }
  images.value = list
}
</script>

<template>
  <div class="listing-prepare">
    <p class="publish-product">{{ productName }}</p>
    <div>
      <el-button size="small" :loading="generating" @click="generateTitle">AI 生成标题</el-button>
    </div>
    <el-input
      v-model="title"
      type="textarea"
      :rows="2"
      placeholder="新标题显示在这里，可修改。不超过 60 个字符，汉字算 2"
    />
    <p v-if="title.trim()" class="title-count" :class="{ 'title-count--over': titleUnits > 60 }">{{ titleUnits }}/60</p>
    <div class="main-images">
      <div class="main-images-head">
        <span>主图</span>
        <el-button size="small" :disabled="loadingImages || images.length < 2" @click="shuffleImages">随机主图</el-button>
      </div>
      <div v-loading="loadingImages">
        <PictureCardUpload
          v-if="images.length || !loadingImages"
          v-model="images"
          :max="10"
          sortable
          ai-retouch
          :material-pick="false"
          :rules="MEDIA_UPLOAD_RULES.main"
          :upload-context="uploadContext"
        />
        <p v-if="!loadingImages && !images.length" class="image-hint">这个商品还没有主图</p>
      </div>
      <p class="image-hint">点图片可 AI 修图，悬停后可擦除，也可拖动排序。抖店最多使用前 5 张，第一张是主图。</p>
    </div>
  </div>
</template>

<style scoped>
.listing-prepare {
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: #303133;
}

.publish-product {
  margin: 0;
  color: #303133;
  line-height: 1.5;
  word-break: break-word;
}

.title-count {
  margin: 0;
  color: #909399;
  font-size: 12px;
  text-align: right;
}

.title-count--over {
  color: #e6a23c;
}

.main-images {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 10px;
}

.main-images-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.image-hint {
  margin: 0;
  color: #909399;
  font-size: 12px;
  line-height: 1.5;
}
</style>
