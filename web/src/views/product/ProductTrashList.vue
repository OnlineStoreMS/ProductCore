<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, RefreshLeft, Search } from '@element-plus/icons-vue'
import {
  batchForceDeleteProducts,
  batchRestoreProducts,
  fetchBrands,
  fetchCategoryTree,
  fetchTrashedProducts,
  forceDeleteProduct,
  restoreProduct,
} from '../../api/product'
import type { Brand, Category, Product } from '../../types/product'

const loading = ref(false)
const tableData = ref<Product[]>([])
const total = ref(0)
const selectedRows = ref<Product[]>([])
const brands = ref<Brand[]>([])
const categories = ref<Category[]>([])

const query = ref({
  keyword: '',
  brandId: undefined as number | undefined,
  categoryId: undefined as number | undefined,
  page: 1,
  pageSize: 10,
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

async function loadMeta() {
  const [b, c] = await Promise.all([fetchBrands(), fetchCategoryTree()])
  brands.value = b
  categories.value = c
}

async function loadData() {
  loading.value = true
  try {
    const data = await fetchTrashedProducts({
      keyword: query.value.keyword || undefined,
      brandId: query.value.brandId,
      categoryId: query.value.categoryId,
      page: query.value.page,
      pageSize: query.value.pageSize,
    })
    tableData.value = data.list
    total.value = data.total
    selectedRows.value = []
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

watch(query, loadData, { deep: true })

onMounted(async () => {
  await loadMeta()
  await loadData()
})

function onPageChange(page: number) {
  query.value.page = page
}

function onSelectionChange(rows: Product[]) {
  selectedRows.value = rows
}

async function handleRestore(row: Product) {
  try {
    await restoreProduct(row.id)
    ElMessage.success('已恢复')
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '恢复失败')
  }
}

async function handleForceDelete(row: Product) {
  try {
    await forceDeleteProduct(row.id)
    ElMessage.success('已彻底删除')
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}

async function handleBatchRestore() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(`确定恢复选中的 ${selectedRows.value.length} 个商品？`, '批量恢复', {
      type: 'info',
    })
    const { success, failed } = await batchRestoreProducts(selectedRows.value.map((r) => r.id))
    if (success) ElMessage.success(`已恢复 ${success} 个商品`)
    if (failed) ElMessage.warning(`${failed} 个恢复失败`)
    await loadData()
  } catch {
    /* cancelled */
  }
}

async function handleBatchForceDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(
      `彻底删除选中的 ${selectedRows.value.length} 个商品后不可恢复，确定继续？`,
      '批量彻底删除',
      { type: 'warning', confirmButtonText: '彻底删除', confirmButtonClass: 'el-button--danger' },
    )
    const { success, failed } = await batchForceDeleteProducts(selectedRows.value.map((r) => r.id))
    if (success) ElMessage.success(`已彻底删除 ${success} 个商品`)
    if (failed) ElMessage.warning(`${failed} 个删除失败`)
    await loadData()
  } catch {
    /* cancelled */
  }
}
</script>

<template>
  <div class="product-trash">
    <el-card class="filter-card">
      <el-form :inline="true" :model="query">
        <el-form-item label="关键字">
          <el-input
            v-model="query.keyword"
            class="keyword-input"
            placeholder="商品名称 / 资料编码 / 货号 / 商品ID"
            clearable
            :prefix-icon="Search"
          />
        </el-form-item>
        <el-form-item label="品牌">
          <el-select v-model="query.brandId" placeholder="全部" clearable style="width: 140px">
            <el-option v-for="b in brands" :key="b.id" :label="b.name" :value="b.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="query.categoryId" placeholder="全部" clearable style="width: 180px">
            <el-option v-for="c in categoryOptions" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-loading="loading">
      <template #header>
        <span>商品回收站 <el-tag size="small" type="info">{{ total }} 条</el-tag></span>
      </template>

      <el-alert
        type="warning"
        :closable="false"
        show-icon
        class="trash-tip"
        title="回收站中的商品可恢复至商品列表；彻底删除后不可恢复，关联 SKU 与分组关系将一并删除。"
      />

      <div v-if="selectedRows.length" class="batch-bar">
        <span>已选 {{ selectedRows.length }} 项</span>
        <el-button type="primary" :icon="RefreshLeft" @click="handleBatchRestore">批量恢复</el-button>
        <el-button type="danger" :icon="Delete" @click="handleBatchForceDelete">批量彻底删除</el-button>
      </div>

      <el-table
        v-if="tableData.length"
        :data="tableData"
        stripe
        border
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column label="商品信息" min-width="260">
          <template #default="{ row }">
            <div class="product-info">
              <el-image
                v-if="row.pic"
                :src="row.pic"
                class="pic"
                fit="cover"
                :preview-src-list="[row.pic]"
                preview-teleported
                hide-on-click-modal
              />
              <div v-else class="pic pic-empty">无图</div>
              <div>
                <div class="title">{{ row.name }}</div>
                <div class="meta">资料编码：{{ row.materialCode || '-' }}</div>
                <div class="meta">商品ID：{{ row.id }}</div>
                <div v-if="row.productSn" class="meta">货号：{{ row.productSn }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="brandName" label="品牌" width="90" />
        <el-table-column prop="categoryName" label="分类" width="100" />
        <el-table-column prop="source" label="来源" width="80" show-overflow-tooltip>
          <template #default="{ row }">{{ row.source || '-' }}</template>
        </el-table-column>
        <el-table-column label="SKU" width="70" align="center">
          <template #default="{ row }">{{ row.skuCount ?? 0 }}</template>
        </el-table-column>
        <el-table-column prop="deleteTime" label="删除时间" width="160" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-popconfirm title="确定恢复该商品？" @confirm="handleRestore(row)">
              <template #reference>
                <el-button type="primary" link :icon="RefreshLeft">恢复</el-button>
              </template>
            </el-popconfirm>
            <el-popconfirm
              title="彻底删除后不可恢复，确定继续？"
              confirm-button-type="danger"
              @confirm="handleForceDelete(row)"
            >
              <template #reference>
                <el-button type="danger" link :icon="Delete">彻底删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-else-if="!loading" description="回收站为空" :image-size="80" />

      <div v-if="total > query.pageSize" class="pagination">
        <el-pagination
          background
          layout="total, prev, pager, next"
          :total="total"
          :current-page="query.page"
          :page-size="query.pageSize"
          @current-change="onPageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.product-trash {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.filter-card :deep(.el-card__body) {
  padding-bottom: 2px;
}

.keyword-input {
  width: 380px;
}

.trash-tip {
  margin-bottom: 12px;
}

.batch-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
  font-size: 13px;
}

.product-info {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.pic {
  width: 64px;
  height: 64px;
  border-radius: 6px;
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

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
