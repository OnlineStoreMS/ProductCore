<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchPublishShops, startProductPublish, type PublishShop } from '../../api/product'

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
const hint = ref('')
const shops = ref<PublishShop[]>([])
const selected = ref('')

watch(
  () => props.modelValue,
  (open) => {
    if (open) void load()
  },
)

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
  if (!shop || !props.productId) {
    ElMessage.warning('请选择可以铺货的店铺')
    return
  }
  starting.value = true
  try {
    const res = await startProductPublish(props.productId, {
      platform: shop.platform,
      platformShopId: shop.platformShopId,
    })
    ElMessage.success('已开始铺货，店铺电脑会下载图片并上架')
    emit('started', {
      productId: props.productId,
      shopId: res.id,
      shopName: res.shopName,
      platformTypeName: res.platformTypeName,
      publishStatus: res.publishStatus,
    })
    emit('update:modelValue', false)
  } catch (e) {
    ElMessage.error((e as Error).message || '铺货失败')
  } finally {
    starting.value = false
  }
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    title="铺货到电商店铺"
    width="520px"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <p class="publish-product">{{ productName }}</p>
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
    </div>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="starting" :disabled="!selected" @click="start">开始铺货</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.publish-product {
  margin: 0 0 12px;
  color: #606266;
}

.publish-shops {
  min-height: 120px;
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
</style>
