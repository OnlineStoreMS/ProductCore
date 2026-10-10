<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Upload } from '@element-plus/icons-vue'
import {
  fetchDistributionItems,
  fetchDistributionShop,
  importDistributionItems,
  type DistributionItem,
  type DistributionShop,
} from '../../api/distribution'

const route = useRoute()
const router = useRouter()
const shopId = computed(() => Number(route.params.id) || 0)

const loading = ref(false)
const importing = ref(false)
const shop = ref<DistributionShop | null>(null)
const tableData = ref<DistributionItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const fileInput = ref<HTMLInputElement | null>(null)
const detailVisible = ref(false)
const detail = ref<DistributionItem | null>(null)

function errorText(e: unknown, fallback: string) {
  const err = e as { response?: { data?: { message?: string } }; message?: string }
  return err.response?.data?.message || err.message || fallback
}

function dash(value?: string) {
  const text = (value || '').trim()
  return text || '-'
}

function itemLink(row: DistributionItem) {
  if (row.itemUrl) return row.itemUrl
  if (row.itemId) return `https://item.taobao.com/item.htm?id=${row.itemId}`
  return ''
}

async function loadShop() {
  shop.value = await fetchDistributionShop(shopId.value)
}

async function loadData() {
  loading.value = true
  try {
    const data = await fetchDistributionItems(shopId.value, {
      keyword: keyword.value || undefined,
      page: page.value,
      pageSize: pageSize.value,
    })
    tableData.value = data.list
    total.value = data.total
  } catch (e) {
    ElMessage.error(errorText(e, '加载失败'))
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    await loadShop()
    await loadData()
  } catch (e) {
    ElMessage.error(errorText(e, '店铺不存在'))
  }
})

function handleSearch() {
  page.value = 1
  loadData()
}

function pickFile() {
  fileInput.value?.click()
}

async function onFile(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  importing.value = true
  try {
    const result = await importDistributionItems(shopId.value, file)
    const imageNote = result.imageFailed ? `，图片失败 ${result.imageFailed}` : ''
    ElMessage.success(`导入完成：新增 ${result.created}，更新 ${result.updated}${imageNote}`)
    page.value = 1
    await loadShop()
    await loadData()
  } catch (e) {
    ElMessage.error(errorText(e, '导入失败'))
  } finally {
    importing.value = false
  }
}

function openDetail(row: DistributionItem) {
  detail.value = row
  detailVisible.value = true
}

function openSame(row: DistributionItem) {
  const url = itemLink(row)
  if (!url) {
    ElMessage.warning('这条商品没有链接')
    return
  }
  window.open(url, '_blank', 'noopener')
}
</script>

