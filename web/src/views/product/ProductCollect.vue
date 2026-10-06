<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createProductCollect, fetchProductCollects, type ProductCollectTask } from '../../api/collect'

const productUrl = ref('')
const submitting = ref(false)
const loading = ref(false)
const tableData = ref<ProductCollectTask[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
let timer = 0

const statusText: Record<string, string> = {
  pending: '排队中',
  claimed: '已领取',
  running: '采集中',
  succeeded: '已完成',
  failed: '失败',
  cancelled: '已取消',
}

function statusLabel(status: string) {
  return statusText[status] || status || '未知'
}

function statusType(status: string): 'success' | 'danger' | 'warning' | 'info' {
  if (status === 'succeeded') return 'success'
  if (status === 'failed' || status === 'cancelled') return 'danger'
  if (status === 'running' || status === 'claimed') return 'warning'
  return 'info'
}

function errorText(e: unknown, fallback: string) {
  const err = e as { response?: { data?: { message?: string } }; message?: string }
  return err.response?.data?.message || err.message || fallback
}

function formatTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function hasActive(rows: ProductCollectTask[]) {
  return rows.some((row) => row.status === 'pending' || row.status === 'claimed' || row.status === 'running')
}

async function loadData(silent = false) {
  if (!silent) loading.value = true
  try {
    const data = await fetchProductCollects(page.value, pageSize.value)
    tableData.value = data.list
    total.value = data.total
  } catch (e) {
    if (!silent) ElMessage.error(errorText(e, '加载失败'))
  } finally {
    loading.value = false
  }
}

async function handleCollect() {
  const url = productUrl.value.trim()
  if (!url) {
    ElMessage.warning('请输入商品链接')
    return
  }
  submitting.value = true
  try {
    await createProductCollect(url)
    ElMessage.success('已发送到采集电脑')
    productUrl.value = ''
    page.value = 1
    await loadData()
  } catch (e) {
    ElMessage.error(errorText(e, '发送失败'))
  } finally {
    submitting.value = false
  }
}

function onPageChange(next: number) {
  page.value = next
  loadData()
}

onMounted(async () => {
  await loadData()
  timer = window.setInterval(() => {
    if (hasActive(tableData.value)) loadData(true)
  }, 4000)
})

onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="collect-page">
    <el-card>
      <template #header>
        <span>商品采集</span>
      </template>
      <p class="hint">
        粘贴淘宝或天猫商品链接。点击采集后，任务会发给当前在线的 WindowsAgent，由固定的 Chrome 测试版打开该链接。
      </p>
      <div class="submit-row">
        <el-input
          v-model="productUrl"
          placeholder="https://item.taobao.com/item.htm?id=..."
          clearable
          @keyup.enter="handleCollect"
        />
        <el-button type="primary" :loading="submitting" @click="handleCollect">采集</el-button>
      </div>
    </el-card>

    <el-card v-loading="loading" class="list-card">
      <template #header>
        <span>采集记录</span>
        <el-button link type="primary" @click="loadData()">刷新</el-button>
      </template>
      <el-table :data="tableData" stripe border>
        <el-table-column label="时间" width="180">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column prop="platformName" label="平台" width="110" />
        <el-table-column label="商品链接" min-width="280" show-overflow-tooltip>
          <template #default="{ row }">
            <a :href="row.productUrl" target="_blank" rel="noreferrer">{{ row.productUrl }}</a>
          </template>
        </el-table-column>
        <el-table-column prop="agentName" label="采集电脑" width="140" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="说明" min-width="220" show-overflow-tooltip />
      </el-table>
      <div class="pager">
        <el-pagination
          background
          layout="total, prev, pager, next"
          :total="total"
          :page-size="pageSize"
          :current-page="page"
          @current-change="onPageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.hint {
  margin: 0 0 16px;
  color: #606266;
  line-height: 1.6;
}

.submit-row {
  display: flex;
  gap: 12px;
}

.list-card {
  margin-top: 16px;
}

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
