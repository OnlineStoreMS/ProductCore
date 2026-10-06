<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchProductCollects, type ProductCollectTask } from '../../api/collect'

const loading = ref(false)
const tableData = ref<ProductCollectTask[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

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

async function loadData() {
  loading.value = true
  try {
    const data = await fetchProductCollects(page.value, pageSize.value)
    tableData.value = data.list
    total.value = data.total
  } catch (e) {
    ElMessage.error(errorText(e, '加载失败'))
  } finally {
    loading.value = false
  }
}

function onPageChange(next: number) {
  page.value = next
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="collect-page">
    <el-card>
      <template #header>
        <span>商品采集</span>
      </template>
      <p class="hint">
        用 Chrome 扩展在商品页人工点至尊宝「悬浮标题采集」「手机端主图视频SKU」「SKU工具」，再上传。不再通过 WindowsAgent 下发链接。
        扩展安装见仓库 <code>extensions/collect</code>。
      </p>
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
        <el-table-column prop="agentName" label="来源" width="140" show-overflow-tooltip />
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
  margin: 0;
  color: #606266;
  line-height: 1.6;
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
