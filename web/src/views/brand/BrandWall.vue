<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { fetchBrands } from '../../api/product'
import type { Brand } from '../../types/product'

const router = useRouter()
const loading = ref(false)
const brands = ref<Brand[]>([])
const keyword = ref('')

const visibleBrands = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  return brands.value
    .filter((b) => (b.productCount || 0) > 0)
    .filter((b) => !q || b.name.toLowerCase().includes(q) || b.firstLetter?.toLowerCase() === q)
    .slice()
    .sort((a, b) => {
      const byCount = (b.productCount || 0) - (a.productCount || 0)
      if (byCount !== 0) return byCount
      return (b.sort || 0) - (a.sort || 0)
    })
})

async function loadData() {
  loading.value = true
  try {
    brands.value = await fetchBrands()
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function openBrand(brand: Brand) {
  router.push({ path: '/products', query: { brandId: String(brand.id) } })
}
</script>

<template>
  <div v-loading="loading" class="brand-wall-page">
    <el-card shadow="never" class="wall-card">
      <template #header>
        <div class="wall-header">
          <div>
            <div class="wall-title">品牌墙</div>
            <div class="wall-sub">展示当前有商品的品牌，点击进入该品牌商品列表</div>
          </div>
          <el-input
            v-model="keyword"
            placeholder="搜索品牌"
            :prefix-icon="Search"
            clearable
            class="wall-search"
          />
        </div>
      </template>

      <div v-if="visibleBrands.length" class="wall-grid">
        <button
          v-for="brand in visibleBrands"
          :key="brand.id"
          type="button"
          class="brand-tile"
          @click="openBrand(brand)"
        >
          <div class="logo-wrap">
            <img v-if="brand.logo" :src="brand.logo" :alt="brand.name" class="logo-img" />
            <span v-else class="logo-fallback">{{ brand.firstLetter || brand.name.slice(0, 1) || '?' }}</span>
          </div>
          <div class="brand-name" :title="brand.name">{{ brand.name }}</div>
          <div class="brand-count">{{ brand.productCount }} 件商品</div>
        </button>
      </div>
      <el-empty v-else description="暂无带商品的品牌" :image-size="88" />
    </el-card>
  </div>
</template>

<style scoped>
.brand-wall-page {
  min-height: 100%;
}

.wall-card {
  border: none;
}

.wall-card :deep(.el-card__header) {
  padding: 18px 20px 12px;
  border-bottom: 1px solid #f0f0f0;
}

.wall-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.wall-title {
  font-size: 18px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.88);
  line-height: 1.3;
}

.wall-sub {
  margin-top: 4px;
  font-size: 13px;
  color: rgba(0, 0, 0, 0.45);
}

.wall-search {
  width: 240px;
}

.wall-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(176px, 1fr));
  gap: 16px;
  padding: 4px 0 8px;
}

.brand-tile {
  appearance: none;
  border: 1px solid #ebebeb;
  background: #fff;
  border-radius: 10px;
  padding: 16px 12px 14px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  transition: border-color 0.15s ease, box-shadow 0.15s ease, transform 0.15s ease;
  text-align: center;
}

.brand-tile:hover {
  border-color: #ffb366;
  box-shadow: 0 8px 20px rgba(255, 119, 0, 0.08);
  transform: translateY(-2px);
}

.logo-wrap {
  width: 140px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-img {
  max-width: 140px;
  max-height: 56px;
  width: auto;
  height: auto;
  object-fit: contain;
  background: transparent;
}

.logo-fallback {
  font-size: 28px;
  font-weight: 700;
  color: #ff7700;
  line-height: 1;
}

.brand-name {
  width: 100%;
  font-size: 14px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.brand-count {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
}
</style>
