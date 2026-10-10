<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  url: string
}>()

const text = ref('')

function formatBytes(n: number) {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(2)} MB`
}

function readSize(url: string) {
  return new Promise<{ width: number; height: number }>((resolve, reject) => {
    const img = new Image()
    img.onload = () => resolve({ width: img.naturalWidth, height: img.naturalHeight })
    img.onerror = () => reject(new Error('image'))
    img.src = url
  })
}

async function readBytes(url: string) {
  try {
    const head = await fetch(url, { method: 'HEAD' })
    const fromHead = Number(head.headers.get('content-length') || 0)
    if (fromHead > 0) return fromHead
    const body = await fetch(url)
    if (!body.ok) return 0
    return (await body.blob()).size
  } catch {
    return 0
  }
}

watch(
  () => props.url,
  async (url) => {
    text.value = ''
    if (!url) return
    const current = url
    try {
      const [size, bytes] = await Promise.all([readSize(url), readBytes(url)])
      if (props.url !== current) return
      const parts = [`${size.width}×${size.height}`]
      if (bytes > 0) parts.push(formatBytes(bytes))
      text.value = parts.join(' · ')
    } catch {
      if (props.url === current) text.value = ''
    }
  },
  { immediate: true },
)
</script>

<template>
  <div v-if="text" class="meta">{{ text }}</div>
</template>

<style scoped>
.meta {
  margin-top: 6px;
  font-size: 12px;
  line-height: 1.4;
  color: #909399;
}
</style>
