<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { fetchPlatformShops } from '../../api/platform'
import { fetchProductListings, updateProductListings } from '../../api/product'
import type { PlatformShop } from '../../types/platform'
import type { ListedShop } from '../../types/product'

const props = defineProps<{
  modelValue: boolean
  productId?: number
  productName?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  saved: [productId: number, shops: ListedShop[]]
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const loading = ref(false)
const saving = ref(false)
const keyword = ref('')
const allShops = ref<PlatformShop[]>([])
const selectedIds = ref<number[]>([])

const displayTitle = computed(() => props.productName || `#${props.productId}`)

const filteredShops = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return allShops.value
  return allShops.value.filter((s) => {
    const hay = [s.name, s.shopCode, s.platformTypeName, s.externalShopId]
      .filter(Boolean)
      .join(' ')
      .toLowerCase()
    return hay.includes(kw)
  })
})

const groupedShops = computed(() => {
  const map = new Map<string, PlatformShop[]>()
  for (const shop of filteredShops.value) {
    const key = shop.platformTypeName || '其他'
    const list = map.get(key) || []
    list.push(shop)
    map.set(key, list)
  }
  return [...map.entries()].map(([typeName, shops]) => ({ typeName, shops }))
})

async function loadData() {
  if (!props.productId) return
  loading.value = true
  try {
    const [shopsRes, bound] = await Promise.all([
      fetchPlatformShops({ status: 1, page: 1, pageSize: 500 }),
      fetchProductListings(props.productId),
    ])
    allShops.value = shopsRes.list
    selectedIds.value = bound.map((s) => s.shopId)
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.modelValue, props.productId] as const,
  ([open, id]) => {
    if (open && id) {
      keyword.value = ''
      loadData()
    }
  },
)

async function handleSave() {
  if (!props.productId) return
  saving.value = true
  try {
    const shops = await updateProductListings(props.productId, selectedIds.value)
    ElMessage.success('铺货店铺已更新')
    emit('saved', props.productId, shops)
    visible.value = false
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="`铺货店铺 · ${displayTitle}`"
    width="560px"
    destroy-on-close
  >
    <div v-loading="loading" class="listing-dialog">
      <el-input
        v-model="keyword"
        placeholder="搜索店铺名称 / 编码 / 平台"
        clearable
        :prefix-icon="Search"
        class="search-input"
      />
      <el-scrollbar max-height="420px">
        <div v-if="groupedShops.length === 0" class="empty">暂无可用店铺</div>
        <div v-for="group in groupedShops" :key="group.typeName" class="shop-group">
          <div class="group-title">{{ group.typeName }}</div>
          <el-checkbox-group v-model="selectedIds" class="shop-list">
            <el-checkbox
              v-for="shop in group.shops"
              :key="shop.id"
              :value="shop.id"
              class="shop-item"
            >
              <div class="shop-row">
                <el-image
                  v-if="shop.platformTypeLogo"
                  :src="shop.platformTypeLogo"
                  class="type-logo"
                  fit="cover"
                />
                <span v-else class="type-logo type-logo--text">
                  {{ shop.platformTypeName?.slice(0, 1) || '?' }}
                </span>
                <span class="shop-name">{{ shop.name }}</span>
                <span v-if="shop.shopCode" class="shop-code">{{ shop.shopCode }}</span>
              </div>
            </el-checkbox>
          </el-checkbox-group>
        </div>
      </el-scrollbar>
    </div>
    <template #footer>
      <span class="footer-tip">已选 {{ selectedIds.length }} 家店铺</span>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.listing-dialog {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 200px;
}

.search-input {
  width: 100%;
}

.shop-group + .shop-group {
  margin-top: 16px;
}

.group-title {
  font-size: 13px;
  font-weight: 600;
  color: #606266;
  margin-bottom: 8px;
}

.shop-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.shop-item {
  margin-right: 0;
  height: auto;
  padding: 4px 0;
}

.shop-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.type-logo {
  width: 22px;
  height: 22px;
  border-radius: 4px;
  flex-shrink: 0;
}

.type-logo--text {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
  color: #606266;
  font-size: 12px;
  font-weight: 600;
}

.shop-name {
  font-size: 14px;
}

.shop-code {
  color: #909399;
  font-size: 12px;
}

.empty {
  color: #909399;
  text-align: center;
  padding: 32px 0;
}

.footer-tip {
  float: left;
  line-height: 32px;
  color: #909399;
  font-size: 13px;
}
</style>
