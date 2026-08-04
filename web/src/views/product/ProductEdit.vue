<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import ErpFormRow from '../../components/product/ErpFormRow.vue'
import PictureCardUpload from '../../components/product/PictureCardUpload.vue'
import SkuSpecEditor from '../../components/product/SkuSpecEditor.vue'
import RichTextEditor from '../../components/product/RichTextEditor.vue'
import {
  createDraftProduct,
  fetchBrands,
  fetchCategoryTree,
  fetchGroupTree,
  fetchKeywords,
  fetchProduct,
  updateProduct,
} from '../../api/product'
import type { Category, ProductForm, ProductGroup, ProductKeyword, ProductMedia, SkuItem, SkuSpec } from '../../types/product'
import { compactSkuSpecs, defaultSkuSpecEditorState, normalizeSkuSpecsForEditor, spuWeightFromSkus, validateSkuSpecsNoDuplicateNames, validateSkuSpecsNoDuplicateValues } from '../../types/product'
import { isValidSkuCode } from '../../utils/skuCode'
import { MEDIA_UPLOAD_RULES } from '../../utils/uploadValidate'
import type { UploadContext } from '../../api/upload'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const saving = ref(false)
const autoSaving = ref(false)
const autoSaveHint = ref('')
const isDraftProduct = ref(false)
const hasEditDraft = ref(false)
const isEdit = computed(() => !!route.params.id)
const productId = computed(() => Number(route.params.id) || 0)

const NONE_BRAND_ID = 0
const NONE_CATEGORY_ID = 0

const formReady = ref(false)
/** 检测到修改后防抖保存，避免连续输入时频繁请求 */
const AUTO_SAVE_DEBOUNCE = 2000

let autoSaveTimer: ReturnType<typeof setTimeout> | null = null
let savedPayloadSnapshot = ''
/** 加载/回写快照时跳过变更监听，避免触发保存或循环更新 */
let suppressChangeTracking = false

const productUploadBase = computed((): UploadContext => ({
  scope: 'product',
  productId: productId.value,
}))

function uploadCtx(resource: UploadContext['resource']): UploadContext {
  return { ...productUploadBase.value, resource }
}

const contentRef = ref<HTMLElement>()
const activeNav = ref('#basic')
const showRichDetail = ref(false)

const brands = ref<{ id: number; name: string }[]>([])
const categories = ref<Category[]>([])
const productGroups = ref<ProductGroup[]>([])
const productKeywords = ref<ProductKeyword[]>([])

const form = ref<ProductForm>({
  name: '',
  subTitle: '',
  materialCode: '',
  source: '',
  productSn: '',
  brandId: NONE_BRAND_ID,
  categoryId: NONE_CATEGORY_ID,
  groupIds: [],
  keywordIds: [],
  pic: '',
  albumPics: [],
  price: 0,
  originalPrice: 0,
  stock: 0,
  unit: '件',
  weight: 0,
  publishStatus: 0,
  verifyStatus: 1,
  sort: 0,
  description: '',
  detailHtml: '',
  skuSpecs: [],
  skus: [],
})

const skuSpecs = ref<SkuSpec[]>(defaultSkuSpecEditorState())
const skus = ref<SkuItem[]>([])
const skuSpecEditorRef = ref<InstanceType<typeof SkuSpecEditor>>()
const specValidateTick = ref(0)

const pics34 = ref<string[]>([])
const detailPics = ref<string[]>([])
/** 富文本正文（不含详情图自动生成的 img） */
const richDetailHtml = ref('')
const materialWhite = ref<string[]>([])
const materialTransparent = ref<string[]>([])
const materialGuide34 = ref<string[]>([])
const materialLong = ref<string[]>([])
const video11 = ref<string[]>([])
const video34 = ref<string[]>([])
const video169 = ref<string[]>([])
const video916 = ref<string[]>([])

const mainPicList = computed({
  get: () => {
    const pics: string[] = []
    if (form.value.pic) pics.push(form.value.pic)
    for (const p of form.value.albumPics) {
      if (p && p !== form.value.pic) pics.push(p)
    }
    return pics
  },
  set: (pics: string[]) => {
    form.value.pic = pics[0] || ''
    form.value.albumPics = pics.slice(1)
  },
})

