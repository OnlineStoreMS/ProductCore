<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { superSearch } from '../../api/search'
import type { SuperSearchItem } from '../../types/search'
import ProductDetailDrawer from '../product/ProductDetailDrawer.vue'
import ProductSkuManageDialog from '../product/ProductSkuManageDialog.vue'
import ProductListingDialog from '../product/ProductListingDialog.vue'
import type { ListedShop } from '../../types/product'

const keyword = ref('')
const panelOpen = ref(false)
const loading = ref(false)
const items = ref<SuperSearchItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20

const rootRef = ref<HTMLElement>()
let debounceTimer: ReturnType<typeof setTimeout> | undefined

const detailVisible = ref(false)
const detailProductId = ref<number>()
const skuManageVisible = ref(false)
const skuManageProductId = ref<number>()
const skuManageProductName = ref('')
const listingVisible = ref(false)
const listingProductId = ref<number>()
const listingProductName = ref('')

function resolvePic(item: SuperSearchItem): string {
  return item.pic || item.productPic || ''
}

async function runSearch(resetPage = true) {
  const q = keyword.value.trim()
  if (!q) {
    items.value = []
    total.value = 0
    return
  }
  if (resetPage) {
    page.value = 1
  }
  loading.value = true
  try {
    const data = await superSearch({ keyword: q, page: page.value, pageSize })
    items.value = data.list
    total.value = data.total
    panelOpen.value = true
  } catch (e) {
    ElMessage.error((e as Error).message || '搜索失败')
  } finally {
    loading.value = false
  }
}

function onFocus() {
  if (keyword.value.trim()) {
    panelOpen.value = true
  }
}

function onEnter() {
  void runSearch(true)
}

function onClear() {
  items.value = []
  total.value = 0
  panelOpen.value = false
}

function onPageChange(p: number) {
  page.value = p
  void runSearch(false)
}

function openDetail(item: SuperSearchItem) {
  detailProductId.value = item.productId
  detailVisible.value = true
  panelOpen.value = false
}

function openSkuManage(item: SuperSearchItem) {
  skuManageProductId.value = item.productId
  skuManageProductName.value = item.productName
  skuManageVisible.value = true
  panelOpen.value = false
}

function openListing(item: SuperSearchItem) {
  listingProductId.value = item.productId
  listingProductName.value = item.productName
  listingVisible.value = true
  panelOpen.value = false
}

function onListingSaved(_productId: number, _shops: ListedShop[]) {
  if (keyword.value.trim()) {
    void runSearch(false)
  }
}

function onDocumentClick(e: MouseEvent) {
  if (!rootRef.value?.contains(e.target as Node)) {
    panelOpen.value = false
  }
}

watch(keyword, (val) => {
  clearTimeout(debounceTimer)
  if (!val.trim()) {
    items.value = []
    total.value = 0
    panelOpen.value = false
    return
  }
  debounceTimer = setTimeout(() => {
    void runSearch(true)
  }, 320)
})

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
})

onUnmounted(() => {
  document.removeEventListener('click', onDocumentClick)
  clearTimeout(debounceTimer)
})
</script>

