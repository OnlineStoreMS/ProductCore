<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { fetchBrands, reorderBrands } from '../../api/product'
import type { Brand } from '../../types/product'

const router = useRouter()
const loading = ref(false)
const brands = ref<Brand[]>([])
const keyword = ref('')
const dragFrom = ref<number | null>(null)
const dragOver = ref<number | null>(null)
const justDragged = ref(false)
const savingOrder = ref(false)

const searching = computed(() => !!keyword.value.trim())

const visibleBrands = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  if (!q) return brands.value
  return brands.value.filter(
    (b) => b.name.toLowerCase().includes(q) || b.firstLetter?.toLowerCase() === q,
  )
})

function wallOrder(list: Brand[]) {
  return list
    .filter((b) => (b.productCount || 0) > 0)
    .slice()
    .sort((a, b) => {
      const bySort = (b.sort || 0) - (a.sort || 0)
      if (bySort !== 0) return bySort
      return a.id - b.id
    })
}

async function loadData() {
  loading.value = true
  try {
    brands.value = wallOrder(await fetchBrands())
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function logoWrapStyle(brand: Brand) {
  if (!brand.logoBg) return undefined
  return { background: brand.logoBg }
}

function openBrand(brand: Brand) {
  if (justDragged.value) {
    justDragged.value = false
    return
  }
  router.push({ path: '/products', query: { brandId: String(brand.id) } })
}

function resetDrag() {
  dragFrom.value = null
  dragOver.value = null
}

function onDragStart(index: number, event: DragEvent) {
  if (searching.value || savingOrder.value) {
    event.preventDefault()
    return
  }
  dragFrom.value = index
  justDragged.value = false
  event.dataTransfer?.setData('text/plain', String(index))
  if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move'
}

function onDragOver(index: number, event: DragEvent) {
  if (dragFrom.value === null || searching.value) return
  event.preventDefault()
  if (event.dataTransfer) event.dataTransfer.dropEffect = 'move'
  dragOver.value = index
}

async function onDrop(index: number) {
  const from = dragFrom.value
  resetDrag()
  if (from === null || from === index || searching.value) return
  const next = brands.value.slice()
  const [moved] = next.splice(from, 1)
  if (!moved) return
  next.splice(index, 0, moved)
  brands.value = next
  justDragged.value = true
  await persistOrder()
}

async function persistOrder() {
  savingOrder.value = true
  try {
    await reorderBrands(brands.value.map((b) => b.id))
    brands.value = brands.value.map((b, i) => ({ ...b, sort: brands.value.length - i }))
  } catch (e) {
    ElMessage.error((e as Error).message || '保存排序失败')
    await loadData()
  } finally {
    savingOrder.value = false
  }
}
</script>

<template>
  <div v-loading="loading" class="brand-wall-page">
    <el-card shadow="never" class="wall-card">
      <template #header>
        <div class="wall-header">
          <div>
            <div class="wall-title">品牌墙</div>
            <div class="wall-sub">展示当前有商品的品牌，拖动卡片可调整顺序，点击进入该品牌商品列表</div>
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
          v-for="(brand, index) in visibleBrands"
          :key="brand.id"
          type="button"
          class="brand-tile"
          :class="{
            'is-dragging': dragFrom === index,
            'is-over': dragOver === index && dragFrom !== index,
            'is-locked': searching,
          }"
          :draggable="!searching"
          @dragstart="onDragStart(index, $event)"
          @dragover="onDragOver(index, $event)"
          @drop="onDrop(index)"
          @dragend="resetDrag"
          @click="openBrand(brand)"
        >
          <div class="logo-wrap" :class="{ 'has-bg': !!brand.logoBg }" :style="logoWrapStyle(brand)">
            <img
              v-if="brand.logo"
              :src="brand.logo"
              :alt="brand.name"
              class="logo-img"
              draggable="false"
            />
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
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
  padding: 4px 0 8px;
}

.brand-tile {
  appearance: none;
  border: 1px solid #ebebeb;
  background: #fff;
  border-radius: 10px;
  padding: 22px 16px 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  cursor: grab;
  transition: border-color 0.15s ease, box-shadow 0.15s ease, transform 0.15s ease;
  text-align: center;
}

.brand-tile:hover {
  border-color: #ffb366;
  box-shadow: 0 8px 20px rgba(255, 119, 0, 0.08);
  transform: translateY(-2px);
}

.brand-tile.is-dragging {
  opacity: 0.45;
  cursor: grabbing;
}

.brand-tile.is-over {
  border-color: #ff7700;
  background: #fff7e6;
}

.brand-tile.is-locked {
  cursor: pointer;
}

.logo-wrap {
  width: 100%;
  height: 96px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-wrap.has-bg {
  border-radius: 8px;
  padding: 8px;
  box-sizing: border-box;
}

.logo-img {
  width: 100%;
  height: 96px;
  object-fit: contain;
  background: transparent;
  pointer-events: none;
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