const categoryOptions = computed(() => {
  const flat: { id: number; name: string }[] = []
  function walk(list: Category[], prefix = '') {
    for (const c of list) {
      flat.push({ id: c.id, name: prefix + c.name })
      if (c.children?.length) walk(c.children, prefix + c.name + ' / ')
    }
  }
  walk(categories.value)
  return flat
})

const groupOptions = computed(() => {
  const flat: { id: number; name: string }[] = []
  function walk(list: ProductGroup[], prefix = '') {
    for (const g of list) {
      flat.push({ id: g.id, name: prefix + g.name })
      if (g.children?.length) walk(g.children, prefix + g.name + ' / ')
    }
  }
  walk(productGroups.value)
  return flat
})

const navGroups = [
  {
    title: '商品资料',
    items: [
      { href: '#basic', label: '基础信息' },
      { href: '#media', label: '图文信息' },
    ],
  },
  {
    title: '销售信息',
    items: [{ href: '#sku', label: '商品规格' }],
  },
]

const MEDIA_HINTS = {
  main: '宽高比为1:1，支持jpg/jpeg/png格式；至多可上传10张，可多选上传，拖拽可调整顺序',
  pic34:
    '宽高比为3:4，单张图片大小不超过5M，支持jpg/jpeg/png格式；至多可上传5张，可多选上传，拖拽可调整顺序',
  video: '推荐 mp4 格式；单个视频不超过200M；1:1 / 3:4 / 16:9 / 9:16 视频需符合对应宽高比',
  material:
    '白底图宽高比为1:1，图片大小不超过5M，支持jpg/jpeg/png格式；透明图宽高比为1:1，图片大小不超过3M，支持png格式 3:4导购图图片大小不超过5M，支持jpg/jpeg/png格式；宝贝长图宽高比为2:3，图片大小不超过5M，支持jpg/jpeg/png格式；',
  detail:
    '单张图片建议高/宽≤2，宽度620-1290像素，高度<2000px，单张图片大小不超过5MB，支持jpg/jpeg/png，最多上传50张，可多选上传，支持拖拽调整顺序',
}

