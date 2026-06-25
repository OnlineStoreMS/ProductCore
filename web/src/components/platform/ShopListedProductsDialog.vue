<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { fetchShopListedProducts } from '../../api/platform'
import type { PlatformShop } from '../../types/platform'
import type { Product } from '../../types/product'

const props = defineProps<{
  modelValue: boolean
  shop?: PlatformShop | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const loading = ref(false)
const keyword = ref('')
const tableData = ref<Product[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 10

const displayTitle = computed(() => props.shop?.name || '')

async function loadData() {
  if (!props.shop?.id) return
  loading.value = true
  try {
    const data = await fetchShopListedProducts(props.shop.id, {
      keyword: keyword.value.trim() || undefined,
      page: page.value,
      pageSize,
    })
    tableData.value = data.list
    total.value = data.total
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.modelValue, props.shop?.id] as const,
  ([open, id]) => {
    if (open && id) {
      keyword.value = ''
      page.value = 1
      loadData()
    }
  },
)

function onSearch() {
  page.value = 1
  loadData()
}

function onPageChange(p: number) {
  page.value = p
  loadData()
}
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="`已铺货商品 · ${displayTitle}`"
    width="860px"
    destroy-on-close
  >
    <div class="toolbar">
      <el-input
        v-model="keyword"
        placeholder="商品名称 / 资料编码 / 货号"
        clearable
        :prefix-icon="Search"
        class="search-input"
        @keyup.enter="onSearch"
        @clear="onSearch"
      />
      <el-button type="primary" @click="onSearch">搜索</el-button>
    </div>

    <el-table v-loading="loading" :data="tableData" stripe border max-height="460">
      <el-table-column label="商品" min-width="260">
        <template #default="{ row }">
          <div class="product-cell">
            <el-image
              v-if="row.pic"
              :src="row.pic"
              class="pic"
              fit="cover"
            />
            <div v-else class="pic pic-empty">无图</div>
            <div>
              <div class="title">{{ row.name }}</div>
              <div class="meta">资料编码：{{ row.materialCode || '-' }}</div>
              <div class="meta">商品ID：{{ row.id }}</div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="brandName" label="品牌" width="90" />
      <el-table-column prop="categoryName" label="分类" width="100" />
      <el-table-column label="价格" width="90" align="right">
        <template #default="{ row }">
          <span class="price">¥{{ row.price }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="stock" label="库存" width="80" align="center" />
      <el-table-column label="上架" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="row.publishStatus ? 'success' : 'info'" size="small">
            {{ row.publishStatus ? '已上架' : '已下架' }}
          </el-tag>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        background
        layout="total, prev, pager, next"
        :total="total"
        :current-page="page"
        :page-size="pageSize"
        @current-change="onPageChange"
      />
    </div>
  </el-dialog>
</template>

<style scoped>
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.search-input {
  width: 320px;
}

.product-cell {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.pic {
  width: 48px;
  height: 48px;
  border-radius: 4px;
  flex-shrink: 0;
}

.pic-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  color: #909399;
  font-size: 12px;
}

.title {
  font-weight: 500;
  font-size: 14px;
  line-height: 1.4;
}

.meta {
  color: #909399;
  font-size: 12px;
  margin-top: 2px;
}

.price {
  color: #f56c6c;
  font-weight: 600;
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