<template>
  <div ref="rootRef" class="super-search">
    <el-input
      v-model="keyword"
      placeholder="超级搜索：SKU 规格值 / 编码 / 货号 / 资料编码"
      :prefix-icon="Search"
      class="super-search-input"
      clearable
      @focus="onFocus"
      @keyup.enter="onEnter"
      @clear="onClear"
    />

    <div v-show="panelOpen && keyword.trim()" class="super-search-panel">
      <div v-if="loading" class="panel-status">搜索中…</div>
      <div v-else-if="!items.length" class="panel-status">未找到匹配的 SKU</div>
      <template v-else>
        <div class="panel-summary">共 {{ total }} 条 SKU 匹配</div>
        <div class="result-list">
          <div v-for="item in items" :key="item.skuId" class="result-item">
            <el-image
              v-if="resolvePic(item)"
              :src="resolvePic(item)"
              class="result-pic"
              fit="cover"
            />
            <div v-else class="result-pic result-pic-empty">无图</div>

            <div class="result-main">
              <div class="result-spec" :title="item.specLabel">{{ item.specLabel }}</div>
              <div class="result-meta">
                <span class="result-code">{{ item.skuCode }}</span>
                <span class="result-divider">·</span>
                <span class="result-product" :title="item.productName">{{ item.productName }}</span>
              </div>
              <div class="result-sub">
                <span v-if="item.materialCode">资料 {{ item.materialCode }}</span>
                <span v-if="item.productSn">货号 {{ item.productSn }}</span>
                <span v-if="item.brandName">{{ item.brandName }}</span>
              </div>
              <div class="result-shops">
                <template v-if="item.listedShops?.length">
                  <span
                    v-for="shop in item.listedShops.slice(0, 4)"
                    :key="shop.shopId"
                    class="shop-chip"
                    :title="shop.shopName"
                  >
                    <el-image
                      v-if="shop.platformTypeLogo"
                      :src="shop.platformTypeLogo"
                      class="shop-logo"
                      fit="cover"
                    />
                    {{ shop.shopName }}
                  </span>
                  <span v-if="item.listedShopCount > 4" class="shop-more">
                    +{{ item.listedShopCount - 4 }}
                  </span>
                </template>
                <span v-else class="shop-empty">未铺货</span>
              </div>
            </div>

            <div class="result-side">
              <div class="result-price">¥{{ item.price }}</div>
              <div class="result-stock" :class="{ 'result-stock-zero': item.stock === 0 }">
                库存 {{ item.stock }}
              </div>
              <div class="result-actions">
                <el-button type="primary" link @click="openDetail(item)">查看</el-button>
                <el-button type="primary" link @click="openSkuManage(item)">SKU</el-button>
                <el-button type="primary" link @click="openListing(item)">铺货</el-button>
              </div>
            </div>
          </div>
        </div>

        <div v-if="total > pageSize" class="panel-pagination">
          <el-pagination
            small
            background
            layout="prev, pager, next"
            :total="total"
            :current-page="page"
            :page-size="pageSize"
            @current-change="onPageChange"
          />
        </div>
      </template>
    </div>

    <ProductDetailDrawer v-model="detailVisible" :product-id="detailProductId" />
    <ProductSkuManageDialog
      v-model="skuManageVisible"
      :product-id="skuManageProductId"
      :product-name="skuManageProductName"
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
.super-search {
  position: relative;
  width: 360px;
}

.super-search-input {
  width: 100%;
}

.super-search-panel {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  z-index: 2000;
  width: min(640px, calc(100vw - 48px));
  max-height: min(70vh, 560px);
  overflow: auto;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  padding: 12px 0;
}

.panel-status,
.panel-summary {
  padding: 12px 16px;
  color: #909399;
  font-size: 13px;
}

.panel-summary {
  color: #606266;
  border-bottom: 1px solid #f0f2f5;
  padding-bottom: 10px;
  margin-bottom: 4px;
}

.result-list {
  display: flex;
  flex-direction: column;
}

.result-item {
  display: flex;
  gap: 12px;
  padding: 10px 16px;
  border-bottom: 1px solid #f5f7fa;
}

.result-item:last-child {
  border-bottom: none;
}

.result-item:hover {
  background: #fafafa;
}

.result-pic {
  width: 52px;
  height: 52px;
  border-radius: 4px;
  flex-shrink: 0;
}

.result-pic-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  color: #c0c4cc;
  font-size: 12px;
}

.result-main {
  flex: 1;
  min-width: 0;
}

.result-spec {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.result-meta {
  margin-top: 4px;
  font-size: 12px;
  color: #606266;
  display: flex;
  align-items: center;
  min-width: 0;
}

.result-code {
  flex-shrink: 0;
  font-family: ui-monospace, monospace;
}

.result-product {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-divider {
  margin: 0 6px;
  color: #dcdfe6;
}

.result-sub {
  margin-top: 4px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 12px;
  color: #909399;
}

.result-shops {
  margin-top: 6px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.shop-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: 120px;
  padding: 0 6px;
  height: 22px;
  border-radius: 4px;
  background: #f5f7fa;
  font-size: 11px;
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.shop-logo {
  width: 14px;
  height: 14px;
  border-radius: 2px;
  flex-shrink: 0;
}

.shop-more,
.shop-empty {
  font-size: 11px;
  color: #909399;
}

.result-side {
  flex-shrink: 0;
  text-align: right;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.result-price {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.result-stock {
  font-size: 12px;
  color: #606266;
}

.result-stock-zero {
  color: var(--el-color-danger);
  font-weight: 500;
}

.result-actions {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.panel-pagination {
  display: flex;
  justify-content: center;
  padding: 10px 16px 4px;
  border-top: 1px solid #f0f2f5;
}
</style>