function escapeRegExp(text: string): string {
  return text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function parseDetailPics(html: string): string[] {
  if (!html) return []
  const urls: string[] = []
  const re = /<img[^>]+src=["']([^"']+)["']/gi
  let m: RegExpExecArray | null
  while ((m = re.exec(html)) !== null) {
    if (m[1]) urls.push(m[1])
  }
  return urls
}

/** 从 detail_html 中移除由详情图列表生成的 img 段落，保留纯富文本部分 */
function stripDetailPicImages(html: string, pics: string[]): string {
  let result = html.trim()
  if (!result || !pics.length) return result
  for (const src of pics) {
    const escaped = escapeRegExp(src)
    result = result
      .replace(new RegExp(`<p>\\s*<img[^>]*\\ssrc=["']${escaped}["'][^>]*/?>\\s*</p>`, 'gi'), '')
      .replace(new RegExp(`<img[^>]*\\ssrc=["']${escaped}["'][^>]*/?>`, 'gi'), '')
  }
  return result.trim()
}

function buildDetailHtml(pics: string[], richHtml: string): string {
  const imgPart = pics.map((src) => `<p><img src="${src}" alt="" /></p>`).join('')
  const rich = richHtml.trim()
  if (!imgPart && !rich) return ''
  if (!rich) return imgPart
  if (!imgPart) return rich
  return imgPart + rich
}

/** 富文本编辑器展示：详情图 + 富文本正文 */
const editorDetailHtml = computed({
  get: () => buildDetailHtml(detailPics.value, richDetailHtml.value),
  set: (html: string) => {
    richDetailHtml.value = stripDetailPicImages(html, detailPics.value)
  },
})

function buildMedia(): ProductMedia | undefined {
  const media: ProductMedia = {}

  if (pics34.value.length) media.pics34 = [...pics34.value]
  else media.pics34 = []

  const videos: NonNullable<ProductMedia['videos']> = {}
  if (video11.value[0]) videos.ratio11 = video11.value[0]
  if (video34.value[0]) videos.ratio34 = video34.value[0]
  if (video169.value[0]) videos.ratio169 = video169.value[0]
  if (video916.value[0]) videos.ratio916 = video916.value[0]
  if (Object.values(videos).some(Boolean)) media.videos = videos

  const materials: NonNullable<ProductMedia['materials']> = {}
  if (materialWhite.value[0]) materials.white = materialWhite.value[0]
  if (materialTransparent.value[0]) materials.transparent = materialTransparent.value[0]
  if (materialGuide34.value[0]) materials.guide34 = materialGuide34.value[0]
  if (materialLong.value[0]) materials.long = materialLong.value[0]
  if (Object.values(materials).some(Boolean)) media.materials = materials

  if (detailPics.value.length) media.detailPics = [...detailPics.value]

  if (
    !media.pics34?.length &&
    !media.videos &&
    !media.materials &&
    !media.detailPics
  ) {
    return { pics34: [] }
  }
  return media
}

function applyMedia(media?: ProductMedia) {
  pics34.value = media?.pics34?.length ? [...media.pics34] : []
  if (media?.videos) {
    video11.value = media.videos.ratio11 ? [media.videos.ratio11] : []
    video34.value = media.videos.ratio34 ? [media.videos.ratio34] : []
    video169.value = media.videos.ratio169 ? [media.videos.ratio169] : []
    video916.value = media.videos.ratio916 ? [media.videos.ratio916] : []
  } else {
    video11.value = []
    video34.value = []
    video169.value = []
    video916.value = []
  }
  if (media?.materials) {
    materialWhite.value = media.materials.white ? [media.materials.white] : []
    materialTransparent.value = media.materials.transparent ? [media.materials.transparent] : []
    materialGuide34.value = media.materials.guide34 ? [media.materials.guide34] : []
    materialLong.value = media.materials.long ? [media.materials.long] : []
  } else {
    materialWhite.value = []
    materialTransparent.value = []
    materialGuide34.value = []
    materialLong.value = []
  }
  detailPics.value = media?.detailPics?.length ? [...media.detailPics] : []
}

async function loadMeta() {
  const [b, c, g, k] = await Promise.all([
    fetchBrands(),
    fetchCategoryTree(),
    fetchGroupTree(),
    fetchKeywords(),
  ])
  brands.value = b
  categories.value = c
  productGroups.value = g
  productKeywords.value = k
}

async function loadProduct() {
  if (!isEdit.value) return
  loading.value = true
  formReady.value = false
  suppressChangeTracking = true
  if (autoSaveTimer) {
    clearTimeout(autoSaveTimer)
    autoSaveTimer = null
  }
  try {
    const id = Number(route.params.id)
    const product = await fetchProduct(id)
    form.value = {
      name: product.name,
      subTitle: product.subTitle,
      materialCode: product.materialCode || '',
      source: product.source || '',
      productSn: product.productSn,
      brandId: product.brandId || NONE_BRAND_ID,
      categoryId: product.categoryId || NONE_CATEGORY_ID,
      groupIds: product.groupIds || [],
      keywordIds: product.keywordIds || [],
      pic: product.pic,
      albumPics: product.albumPics || [],
      productVideo: product.productVideo,
      price: product.price,
      originalPrice: product.originalPrice,
      stock: product.stock,
      unit: product.unit,
      weight: product.weight,
      publishStatus: product.publishStatus,
      verifyStatus: product.verifyStatus,
      sort: product.sort,
      description: product.description,
      detailHtml: '',
      skuSpecs: product.skuSpecs || [],
      skus: product.skus || [],
    }
    skuSpecs.value = normalizeSkuSpecsForEditor(product.skuSpecs)
    skus.value = JSON.parse(JSON.stringify(product.skus || []))
    applyMedia(product.media)
    if (!product.media?.detailPics?.length) {
      detailPics.value = parseDetailPics(product.detailHtml)
    }
    const richHtml = stripDetailPicImages(product.detailHtml || '', detailPics.value)
    richDetailHtml.value = richHtml
    showRichDetail.value = false
    if (!product.media?.videos?.ratio11 && product.productVideo) {
      video11.value = [product.productVideo]
    }
    isDraftProduct.value = (product.isDraft ?? 0) === 1
    hasEditDraft.value = !!product.hasEditDraft
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
    await nextTick()
    refreshSavedSnapshot()
    formReady.value = true
    await nextTick()
    suppressChangeTracking = false
  }
}

onMounted(async () => {
  await loadMeta()
  if (!isEdit.value) {
    loading.value = true
    try {
      const draft = await createDraftProduct()
      await router.replace(`/products/${draft.id}/edit`)
    } catch (e) {
      ElMessage.error((e as Error).message || '创建草稿失败')
      router.push('/products')
      return
    } finally {
      loading.value = false
    }
  }
  await loadProduct()
  contentRef.value?.addEventListener('scroll', onContentScroll, { passive: true })
})

watch(
  [
    () => ({
      name: form.value.name,
      subTitle: form.value.subTitle,
      materialCode: form.value.materialCode,
      source: form.value.source,
      productSn: form.value.productSn,
      brandId: form.value.brandId,
      categoryId: form.value.categoryId,
      groupIds: form.value.groupIds,
      keywordIds: form.value.keywordIds,
      pic: form.value.pic,
      albumPics: form.value.albumPics,
      unit: form.value.unit,
      sort: form.value.sort,
      publishStatus: form.value.publishStatus,
      description: form.value.description,
    }),
    detailPics,
    pics34,
    materialWhite,
    materialTransparent,
    materialGuide34,
    materialLong,
    video11,
    video34,
    video169,
    video916,
    richDetailHtml,
    skus,
    skuSpecs,
  ],
  onEditChange,
  { deep: true },
)

onUnmounted(() => {
  contentRef.value?.removeEventListener('scroll', onContentScroll)
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
})

/** 只读快照，不调用 syncFormPayload，避免编辑过程中反复改写表单/SKU */
function captureEditSnapshot(): string {
  const f = form.value
  return JSON.stringify({
    name: f.name,
    subTitle: f.subTitle,
    materialCode: f.materialCode,
    source: f.source,
    productSn: f.productSn,
    brandId: f.brandId,
    categoryId: f.categoryId,
    groupIds: f.groupIds,
    keywordIds: f.keywordIds,
    pic: f.pic,
    albumPics: f.albumPics,
    unit: f.unit,
    sort: f.sort,
    publishStatus: f.publishStatus,
    description: f.description,
    detailPics: detailPics.value,
    pics34: pics34.value,
    materialWhite: materialWhite.value,
    materialTransparent: materialTransparent.value,
    materialGuide34: materialGuide34.value,
    materialLong: materialLong.value,
    video11: video11.value,
    video34: video34.value,
    video169: video169.value,
    video916: video916.value,
    richDetailHtml: richDetailHtml.value,
    skus: skus.value,
    skuSpecs: skuSpecs.value,
  })
}

function refreshSavedSnapshot() {
  savedPayloadSnapshot = captureEditSnapshot()
}

function onEditChange() {
  if (suppressChangeTracking || !formReady.value || !isEdit.value || loading.value || autoSaving.value) return
  scheduleAutoSave()
}

function scheduleAutoSave() {
  if (!isEdit.value || loading.value || autoSaving.value) return
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  autoSaveTimer = setTimeout(() => {
    autoSaveTimer = null
    if (suppressChangeTracking || !formReady.value || autoSaving.value) return
    if (captureEditSnapshot() === savedPayloadSnapshot) return
    void persistProduct({ finalize: false, silent: true })
  }, AUTO_SAVE_DEBOUNCE)
}

function syncFormPayload() {
  skuSpecEditorRef.value?.flushSpecsToParent()
  skuSpecEditorRef.value?.ensureAllSkuCodesSilent()
  form.value.skus = skus.value
  form.value.skuSpecs = skus.value.length > 0 ? compactSkuSpecs(skuSpecs.value) : []
  if (skus.value.length) {
    const prices = skus.value.map((s) => s.price).filter(Boolean)
    form.value.price = prices.length ? Math.min(...prices) : form.value.price
    form.value.stock = skus.value.reduce((sum, s) => sum + (s.stock || 0), 0)
    const marketPrices = skus.value
      .map((s) => s.marketPrice)
      .filter((v): v is number => v !== undefined && v > 0)
    if (marketPrices.length) form.value.originalPrice = Math.max(...marketPrices)
    form.value.weight = spuWeightFromSkus(skus.value, form.value.weight)
  }
  form.value.productVideo = video11.value[0] || ''
  form.value.media = buildMedia()
  form.value.detailHtml = buildDetailHtml(detailPics.value, richDetailHtml.value)
}

function validateForm(finalize: boolean): string | null {
  if (finalize && !form.value.name?.trim()) {
    return '请填写商品标题'
  }
  if (!finalize) return null

  const specNameDupErr = validateSkuSpecsNoDuplicateNames(skuSpecs.value)
  if (specNameDupErr) {
    specValidateTick.value++
    skuSpecEditorRef.value?.recomputeDuplicateErrors()
    return specNameDupErr
  }
  if (skus.value.length > 0) {
    const specDupErr = validateSkuSpecsNoDuplicateValues(skuSpecs.value)
    if (specDupErr) {
      specValidateTick.value++
      skuSpecEditorRef.value?.recomputeDuplicateErrors()
      return specDupErr
    }
    if (skus.value.some((s) => !s.skuCode?.trim())) {
      return '请为每条规格生成规格编码'
    }
    if (skus.value.some((s) => !isValidSkuCode(s.skuCode))) {
      return '规格编码只能包含字母和数字'
    }
    const codes = skus.value.map((s) => s.skuCode)
    if (new Set(codes).size !== codes.length) {
      return '规格编码不能重复'
    }
  }
  return null
}

async function persistProduct(opts: { finalize: boolean; silent?: boolean }) {
  if (!productId.value) return
  if (!opts.finalize && captureEditSnapshot() === savedPayloadSnapshot) return
  syncFormPayload()
  const err = validateForm(opts.finalize)
  if (err) {
    if (!opts.silent) {
      ElMessage.warning(err)
      scrollTo(err.includes('规格') ? '#sku' : '#basic')
    }
    return
  }

  const payload = {
    ...form.value,
    brandId: form.value.brandId || NONE_BRAND_ID,
    categoryId: form.value.categoryId || NONE_CATEGORY_ID,
    finalize: opts.finalize,
  }

  if (opts.finalize) saving.value = true
  else autoSaving.value = true

  try {
    await updateProduct(productId.value, payload)
    if (opts.finalize) {
      isDraftProduct.value = false
      hasEditDraft.value = false
      ElMessage.success('商品已保存')
      router.push('/products')
    } else {
      suppressChangeTracking = true
      refreshSavedSnapshot()
      await nextTick()
      suppressChangeTracking = false
      if (isDraftProduct.value) {
        if (!opts.silent) {
          ElMessage.success('已自动保存至草稿箱')
        } else {
          autoSaveHint.value = `已自动保存至草稿箱 ${new Date().toLocaleTimeString()}`
        }
      } else {
        hasEditDraft.value = true
        if (!opts.silent) {
          ElMessage.success('已自动保存编辑草稿')
        } else {
          autoSaveHint.value = `已自动保存编辑草稿 ${new Date().toLocaleTimeString()}`
        }
      }
    }
  } catch (e) {
    if (!opts.silent) {
      ElMessage.error((e as Error).message || '保存失败')
    }
  } finally {
    saving.value = false
    autoSaving.value = false
  }
}

function onContentScroll() {
  const container = contentRef.value
  if (!container) return
  const anchor = container.getBoundingClientRect().top
  const ids = ['#basic', '#media', '#sku']
  for (const id of [...ids].reverse()) {
    const section = container.querySelector(id)
    if (section && section.getBoundingClientRect().top - anchor <= 24) {
      activeNav.value = id
      break
    }
  }
}

async function handleSave() {
  await persistProduct({ finalize: true })
}

function scrollTo(href: string) {
  activeNav.value = href
  const container = contentRef.value
  const section = container?.querySelector(href) as HTMLElement | null
  if (!container || !section) return
  const top =
    section.getBoundingClientRect().top -
    container.getBoundingClientRect().top +
    container.scrollTop
  container.scrollTo({ top: Math.max(0, top - 8), behavior: 'smooth' })
}
</script>

<template>
  <div v-loading="loading" class="product-edit-page">
    <div class="page-body">
      <nav class="page-nav">
        <div v-for="group in navGroups" :key="group.title" class="nav-group">
          <div class="nav-group-title">{{ group.title }}</div>
          <a
            v-for="item in group.items"
            :key="item.href"
            class="nav-link"
            :class="{ active: activeNav === item.href }"
            :href="item.href"
            @click.prevent="scrollTo(item.href)"
          >
            {{ item.label }}
          </a>
        </div>
      </nav>

      <div ref="contentRef" class="page-content">
        <section id="basic" class="content-block">
          <h3 class="block-title">基础信息</h3>
          <div class="form-grid">
            <div class="form-line form-line-full">
              <label class="form-label">分类</label>
              <el-select
                v-model="form.categoryId"
                class="form-control form-control-short"
              >
                <el-option label="无分类" :value="NONE_CATEGORY_ID" />
                <el-option v-for="c in categoryOptions" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
            </div>
            <div class="form-line">
              <label class="form-label">资料编码</label>
              <el-input v-model="form.materialCode" placeholder="来源资料编码，选填" class="form-control" />
            </div>
            <div class="form-line">
              <label class="form-label">商品来源</label>
              <el-select
                v-model="form.source"
                placeholder="请选择或输入"
                filterable
                allow-create
                clearable
                class="form-control"
              >
                <el-option label="淘宝" value="淘宝" />
                <el-option label="抖音" value="抖音" />
                <el-option label="闲鱼" value="闲鱼" />
                <el-option label="小红书" value="小红书" />
              </el-select>
            </div>
            <div class="form-line">
              <label class="form-label">货号</label>
              <el-input v-model="form.productSn" placeholder="选填" class="form-control" />
            </div>
            <div class="form-line">
              <label class="form-label required">商品标题</label>
              <el-input
                v-model="form.name"
                placeholder="请输入"
                maxlength="100"
                show-word-limit
                class="form-control"
              />
            </div>
            <div class="form-line">
              <label class="form-label">导购短标题</label>
              <el-input
                v-model="form.subTitle"
                placeholder="请输入"
                maxlength="100"
                show-word-limit
                class="form-control"
              />
            </div>
            <div class="form-line form-line-full">
              <label class="form-label">商品描述</label>
              <el-input
                v-model="form.description"
                type="textarea"
                :rows="3"
                placeholder="请输入"
                maxlength="500"
                show-word-limit
                class="form-control"
              />
            </div>
            <div class="form-line">
              <label class="form-label">品牌</label>
              <el-select
                v-model="form.brandId"
                class="form-control"
              >
                <el-option label="无品牌" :value="NONE_BRAND_ID" />
                <el-option v-for="b in brands" :key="b.id" :label="b.name" :value="b.id" />
              </el-select>
            </div>
            <div class="form-line form-line-full">
              <label class="form-label">商品分组</label>
              <el-select v-model="form.groupIds" multiple placeholder="请选择" class="form-control">
                <el-option v-for="g in groupOptions" :key="g.id" :label="g.name" :value="g.id" />
              </el-select>
            </div>
            <div class="form-line form-line-full">
              <label class="form-label">关键词</label>
              <el-select
                v-model="form.keywordIds"
                multiple
                filterable
                placeholder="请选择"
                class="form-control"
              >
                <el-option v-for="k in productKeywords" :key="k.id" :label="k.name" :value="k.id" />
              </el-select>
            </div>
            <div class="form-line form-line-inline">
              <label class="form-label">单位</label>
              <el-input v-model="form.unit" style="width: 100px" />
              <label class="form-label form-label-inline">排序</label>
              <el-input-number v-model="form.sort" :min="0" controls-position="right" style="width: 120px" />
              <label class="form-label form-label-inline">上架</label>
              <el-switch v-model="form.publishStatus" :active-value="1" :inactive-value="0" />
            </div>
          </div>
        </section>

        <section id="media" class="content-block">
          <h3 class="block-title">图文信息</h3>

          <ErpFormRow label="商品主图" required :hint="MEDIA_HINTS.main">
            <PictureCardUpload v-model="mainPicList" :max="10" sortable :rules="MEDIA_UPLOAD_RULES.main" :upload-context="uploadCtx('main')" />
          </ErpFormRow>

          <ErpFormRow label="3:4主图" :hint="MEDIA_HINTS.pic34">
            <PictureCardUpload v-model="pics34" :max="5" sortable :rules="MEDIA_UPLOAD_RULES.pic34" :upload-context="uploadCtx('pics34')" />
          </ErpFormRow>

          <ErpFormRow label="商品视频" :hint="MEDIA_HINTS.video">
            <div class="video-ratio-list">
              <div class="ratio-item">
                <PictureCardUpload
                  v-model="video11"
                  :max="1"
                  mode="video"
                  upload-label="上传视频"
                  :rules="MEDIA_UPLOAD_RULES.video11"
                  :upload-context="uploadCtx('video_11')"
                />
                <span>1:1视频</span>
              </div>
              <div class="ratio-item">
                <PictureCardUpload
                  v-model="video34"
                  :max="1"
                  mode="video"
                  upload-label="上传视频"
                  :rules="MEDIA_UPLOAD_RULES.video34"
                  :upload-context="uploadCtx('video_34')"
                />
                <span>3:4视频</span>
              </div>
              <div class="ratio-item">
                <PictureCardUpload
                  v-model="video169"
                  :max="1"
                  mode="video"
                  upload-label="上传视频"
                  :rules="MEDIA_UPLOAD_RULES.video169"
                  :upload-context="uploadCtx('video_169')"
                />
                <span>16:9视频</span>
              </div>
              <div class="ratio-item">
                <PictureCardUpload
                  v-model="video916"
                  :max="1"
                  mode="video"
                  upload-label="上传视频"
                  :rules="MEDIA_UPLOAD_RULES.video916"
                  :upload-context="uploadCtx('video_916')"
                />
                <span>9:16视频</span>
              </div>
            </div>
          </ErpFormRow>

          <ErpFormRow label="素材图片" :hint="MEDIA_HINTS.material">
            <div class="material-list">
              <div class="material-item">
                <PictureCardUpload v-model="materialWhite" :max="1" :rules="MEDIA_UPLOAD_RULES.materialWhite" :upload-context="uploadCtx('material_white')" />
                <span>白底图</span>
              </div>
              <div class="material-item">
                <PictureCardUpload
                  v-model="materialTransparent"
                  :max="1"
                  :rules="MEDIA_UPLOAD_RULES.materialTransparent"
                  :upload-context="uploadCtx('material_transparent')"
                />
                <span>透明图</span>
              </div>
              <div class="material-item">
                <PictureCardUpload v-model="materialGuide34" :max="1" :rules="MEDIA_UPLOAD_RULES.materialGuide34" :upload-context="uploadCtx('material_guide34')" />
                <span>3:4导购图</span>
              </div>
              <div class="material-item">
                <PictureCardUpload v-model="materialLong" :max="1" :rules="MEDIA_UPLOAD_RULES.materialLong" :upload-context="uploadCtx('material_long')" />
                <span>宝贝长图</span>
              </div>
            </div>
          </ErpFormRow>

          <ErpFormRow label="商品详情图" :hint="MEDIA_HINTS.detail">
            <PictureCardUpload v-model="detailPics" :max="50" sortable :rules="MEDIA_UPLOAD_RULES.detail" :upload-context="uploadCtx('detail')" />
          </ErpFormRow>

          <div class="rich-toggle">
            <el-button link type="primary" @click="showRichDetail = !showRichDetail">
              {{ showRichDetail ? '收起富文本' : '富文本详情（可选）' }}
            </el-button>
          </div>
          <ErpFormRow v-if="showRichDetail" label="富文本详情">
            <RichTextEditor v-model="editorDetailHtml" />
          </ErpFormRow>
        </section>

        <section id="sku" class="content-block">
          <h3 class="block-title">销售信息</h3>
          <SkuSpecEditor
            ref="skuSpecEditorRef"
            v-model="skuSpecs"
            v-model:skus="skus"
            :validate-tick="specValidateTick"
            :upload-context="uploadCtx('spec')"
          />
        </section>
      </div>
    </div>

    <footer class="page-footer">
      <div v-if="isEdit" class="footer-hint">
        <span v-if="autoSaving" class="autosave-status">正在自动保存…</span>
        <span v-else-if="autoSaveHint" class="autosave-status">{{ autoSaveHint }}</span>
        <span v-else-if="isDraftProduct" class="autosave-status muted">草稿编辑中，失焦或变更后自动保存至草稿箱</span>
        <span v-else class="autosave-status muted">编辑中，失焦或变更后自动保存编辑草稿（商品列表仍显示已发布版本）</span>
      </div>
      <div class="footer-actions">
        <el-button @click="router.back()">取 消</el-button>
        <el-button type="primary" class="save-btn" :loading="saving" @click="handleSave">
          保存并完成
        </el-button>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.product-edit-page {
  background: #fff;
  border-radius: 4px;
  display: flex;
  flex-direction: column;
  height: calc(100vh - 56px - 40px);
  overflow: hidden;
}

.page-body {
  flex: 1;
  display: flex;
  min-height: 0;
  overflow: hidden;
}

.page-nav {
  width: 132px;
  flex-shrink: 0;
  border-right: 1px solid #f0f0f0;
  padding: 16px 0;
  background: #fafafa;
  overflow-y: auto;
}

.nav-group + .nav-group {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.nav-group-title {
  padding: 4px 16px 8px;
  font-size: 13px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
  white-space: nowrap;
}

.nav-link {
  display: block;
  padding: 8px 16px 8px 20px;
  font-size: 14px;
  color: rgba(0, 0, 0, 0.65);
  text-decoration: none;
  border-left: 2px solid transparent;
  white-space: nowrap;
}

.nav-link:hover,
.nav-link.active {
  color: #ff7700;
  background: #fff7e6;
  border-left-color: #ff7700;
}

.page-content {
  flex: 1;
  min-width: 0;
  min-height: 0;
  padding: 20px 32px 40px;
  font-size: 15px;
  overflow-y: auto;
  overflow-x: hidden;
}

.content-block + .content-block {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #f0f0f0;
}

.block-title {
  margin: 0 0 16px;
  font-size: 16px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px 32px;
  max-width: 1200px;
}

.form-line {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
}

.form-line-full {
  grid-column: 1 / -1;
}

.form-line-inline {
  grid-column: 1 / -1;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px 16px;
}

.form-label {
  width: 120px;
  flex-shrink: 0;
  text-align: right;
  font-size: 15px;
  color: rgba(0, 0, 0, 0.85);
  line-height: 32px;
  white-space: nowrap;
}

.form-label-inline {
  width: auto;
  text-align: left;
  line-height: 1;
}

.form-label.required::before {
  content: '*';
  color: #ff4d4f;
  margin-right: 2px;
}

.form-control {
  flex: 1;
  min-width: 0;
}

.form-control-short {
  flex: 0 0 auto;
  width: 100ch;
  max-width: calc(100% - 132px);
}

.rich-toggle {
  margin: 0 0 8px 132px;
}

.video-ratio-list,
.material-list {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}

.ratio-item,
.material-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: rgba(0, 0, 0, 0.65);
  white-space: nowrap;
}

.page-content :deep(.el-input__inner),
.page-content :deep(.el-textarea__inner) {
  font-size: 14px;
}

.page-content :deep(.erp-form-row) {
  max-width: none;
}

.page-content :deep(.erp-form-label),
.page-content :deep(.form-label) {
  white-space: nowrap;
}

.page-footer {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 24px;
  border-top: 1px solid #f0f0f0;
  background: #fff;
}

.footer-hint {
  flex: 1;
  min-width: 0;
}

.footer-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.autosave-status {
  font-size: 13px;
  color: #67c23a;
}

.autosave-status.muted {
  color: #909399;
}

.save-btn {
  --el-button-bg-color: #ff7700;
  --el-button-border-color: #ff7700;
  --el-button-hover-bg-color: #ff8c1a;
  --el-button-hover-border-color: #ff8c1a;
  --el-button-active-bg-color: #e66a00;
  --el-button-active-border-color: #e66a00;
}

@media (max-width: 900px) {
  .form-grid {
    grid-template-columns: 1fr;
  }

  .rich-toggle {
    margin-left: 0;
  }
}
</style>
