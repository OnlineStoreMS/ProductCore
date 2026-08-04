<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Edit, Delete, View, Upload, Download } from '@element-plus/icons-vue'
import ProductDetailDrawer from '../../components/product/ProductDetailDrawer.vue'
import ProductImportDialog from '../../components/product/ProductImportDialog.vue'
import ProductSkuManageDialog from '../../components/product/ProductSkuManageDialog.vue'
import ProductListingDialog from '../../components/product/ProductListingDialog.vue'
import {
  deleteProduct,
  fetchBrands,
  fetchCategoryTree,
  fetchGroupTree,
  fetchKeywords,
  fetchProducts,
  exportProduct,
  updateProductPublishStatus,
} from '../../api/product'
import type { Brand, Category, ListedShop, Product, ProductGroup, ProductKeyword, ProductSkus } from '../../types/product'

const router = useRouter()
const loading = ref(false)
const tableData = ref<Product[]>([])
const total = ref(0)
const brands = ref<Brand[]>([])
const categories = ref<Category[]>([])
const productGroups = ref<ProductGroup[]>([])
const productKeywords = ref<ProductKeyword[]>([])

const query = ref({
  keyword: '',
  brandId: undefined as number | undefined,
  categoryId: undefined as number | undefined,
  publishStatus: undefined as number | undefined,
  groupId: undefined as number | undefined,
  keywordId: undefined as number | undefined,
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

async function loadData() {
  loading.value = true
  try {
    const data = await fetchProducts({
      keyword: query.value.keyword || undefined,
      brandId: query.value.brandId,
      categoryId: query.value.categoryId,
      groupId: query.value.groupId,
      keywordId: query.value.keywordId,
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

function handleEdit(row: Product) {
  router.push(`/products/${row.id}/edit`)
}

async function handleDelete(row: Product) {
  try {
    await deleteProduct(row.id)
    ElMessage.success('已移入回收站')
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

const detailVisible = ref(false)
const detailProductId = ref<number>()
const importVisible = ref(false)

function handleView(row: Product) {
  detailProductId.value = row.id
  detailVisible.value = true
}

const skuManageVisible = ref(false)
const skuManageProductId = ref<number>()
const skuManageProductName = ref('')

function handleManageSku(row: Product) {
  skuManageProductId.value = row.id
  skuManageProductName.value = row.name
  skuManageVisible.value = true
}

const exportingId = ref<number>()

async function handleExport(row: Product) {
  exportingId.value = row.id
  try {
    await exportProduct(row.id)
    ElMessage.success('导出成功')
  } catch (e) {
    ElMessage.error((e as Error).message || '导出失败')
  } finally {
    exportingId.value = undefined
  }
}

function onSkuSaved(updated: ProductSkus) {
  const row = tableData.value.find((r) => r.id === updated.id)
  if (row) {
    row.skuCount = updated.skuCount
    row.stock = updated.stock
    row.price = updated.price
    row.originalPrice = updated.originalPrice
  }
}

const listingVisible = ref(false)
const listingProductId = ref<number>()
const listingProductName = ref('')

function handleManageListing(row: Product) {
  listingProductId.value = row.id
  listingProductName.value = row.name
  listingVisible.value = true
}

function onListingSaved(productId: number, shops: ListedShop[]) {
  const row = tableData.value.find((r) => r.id === productId)
  if (row) {
    row.listedShops = shops
    row.listedShopCount = shops.length
  }
}
</script>

<template>
  <div class="product-list">
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
        <el-form-item label="分组">
          <el-select v-model="query.groupId" placeholder="全部" clearable style="width: 180px">
            <el-option v-for="g in groupOptions" :key="g.id" :label="g.name" :value="g.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-select v-model="query.keywordId" placeholder="全部" clearable filterable style="width: 160px">
            <el-option v-for="k in productKeywords" :key="k.id" :label="k.name" :value="k.id" />
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
        <div class="header-actions">
          <el-button :icon="Upload" @click="importVisible = true">导入商品</el-button>
        </div>
      </template>

      <el-table :data="tableData" stripe border>
        <el-table-column type="selection" width="45" />
        <el-table-column label="商品信息" min-width="260" fixed>
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
                <div v-if="row.hasEditDraft" class="meta edit-draft-tag">有未发布编辑</div>
                <div v-if="row.productSn" class="meta">货号：{{ row.productSn }}</div>
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
          <template #default="{ row }">
            <el-button link type="primary" class="sku-count-btn" @click="handleManageSku(row)">
              {{ row.skuCount ?? row.skus?.length ?? 0 }}
            </el-button>
          </template>
        </el-table-column>
        <el-table-column prop="stock" label="库存" width="80" align="center" />
        <el-table-column prop="sale" label="销量" width="70" align="center" />
        <el-table-column prop="source" label="商品来源" width="100" show-overflow-tooltip>
          <template #default="{ row }">{{ row.source || '-' }}</template>
        </el-table-column>
        <el-table-column label="已铺货店铺" min-width="180">
          <template #default="{ row }">
            <div class="listing-cell" @click="handleManageListing(row)">
              <template v-if="row.listedShops?.length">
                <div
                  v-for="shop in row.listedShops"
                  :key="shop.shopId"
                  class="shop-row"
                >
                  <el-image
                    v-if="shop.platformTypeLogo"
                    :src="shop.platformTypeLogo"
                    class="type-logo"
                    fit="cover"
                  />
                  <span v-else class="type-logo type-logo--text">
                    {{ shop.platformTypeName?.slice(0, 1) || shop.shopName?.slice(0, 1) || '?' }}
                  </span>
                  <span class="shop-name" :title="shop.shopName">{{ shop.shopName }}</span>
                </div>
              </template>
              <el-button v-else link type="primary" class="listing-btn">绑定店铺</el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="160" />
        <el-table-column label="上架" width="80" align="center">
          <template #default="{ row }">
            <el-switch :model-value="!!row.publishStatus" @change="togglePublish(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="View" @click="handleView(row)">查看</el-button>
            <el-button type="primary" link :icon="Edit" @click="handleEdit(row)">编辑</el-button>
            <el-button
              type="primary"
              link
              :icon="Download"
              :loading="exportingId === row.id"
              @click="handleExport(row)"
            >
              导出
            </el-button>
            <el-popconfirm title="确定将该商品移入回收站？" @confirm="handleDelete(row)">
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

    <ProductDetailDrawer v-model="detailVisible" :product-id="detailProductId" />
    <ProductImportDialog
      v-model="importVisible"
      :brands="brands"
      :category-options="categoryOptions"
      @success="loadData"
    />
    <ProductSkuManageDialog
      v-model="skuManageVisible"
      :product-id="skuManageProductId"
      :product-name="skuManageProductName"
      @saved="onSkuSaved"
    />
    <ProductListingDialog
      v-model="listingVisible"
      :product-id="listingProductId"
      :product-name="listingProductName"
      @saved="onListingSaved"
    />
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

.keyword-input {
  width: 380px;
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

.edit-draft-tag {
  color: #e6a23c;
}

.price {
  color: #f56c6c;
  font-weight: 600;
}

.sku-count-btn {
  font-weight: 600;
  padding: 0;
  min-height: auto;
}

.listing-cell {
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 2px 0;
}

.shop-row {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.type-logo {
  width: 20px;
  height: 20px;
  border-radius: 4px;
  flex-shrink: 0;
}

.type-logo--text {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
  color: #606266;
  font-size: 11px;
  font-weight: 600;
}

.shop-name {
  font-size: 13px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.listing-cell:hover .shop-name {
  color: var(--el-color-primary);
}

.listing-btn {
  padding: 0;
  min-height: auto;
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

.header-actions {
  display: flex;
  gap: 8px;
}
</style>
