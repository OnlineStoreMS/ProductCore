<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowDown, Expand, Fold, HomeFilled, SwitchButton } from '@element-plus/icons-vue'
import Sidebar from './Sidebar.vue'
import SuperSearchPanel from '../components/search/SuperSearchPanel.vue'
import { portalAppsUrl, portalLoginUrl } from '../utils/auth'
import { useSessionStore } from '../stores/session'

const route = useRoute()
const collapsed = ref(false)
const sessionStore = useSessionStore()

const userInitial = computed(() => {
  const name = sessionStore.session?.user.displayName?.trim()
  return name ? name[0].toUpperCase() : '?'
})

onMounted(() => {
  void sessionStore.load()
})

function backToPortal() {
  window.location.href = portalAppsUrl()
}

function logout() {
  sessionStore.clear()
  window.location.href = portalLoginUrl()
}

const breadcrumbs = computed(() => {
  const title = (route.meta.title as string) || 'ProductCore'
  if (route.path === '/products/drafts') {
    return ['商品管理', '商品草稿箱']
  }
  if (route.path === '/products/trash') {
    return ['商品管理', '商品回收站']
  }
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
          <SuperSearchPanel />
          <el-dropdown trigger="click" @command="(cmd: string) => cmd === 'logout' ? logout() : backToPortal()">
            <div class="user-trigger">
              <el-avatar :size="32" class="user-avatar">{{ userInitial }}</el-avatar>
              <div v-if="sessionStore.session" class="user-meta">
                <span class="user-name">{{ sessionStore.session.user.displayName }}</span>
                <span class="tenant-name">{{ sessionStore.session.tenant.name }}</span>
              </div>
              <span v-else class="user-loading">加载中…</span>
              <el-icon class="user-arrow"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item disabled>
                  <div class="dropdown-profile">
                    <div class="dropdown-name">
                      {{ sessionStore.session?.user.displayName }}
                      <el-tag v-if="sessionStore.session?.user.isPlatform" size="small" type="warning">平台</el-tag>
                    </div>
                    <div class="dropdown-email">{{ sessionStore.session?.user.email }}</div>
                    <div class="dropdown-tenant">租户：{{ sessionStore.session?.tenant.name }}</div>
                  </div>
                </el-dropdown-item>
                <el-dropdown-item :icon="HomeFilled" command="portal">应用中心</el-dropdown-item>
                <el-dropdown-item :icon="SwitchButton" command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
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

.user-trigger {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 8px;
  transition: background 0.2s;
}

.user-trigger:hover {
  background: #f5f7fa;
}

.user-avatar {
  background: #409eff;
  color: #fff;
  flex-shrink: 0;
}

.user-meta {
  display: flex;
  flex-direction: column;
  line-height: 1.3;
  max-width: 160px;
}

.user-name {
  font-size: 14px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tenant-name {
  font-size: 12px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-loading {
  font-size: 13px;
  color: #909399;
}

.user-arrow {
  color: #909399;
  font-size: 12px;
}

.dropdown-profile {
  line-height: 1.5;
  padding: 2px 0;
}

.dropdown-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  color: #303133;
}

.dropdown-email,
.dropdown-tenant {
  font-size: 12px;
  color: #909399;
}

.content {
  flex: 1;
  overflow: auto;
  padding: 20px;
}
</style>
