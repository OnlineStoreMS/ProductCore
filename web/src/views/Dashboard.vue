<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Box, Collection, Grid, TrendCharts } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { fetchDashboardStats } from '../api/product'
import type { Product, ProductGroup } from '../types/product'

const loading = ref(false)
const recentProducts = ref<Product[]>([])
const groups = ref<ProductGroup[]>([])

const stats = ref([
  { title: '商品总数', value: 0, icon: Box, color: '#409eff' },
  { title: '已上架', value: 0, icon: TrendCharts, color: '#67c23a' },
  { title: 'SKU 总数', value: 0, icon: Grid, color: '#e6a23c' },
  { title: '商品分组', value: 0, icon: Collection, color: '#909399' },
])

onMounted(async () => {
  loading.value = true
  try {
    const data = await fetchDashboardStats()
    stats.value[0].value = data.productTotal
    stats.value[1].value = data.publishedTotal
    stats.value[2].value = data.skuTotal
    stats.value[3].value = data.groupTotal
    recentProducts.value = data.recentProducts
    groups.value = data.groups
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-loading="loading" class="dashboard">
    <el-row :gutter="20" class="stat-row">
      <el-col v-for="item in stats" :key="item.title" :xs="24" :sm="12" :lg="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-inner">
            <div class="stat-info">
              <div class="stat-title">{{ item.title }}</div>
              <div class="stat-value">{{ item.value }}</div>
            </div>
            <div class="stat-icon" :style="{ background: item.color + '20', color: item.color }">
              <el-icon :size="28"><component :is="item.icon" /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :xs="24" :lg="14">
        <el-card>
          <template #header>
            <span>最近商品</span>
            <router-link to="/products"><el-button type="primary" link>查看全部</el-button></router-link>
          </template>
          <el-table :data="recentProducts" stripe>
            <el-table-column label="商品" min-width="200">
              <template #default="{ row }">
                <div class="product-cell">
                  <el-image :src="row.pic" class="thumb" fit="cover" />
                  <div>
                    <div class="name">{{ row.name }}</div>
                    <div class="sn">{{ row.productSn }}</div>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="brandName" label="品牌" width="90" />
            <el-table-column label="价格" width="100">
              <template #default="{ row }">¥{{ row.price }}</template>
            </el-table-column>
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.publishStatus ? 'success' : 'info'" size="small">
                  {{ row.publishStatus ? '上架' : '下架' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="10">
        <el-card>
          <template #header>商品分组概览</template>
          <el-table :data="groups" size="small">
            <el-table-column prop="name" label="分组名称" />
            <el-table-column prop="productCount" label="商品数" width="80" align="center" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.stat-card :deep(.el-card__body) {
  padding: 20px;
}

.stat-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-title {
  color: #909399;
  font-size: 14px;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.product-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.thumb {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  flex-shrink: 0;
}

.name {
  font-weight: 500;
  font-size: 14px;
}

.sn {
  color: #909399;
  font-size: 12px;
}

:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
