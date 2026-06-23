<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Box, Collection, FolderOpened, Grid, HomeFilled, Menu as MenuIcon, Setting,
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const collapsed = defineModel<boolean>('collapsed', { default: false })

const activeMenu = computed(() => route.path)

const menuItems = [
  { path: '/dashboard', title: '工作台', icon: HomeFilled },
  {
    title: '商品管理',
    icon: Box,
    children: [
      { path: '/products', title: '商品列表', icon: Grid },
      { path: '/products/create', title: '添加商品', icon: Collection },
      { path: '/groups', title: '商品分组', icon: FolderOpened },
    ],
  },
  {
    title: '基础数据',
    icon: Setting,
    children: [
      { path: '/categories', title: '商品分类', icon: MenuIcon },
      { path: '/brands', title: '品牌管理', icon: Collection },
    ],
  },
]

function navigate(path: string) {
  router.push(path)
}
</script>

<template>
  <aside class="sidebar" :class="{ collapsed }">
    <div class="logo">
      <div class="logo-icon">P</div>
      <transition name="fade">
        <span v-if="!collapsed" class="logo-text">ProductCore</span>
      </transition>
    </div>
    <el-scrollbar class="menu-scroll">
      <el-menu
        :default-active="activeMenu"
        :collapse="collapsed"
        background-color="transparent"
        text-color="rgba(255,255,255,0.75)"
        active-text-color="#fff"
        :collapse-transition="false"
      >
        <template v-for="item in menuItems" :key="item.title">
          <el-sub-menu v-if="item.children" :index="item.title">
            <template #title>
              <el-icon><component :is="item.icon" /></el-icon>
              <span>{{ item.title }}</span>
            </template>
            <el-menu-item
              v-for="child in item.children"
              :key="child.path"
              :index="child.path"
              @click="navigate(child.path)"
            >
              <el-icon><component :is="child.icon" /></el-icon>
              <span>{{ child.title }}</span>
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item v-else :index="item.path" @click="navigate(item.path!)">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.title }}</span>
          </el-menu-item>
        </template>
      </el-menu>
    </el-scrollbar>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 220px;
  background: linear-gradient(180deg, #1a2332 0%, #0f1419 100%);
  display: flex;
  flex-direction: column;
  transition: width 0.25s;
  flex-shrink: 0;
}

.sidebar.collapsed {
  width: 64px;
}

.logo {
  height: 56px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  gap: 10px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.logo-icon {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, #409eff, #67c23a);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  color: #fff;
  font-size: 16px;
  flex-shrink: 0;
}

.logo-text {
  font-size: 16px;
  font-weight: 600;
  color: #fff;
  white-space: nowrap;
}

.menu-scroll {
  flex: 1;
  padding: 8px 0;
}

.sidebar :deep(.el-menu) {
  border-right: none;
}

.sidebar :deep(.el-menu-item.is-active) {
  background: rgba(64, 158, 255, 0.2) !important;
  border-radius: 6px;
  margin: 2px 8px;
  width: calc(100% - 16px);
}

.sidebar :deep(.el-sub-menu .el-menu-item) {
  padding-left: 48px !important;
}

.sidebar :deep(.el-sub-menu__title:hover),
.sidebar :deep(.el-menu-item:hover) {
  background: rgba(255, 255, 255, 0.06) !important;
}
</style>
