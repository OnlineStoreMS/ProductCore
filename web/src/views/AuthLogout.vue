<script setup lang="ts">
import { onMounted } from 'vue'
import { clearToken, redirectToPortal } from '../utils/auth'

onMounted(() => {
  clearToken()
  // UserCore 退出时通过 iframe 调用，仅清本地 token，不跳转父页面
  if (window.self !== window.top) return
  redirectToPortal()
})
</script>

<template>
  <div class="auth-logout">正在退出…</div>
</template>

<style scoped>
.auth-logout {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
}
</style>
