<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowDown } from '@element-plus/icons-vue'
import {
  Box, Collection, CollectionTag, Delete, Document, FolderOpened, Grid, HomeFilled, Menu as MenuIcon, Setting, Shop,
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const collapsed = defineModel<boolean>('collapsed', { default: false })

type MenuLeaf = { path: string; title: string; icon: object }
type MenuGroup = { title: string; icon: object; children: MenuLeaf[] }
type MenuEntry = MenuLeaf | MenuGroup

function isGroup(item: MenuEntry): item is MenuGroup {
  return 'children' in item
}

const activeMenu = computed(() => route.path)

const menuItems: MenuEntry[] = [
  { path: '/dashboard', title: '工作台', icon: HomeFilled },
  {
    title: '商品管理',
    icon: Box,
    children: [
      { path: '/products', title: '商品列表', icon: Grid },
      { path: '/products/drafts', title: '商品草稿箱', icon: Document },
      { path: '/groups', title: '商品分组', icon: FolderOpened },
      { path: '/keywords', title: '关键词管理', icon: CollectionTag },
      { path: '/products/trash', title: '商品回收站', icon: Delete },
    ],
  },
  {
    title: '渠道管理',
    icon: Shop,
    children: [
      { path: '/platform-shops', title: '店铺管理', icon: Shop },
      { path: '/platform-types', title: '店铺类型', icon: Grid },
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

function openKeysForPath(path: string): string[] {
  if (path.startsWith('/products') || path.startsWith('/groups') || path.startsWith('/keywords')) return ['商品管理']
  if (path.startsWith('/platform')) return ['渠道管理']
  if (path.startsWith('/categories') || path.startsWith('/brands')) return ['基础数据']
  return []
}

const openKeys = ref<string[]>(openKeysForPath(route.path))

watch(
  () => route.path,
  (path) => {
    for (const key of openKeysForPath(path)) {
      if (!openKeys.value.includes(key)) openKeys.value.push(key)
    }
  },
)

function isOpen(title: string) {
  return openKeys.value.includes(title)
}

function toggleOpen(title: string) {
  const i = openKeys.value.indexOf(title)
  if (i >= 0) openKeys.value.splice(i, 1)
  else openKeys.value.push(title)
}

function isGroupActive(item: MenuGroup) {
  return item.children.some((c) => c.path === activeMenu.value)
}

function navigate(path: string) {
  router.push(path)
}
</script>

<template>
  <aside class="sidebar" :class="{ collapsed }">
    <div class="logo">
      <div class="logo-icon">P</div>
      <span v-if="!collapsed" class="logo-text">ProductCore</span>
    </div>

    <div class="menu-scroll">
      <nav class="nav-menu">
        <template v-for="item in menuItems" :key="isGroup(item) ? item.title : item.path">
          <!-- 一级菜单 -->
          <div
            v-if="!isGroup(item)"
            class="nav-item"
            :class="{ 'is-active': activeMenu === item.path }"
            :title="collapsed ? item.title : undefined"
            @click="navigate(item.path)"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span v-if="!collapsed" class="nav-label">{{ item.title }}</span>
          </div>

          <!-- 分组菜单：v-show 即时展开/收起，无高度动画 -->
          <el-popover
            v-else-if="collapsed"
            placement="right-start"
            trigger="click"
            :width="168"
            popper-class="sidebar-submenu-popper"
          >
            <template #reference>
              <div
                class="nav-item nav-group-title"
                :class="{ 'is-active': isGroupActive(item) }"
                :title="item.title"
              >
                <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
              </div>
            </template>
            <div class="popover-menu">
              <div
                v-for="child in item.children"
                :key="child.path"
                class="popover-item"
                :class="{ 'is-active': activeMenu === child.path }"
                @click="navigate(child.path)"
              >
                {{ child.title }}
              </div>
            </div>
          </el-popover>

          <div v-else class="nav-group">
            <div
              class="nav-item nav-group-title"
              :class="{ 'is-active': isGroupActive(item) }"
              @click="toggleOpen(item.title)"
            >
              <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
              <span v-if="!collapsed" class="nav-label">{{ item.title }}</span>
              <el-icon v-if="!collapsed" class="nav-arrow" :class="{ open: isOpen(item.title) }">
                <ArrowDown />
              </el-icon>
            </div>
            <div v-show="!collapsed && isOpen(item.title)" class="nav-children">
              <div
                v-for="child in item.children"
                :key="child.path"
                class="nav-item nav-child"
                :class="{ 'is-active': activeMenu === child.path }"
                @click="navigate(child.path)"
              >
                <el-icon class="nav-icon"><component :is="child.icon" /></el-icon>
                <span class="nav-label">{{ child.title }}</span>
              </div>
            </div>
          </div>
        </template>
      </nav>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 220px;
  background: linear-gradient(180deg, #1a2332 0%, #0f1419 100%);
  display: flex;
  flex-direction: column;
  transition: width 0.2s ease;
  flex-shrink: 0;
  contain: layout style;
  isolation: isolate;
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
  flex-shrink: 0;
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
  overflow-y: auto;
  overflow-x: hidden;
  padding: 8px 0;
}

.nav-menu {
  display: flex;
  flex-direction: column;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 44px;
  padding: 0 12px;
  margin: 2px 8px;
  border-radius: 6px;
  color: rgba(255, 255, 255, 0.75);
  cursor: pointer;
  user-select: none;
}

.sidebar.collapsed .nav-item {
  justify-content: center;
  padding: 0;
  margin: 2px 8px;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.06);
}

.nav-item.is-active {
  background: rgba(64, 158, 255, 0.2);
  color: #fff;
}

.nav-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.nav-label {
  flex: 1;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.nav-group-title {
  position: relative;
}

.nav-arrow {
  font-size: 12px;
  flex-shrink: 0;
  transition: none;
}

.nav-arrow.open {
  transform: rotate(180deg);
}

.nav-children .nav-child {
  padding-left: 40px;
  height: 40px;
}
</style>

<style>
.sidebar-submenu-popper {
  padding: 4px 0 !important;
}

.sidebar-submenu-popper .popover-menu {
  display: flex;
  flex-direction: column;
}

.sidebar-submenu-popper .popover-item {
  padding: 8px 16px;
  font-size: 14px;
  cursor: pointer;
  color: #303133;
}

.sidebar-submenu-popper .popover-item:hover {
  background: #f5f7fa;
}

.sidebar-submenu-popper .popover-item.is-active {
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}
</style>
