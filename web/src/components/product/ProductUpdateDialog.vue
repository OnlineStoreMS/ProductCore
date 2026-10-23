<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchPublishShops, startProductPublishUpdate, type PublishShop } from '../../api/product'
import ListingPrepareFields from './ListingPrepareFields.vue'

const props = defineProps<{
  modelValue: boolean
  productId?: number
  productName?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  started: [payload: { productId: number; shopId: number; shopName: string; platformTypeName: string; publishStatus: string }]
}>()

const loading = ref(false)
const starting = ref(false)
const listingTitle = ref('')
const mainImages = ref<string[]>([])
const imagesReady = ref(false)
const hint = ref('')
const shops = ref<PublishShop[]>([])
const selected = ref('')
const platformProductId = ref('')
const filledFromShop = ref('')
const scope = ref<'all' | 'sku'>('all')

watch(
  () => props.modelValue,
  (open) => {
    if (open) void load()
  },
)

watch(selected, (key) => {
  const shop = shops.value.find((s) => shopKey(s) === key)
  const saved = (shop?.platformProductId || '').trim()
  if (!platformProductId.value || platformProductId.value === filledFromShop.value) {
    platformProductId.value = saved
    filledFromShop.value = saved
  }
})

function shopKey(shop: PublishShop) {
  return `${shop.platform}|${shop.platformShopId}`
}

function statusText(status?: string) {
  if (status === 'running') return '铺货中'
  if (status === 'succeeded') return '已铺货'
  if (status === 'failed') return '上次失败'
  return ''
}

async function load() {
  loading.value = true
  selected.value = ''
  shops.value = []
  hint.value = ''
  platformProductId.value = ''
  filledFromShop.value = ''
  scope.value = 'all'
  try {
    const data = await fetchPublishShops(props.productId)
    shops.value = data.list || []
    hint.value = data.hint || ''
    const first = shops.value.find((s) => s.publishReady && s.publishStatus !== 'running')
    if (first) selected.value = shopKey(first)
  } catch (e) {
    ElMessage.error((e as Error).message || '读取店铺失败')
  } finally {
    loading.value = false
  }
}

async function start() {
  const shop = shops.value.find((s) => shopKey(s) === selected.value)
  const productNo = platformProductId.value.trim()
  if (!shop || !props.productId) {
    ElMessage.warning('请选择可以铺货的店铺')
    return
  }
  if (!/^\d{6,32}$/.test(productNo)) {
    ElMessage.warning('请输入店铺里的商品 ID')
    return
  }
  starting.value = true
  try {
    const usePrepared = scope.value === 'all'
    const title = usePrepared ? listingTitle.value.trim() : ''
    const res = await startProductPublishUpdate(props.productId, {
      platform: shop.platform,
      platformShopId: shop.platformShopId,
      platformProductId: productNo,
      scope: scope.value,
      title: title || undefined,
      mainImages: usePrepared && imagesReady.value ? mainImages.value : undefined,
    })
    ElMessage.success(
      scope.value === 'sku'
        ? '已开始更新商品规格，标题和图片保持店铺里现在的内容'
        : title
          ? '已用生成的电商标题开始更新'
          : '已开始更新，标题使用商品现有名称',
    )
    emit('started', {
      productId: props.productId,
      shopId: res.id,
      shopName: res.shopName,
      platformTypeName: res.platformTypeName,
      publishStatus: res.publishStatus,
    })
    emit('update:modelValue', false)
  } catch (e) {
    ElMessage.error((e as Error).message || '更新铺货失败')
  } finally {
    starting.value = false
  }
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    title="更新铺货"
    width="780px"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <ListingPrepareFields
      :active="modelValue"
      :product-id="productId"
      :product-name="productName"
      v-model:title="listingTitle"
      v-model:images="mainImages"
      v-model:ready="imagesReady"
    />
    <div v-loading="loading" class="publish-shops">
      <el-empty v-if="!loading && !shops.length" :description="hint || '没有在线的电商店铺'" />
      <el-radio-group v-else v-model="selected" class="publish-list">
        <el-radio
          v-for="shop in shops"
          :key="shopKey(shop)"
          :value="shopKey(shop)"
          :disabled="!shop.publishReady || shop.publishStatus === 'running'"
          class="publish-item"
        >
          <span class="publish-name">{{ shop.platformShopName || shop.platformShopId }}</span>
          <span class="publish-meta">{{ shop.agentName }}</span>
          <span v-if="statusText(shop.publishStatus)" class="publish-status">{{ statusText(shop.publishStatus) }}</span>
          <span v-if="!shop.publishReady" class="publish-reason">{{ shop.reason }}</span>
        </el-radio>
      </el-radio-group>
      <p v-if="shops.length && hint" class="publish-hint">{{ hint }}</p>
      <label class="product-id">
        <span>商品 ID</span>
        <el-input
          v-model="platformProductId"
          placeholder="店铺里已有商品的 ID"
          maxlength="32"
          clearable
        />
      </label>
      <div class="update-scope">
        <span>更新范围</span>
        <el-radio-group v-model="scope">
          <el-radio value="all">全部信息</el-radio>
          <el-radio value="sku">只更新商品规格</el-radio>
        </el-radio-group>
        <p v-if="scope === 'sku'" class="publish-hint">只重填规格、规格图和价格库存。标题、主图、详情图和发货时间保持不动。</p>
        <p v-else class="publish-hint">填了新标题就用新标题。主图按上面的顺序，最多取前 5 张。</p>
      </div>
    </div>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="starting" :disabled="!selected || !platformProductId.trim()" @click="start">
        开始更新
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.publish-shops {
  min-height: 160px;
  margin-top: 16px;
}

.publish-list {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
  width: 100%;
}

.publish-item {
  margin-right: 0;
  height: auto;
  align-items: flex-start;
  white-space: normal;
}

.publish-name {
  font-weight: 600;
  margin-right: 8px;
}

.publish-meta,
.publish-status,
.publish-reason {
  color: #909399;
  font-size: 12px;
  margin-right: 8px;
}

.publish-reason {
  color: #e6a23c;
}

.publish-hint {
  margin: 12px 0 0;
  color: #909399;
  font-size: 13px;
}

.product-id,
.update-scope {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 16px;
  color: #303133;
}
</style>