<template>
  <div class="page">
    <el-card v-loading="loading">
      <template #header>
        <div class="title-wrap">
          <el-button link type="primary" @click="router.push('/distribution-shops')">返回</el-button>
          <span>{{ shop?.name || '店铺商品' }}</span>
          <el-tag v-if="shop?.platformTypeName" size="small">{{ shop.platformTypeName }}</el-tag>
          <el-tag size="small" type="info">{{ total }} 件</el-tag>
        </div>
        <div class="header-actions">
          <el-input
            v-model="keyword"
            clearable
            placeholder="标题 / 商品ID"
            style="width: 220px"
            @keyup.enter="handleSearch"
            @clear="handleSearch"
          />
          <el-button @click="handleSearch">查询</el-button>
          <el-button type="primary" :icon="Upload" :loading="importing" @click="pickFile">导入店铺商品</el-button>
          <input ref="fileInput" type="file" accept=".xls,.xlsx,.html,.htm" hidden @change="onFile" />
        </div>
      </template>

      <el-table :data="tableData" stripe border>
        <el-table-column type="selection" width="48" />
        <el-table-column label="商品信息" min-width="360">
          <template #default="{ row }">
            <div class="goods">
              <el-image :src="row.picUrl" fit="cover" class="goods-pic">
                <template #error><div class="goods-pic goods-pic--empty">无图</div></template>
              </el-image>
              <div class="goods-text">
                <a
                  v-if="itemLink(row)"
                  class="goods-title goods-link"
                  :href="itemLink(row)"
                  target="_blank"
                  rel="noreferrer"
                  :title="row.title"
                >{{ row.title || '查看商品' }}</a>
                <div v-else class="goods-title">{{ row.title || '-' }}</div>
                <div class="goods-id">商品ID: {{ row.itemId }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="月成交笔数" width="120" align="center">
          <template #default="{ row }">{{ dash(row.monthDeals) }}</template>
        </el-table-column>
        <el-table-column label="月代销" width="100" align="center">
          <template #default="{ row }">{{ dash(row.monthConsign) }}</template>
        </el-table-column>
        <el-table-column label="销量" width="100" align="center">
          <template #default="{ row }">{{ dash(row.sales) }}</template>
        </el-table-column>
        <el-table-column label="价格" width="100" align="center">
          <template #default="{ row }">{{ dash(row.price) }}</template>
        </el-table-column>
        <el-table-column label="发货时间" width="120" align="center">
          <template #default="{ row }">{{ dash(row.shipTime) }}</template>
        </el-table-column>
        <el-table-column label="上架时间" width="160" align="center">
          <template #default="{ row }">{{ dash(row.listedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <div class="ops">
              <el-button link type="primary" @click="openDetail(row)">详情</el-button>
              <el-button link type="primary" @click="openSame(row)">同款铺货</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="pager">
        <el-pagination
          background
          layout="total, prev, pager, next"
          :total="total"
          :page-size="pageSize"
          :current-page="page"
          @current-change="(p: number) => { page = p; loadData() }"
        />
      </div>
    </el-card>

    <el-dialog v-model="detailVisible" :title="detail?.title || '商品详情'" width="640px" @closed="detail = null">
      <template v-if="detail">
        <div class="detail-head">
          <el-image :src="detail.picUrl" fit="cover" class="detail-pic" />
          <div>
            <a
              v-if="itemLink(detail)"
              class="goods-title goods-link"
              :href="itemLink(detail)"
              target="_blank"
              rel="noreferrer"
            >{{ detail.title || '查看商品' }}</a>
            <div v-else class="goods-title">{{ detail.title }}</div>
            <div class="goods-id">商品ID: {{ detail.itemId }}</div>
          </div>
        </div>
        <el-descriptions :column="2" border class="detail-desc">
          <el-descriptions-item label="价格">{{ dash(detail.price) }}</el-descriptions-item>
          <el-descriptions-item label="销量">{{ dash(detail.sales) }}</el-descriptions-item>
          <el-descriptions-item label="月成交笔数">{{ dash(detail.monthDeals) }}</el-descriptions-item>
          <el-descriptions-item label="月代销">{{ dash(detail.monthConsign) }}</el-descriptions-item>
          <el-descriptions-item label="评论数">{{ dash(detail.commentCount) }}</el-descriptions-item>
          <el-descriptions-item label="发货时间">{{ dash(detail.shipTime) }}</el-descriptions-item>
          <el-descriptions-item label="上架时间">{{ dash(detail.listedAt) }}</el-descriptions-item>
          <el-descriptions-item label="类目">{{ dash(detail.category) }}</el-descriptions-item>
          <el-descriptions-item label="标签">{{ dash(detail.tags) }}</el-descriptions-item>
          <el-descriptions-item label="来源店铺">{{ dash(detail.sourceShopName) }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.title-wrap,
.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.goods {
  display: flex;
  gap: 10px;
  align-items: center;
}

.goods-pic {
  width: 56px;
  height: 56px;
  border-radius: 4px;
  flex: none;
  background: #f3f4f6;
}

.goods-pic--empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
  font-size: 12px;
}

.goods-text {
  min-width: 0;
}

.goods-title {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
}

.goods-link {
  color: #2563eb;
  text-decoration: none;
}

.goods-link:hover {
  text-decoration: underline;
}

.goods-id {
  margin-top: 4px;
  color: #6b7280;
  font-size: 12px;
}

.ops {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.detail-head {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.detail-pic {
  width: 88px;
  height: 88px;
  border-radius: 4px;
  flex: none;
}

.detail-desc {
  margin-top: 8px;
}

:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}
</style>
