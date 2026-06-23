<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, Search, Edit, Delete } from '@element-plus/icons-vue'
import {
  deleteProduct,
  fetchBrands,
  fetchCategoryTree,
  fetchGroups,
  fetchProducts,
  updateProductPublishStatus,
} from '../../api/product'
import type { Brand, Category, Product, ProductGroup } from '../../types/product'

const router = useRouter()
const loading = ref(false)
const tableData = ref<Product[]>([])
const total = ref(0)
const brands = ref<Brand[]>([])
const categories = ref<Category[]>([])
const productGroups = ref<ProductGroup[]>([])

const query = ref({
  keyword: '',
  brandId: undefined as number | undefined,
  categoryId: undefined as number | undefined,
  publishStatus: undefined as number | undefined,
  groupId: undefined as number | undefined,
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
  const [b, c, g] = await Promise.all([fetchBrands(), fetchCategoryTree(), fetchGroups()])
  brands.value = b
  categories.value = c
  productGroups.value = g
}

async function loadData() {
  loading.value = true
  try {
    const data = await fetchProducts({
      keyword: query.value.keyword || undefined,
      brandId: query.value.brandId,
      categoryId: query.value.categoryId,
      groupId: query.value.groupId,
      publishStatus: query.value.publishStatus,
      page: query.value.page,
      pageSize: query.value.pageSize,
    })
    tableData.value = data.list.map((p) => ({ ...p, skus: p.skus || [] }))
    total.value = data.total
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

function handleCreate() {
  router.push('/products/create')
}

function handleEdit(row: Product) {
  router.push(`/products/${row.id}/edit`)
}

async function handleDelete(row: Product) {
  try {
    await deleteProduct(row.id)
    ElMessage.success('已删除')
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}

async function togglePublish(row: Product) {
  const next = row.publishStatus ? 0 : 1
  try {
    await updateProductPublishStatus(row.id, next as 0 | 1)
    row.publishStatus = next as 0 | 1
    ElMessage.success(next ? '已上架' : '已下架')
  } catch (e) {
    ElMessage.error((e as Error).message || '操作失败')
  }
}

function onPageChange(page: number) {
  query.value.page = page
}
</script>

<template>
  <div class="product-list">
    <el-card class="filter-card">
      <el-form :inline="true" :model="query">
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="商品名称 / 货号" clearable :prefix-icon="Search" />
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
        <el-form-item label="分组">
          <el-select v-model="query.groupId" placeholder="全部" clearable style="width: 140px">
            <el-option v-for="g in productGroups" :key="g.id" :label="g.name" :value="g.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="上架状态">
          <el-select v-model="query.publishStatus" placeholder="全部" clearable style="width: 120px">
            <el-option label="已上架" :value="1" />
            <el-option label="已下架" :value="0" />
          </el-select>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-loading="loading">
      <template #header>
        <span>商品列表 <el-tag size="small" type="info">{{ total }} 条</el-tag></span>
        <el-button type="primary" :icon="Plus" @click="handleCreate">添加商品</el-button>
      </template>

      <el-table :data="tableData" stripe border>
        <el-table-column type="selection" width="45" />
        <el-table-column label="商品信息" min-width="260" fixed>
          <template #default="{ row }">
            <div class="product-info">
              <el-image :src="row.pic" class="pic" fit="cover" :preview-src-list="[row.pic]" />
              <div>
                <div class="title">{{ row.name }}</div>
                <div class="meta">货号：{{ row.productSn }}</div>
                <div class="meta">{{ row.subTitle }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="brandName" label="品牌" width="90" />
        <el-table-column prop="categoryName" label="分类" width="100" />
        <el-table-column label="价格" width="100" align="right">
          <template #default="{ row }">
            <span class="price">¥{{ row.price }}</span>
          </template>
        </el-table-column>
        <el-table-column label="SKU" width="70" align="center">
          <template #default="{ row }">{{ row.skuCount ?? row.skus?.length ?? '-' }}</template>
        </el-table-column>
        <el-table-column prop="stock" label="库存" width="80" align="center" />
        <el-table-column prop="sale" label="销量" width="70" align="center" />
        <el-table-column label="上架" width="80" align="center">
          <template #default="{ row }">
            <el-switch :model-value="!!row.publishStatus" @change="togglePublish(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="Edit" @click="handleEdit(row)">编辑</el-button>
            <el-popconfirm title="确定删除该商品？" @confirm="handleDelete(row)">
              <template #reference>
                <el-button type="danger" link :icon="Delete">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
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
.product-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.filter-card :deep(.el-card__body) {
  padding-bottom: 2px;
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

:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
