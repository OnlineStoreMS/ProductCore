<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Expand, Fold, Search, Bell } from '@element-plus/icons-vue'
import Sidebar from './Sidebar.vue'

const route = useRoute()
const collapsed = ref(false)

const breadcrumbs = computed(() => {
  const title = (route.meta.title as string) || 'ProductCore'
  if (route.path.startsWith('/products') && route.params.id) {
    return ['商品管理', '编辑商品']
  }
  if (route.path === '/products/create') {
    return ['商品管理', '添加商品']
  }
  if (route.path.startsWith('/products')) {
    return ['商品管理', '商品列表']
  }
  if (route.path.startsWith('/categories')) return ['基础数据', '商品分类']
  if (route.path.startsWith('/brands')) return ['基础数据', '品牌管理']
  if (route.path.startsWith('/groups')) return ['商品管理', '商品分组']
  return ['首页', title]
})
</script>

<template>
  <div class="admin-layout">
    <Sidebar v-model:collapsed="collapsed" />
    <div class="main-area">
      <header class="header">
        <div class="header-left">
          <el-button :icon="collapsed ? Expand : Fold" text @click="collapsed = !collapsed" />
          <el-breadcrumb separator="/">
            <el-breadcrumb-item v-for="(item, i) in breadcrumbs" :key="i">{{ item }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-input placeholder="搜索商品 / SKU / 货号" :prefix-icon="Search" class="search-input" clearable />
          <el-badge :value="3" class="notice-badge">
            <el-button :icon="Bell" circle />
          </el-badge>
          <el-avatar :size="32" src="https://api.dicebear.com/7.x/avataaars/svg?seed=admin" />
        </div>
      </header>
      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<style scoped>
.admin-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  background: #f0f2f5;
}

.header {
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.search-input {
  width: 260px;
}

.content {
  flex: 1;
  overflow: auto;
  padding: 20px;
}
</style>
