<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'

const props = defineProps<{
  src: string
  aspectW: number
  aspectH: number
  title?: string
}>()

const emit = defineEmits<{
  confirm: [crop: { x: number; y: number; w: number; h: number }]
  skipCenter: []
  cancel: []
}>()

const wrap = ref<HTMLElement | null>(null)
const nat = reactive({ w: 0, h: 0 })
const box = reactive({ x: 0, y: 0, w: 0, h: 0 }) // 相对原图像素
const dragging = ref(false)
const dragMode = ref<'move' | 'resize'>('move')
const start = reactive({ mx: 0, my: 0, x: 0, y: 0, w: 0, h: 0 })

const aspect = computed(() => props.aspectW / props.aspectH)

const display = computed(() => {
  const maxW = 520
  const maxH = 420
  if (!nat.w || !nat.h) return { w: maxW, h: maxH, scale: 1 }
  const scale = Math.min(maxW / nat.w, maxH / nat.h, 1)
  return { w: Math.round(nat.w * scale), h: Math.round(nat.h * scale), scale }
})

const overlayStyle = computed(() => {
  const s = display.value.scale || 1
  return {
    left: `${box.x * s}px`,
    top: `${box.y * s}px`,
    width: `${box.w * s}px`,
    height: `${box.h * s}px`,
  }
})

function initCrop() {
  if (!nat.w || !nat.h) return
  let cw: number
  let ch: number
  if (nat.w / nat.h > aspect.value) {
    ch = nat.h
    cw = Math.round(nat.h * aspect.value)
  } else {
    cw = nat.w
    ch = Math.round(nat.w / aspect.value)
  }
  box.w = cw
  box.h = ch
  box.x = Math.round((nat.w - cw) / 2)
  box.y = Math.round((nat.h - ch) / 2)
}

function onImgLoad(e: Event) {
  const img = e.target as HTMLImageElement
  nat.w = img.naturalWidth
  nat.h = img.naturalHeight
  initCrop()
}

function onPointerDown(e: PointerEvent, mode: 'move' | 'resize') {
  e.preventDefault()
  e.stopPropagation()
  dragging.value = true
  dragMode.value = mode
  start.mx = e.clientX
  start.my = e.clientY
  start.x = box.x
  start.y = box.y
  start.w = box.w
  start.h = box.h
  ;(e.currentTarget as HTMLElement).setPointerCapture?.(e.pointerId)
}

function onPointerMove(e: PointerEvent) {
  if (!dragging.value) return
  const s = display.value.scale || 1
  const dx = (e.clientX - start.mx) / s
  const dy = (e.clientY - start.my) / s
  if (dragMode.value === 'move') {
    box.x = Math.min(Math.max(0, Math.round(start.x + dx)), nat.w - box.w)
    box.y = Math.min(Math.max(0, Math.round(start.y + dy)), nat.h - box.h)
    return
  }
  // 右下角缩放，保持比例
  let nw = Math.round(start.w + dx)
  let nh = Math.round(nw / aspect.value)
  const minSide = 40
  if (nw < minSide) {
    nw = minSide
    nh = Math.round(nw / aspect.value)
  }
  if (start.x + nw > nat.w) {
    nw = nat.w - start.x
    nh = Math.round(nw / aspect.value)
  }
  if (start.y + nh > nat.h) {
    nh = nat.h - start.y
    nw = Math.round(nh * aspect.value)
  }
  box.w = Math.max(minSide, nw)
  box.h = Math.max(minSide, nh)
}

function onPointerUp() {
  dragging.value = false
}

function confirm() {
  if (!nat.w || !nat.h) return
  emit('confirm', {
    x: box.x / nat.w,
    y: box.y / nat.h,
    w: box.w / nat.w,
    h: box.h / nat.h,
  })
}

watch(
  () => [props.src, props.aspectW, props.aspectH],
  () => {
    nat.w = 0
    nat.h = 0
  },
)

onMounted(() => {
  // img onload handles init
})
</script>

<template>
  <div class="crop-step">
    <div class="crop-title">{{ title || `裁剪为 ${aspectW}:${aspectH}` }}</div>
    <p class="crop-hint">拖动框调整位置，右下角拖动缩放；也可一键居中裁剪。</p>
    <div
      ref="wrap"
      class="crop-stage"
      :style="{ width: `${display.w}px`, height: `${display.h}px` }"
      @pointermove="onPointerMove"
      @pointerup="onPointerUp"
      @pointercancel="onPointerUp"
    >
      <img :src="src" class="crop-img" alt="裁剪" draggable="false" @load="onImgLoad" />
      <div
        v-if="nat.w"
        class="crop-box"
        :style="overlayStyle"
        @pointerdown="onPointerDown($event, 'move')"
      >
        <span class="crop-handle" @pointerdown="onPointerDown($event, 'resize')" />
      </div>
    </div>
    <div class="crop-actions">
      <el-button @click="emit('cancel')">取消</el-button>
      <el-button @click="emit('skipCenter')">居中裁剪</el-button>
      <el-button type="primary" @click="confirm">使用此裁剪</el-button>
    </div>
  </div>
</template>

<style scoped>
.crop-step { text-align: center; }
.crop-title { font-weight: 600; margin-bottom: 6px; }
.crop-hint { margin: 0 0 12px; font-size: 12px; color: #909399; }
.crop-stage {
  position: relative;
  margin: 0 auto;
  background: #1a1a1a;
  border-radius: 8px;
  overflow: hidden;
  touch-action: none;
  user-select: none;
}
.crop-img {
  width: 100%;
  height: 100%;
  object-fit: fill;
  display: block;
  pointer-events: none;
}
.crop-box {
  position: absolute;
  border: 2px solid #409eff;
  box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.45);
  cursor: move;
  box-sizing: border-box;
}
.crop-handle {
  position: absolute;
  right: -6px;
  bottom: -6px;
  width: 14px;
  height: 14px;
  background: #409eff;
  border-radius: 2px;
  cursor: nwse-resize;
}
.crop-actions {
  margin-top: 14px;
  display: flex;
  justify-content: center;
  gap: 8px;
}
</style>
