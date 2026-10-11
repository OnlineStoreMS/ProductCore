<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { eraseProductImage, snapProductImage } from '../../api/ai'
import type { UploadContext } from '../../api/upload'

type Limit = { box: number[]; polygon: number[][] | null }
type Drag =
  | { kind: 'box'; start: number[]; current: number[] }
  | { kind: 'lasso'; points: number[][] }
  | null

const props = defineProps<{
  modelValue: boolean
  sources: string[]
  index: number
  uploadContext?: UploadContext
  canAdd: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:index': [value: number]
  replace: [url: string]
  add: [url: string]
}>()

const canvasRef = ref<HTMLCanvasElement>()
const wrapRef = ref<HTMLDivElement>()
const shape = ref<'box' | 'contour'>('box')
const angle = ref(0)
const busy = ref(false)
const status = ref('在图上框选或圈选后点擦除')
const zoomLabel = ref('100%')
const workingUrl = ref('')
const history = ref<string[]>([])
const boxes = ref<number[][]>([])
const limits = ref<Limit[]>([])
const contours = ref<Array<number[][][] | null>>([])
const shapes = ref<string[]>([])
const labels = ref<string[]>([])

let image: HTMLImageElement | null = null
let zoom = 1
let pointer: { x: number; y: number } | null = null
let drag: Drag = null
let pan: { x: number; y: number; left: number; top: number } | null = null
let spaceDown = false
let gesture: { x: number; y: number; t: number } | null = null
let loadToken = 0

const canCommit = computed(() => history.value.length > 0 && !!workingUrl.value && !busy.value)
const undoLabel = computed(() => (history.value.length ? `撤销（${history.value.length}）` : '撤销'))
const pageLabel = computed(() => `${props.index + 1} / ${Math.max(props.sources.length, 1)}`)
const canPrev = computed(() => props.index > 0)
const canNext = computed(() => props.index < props.sources.length - 1)

function currentSource() {
  return props.sources[props.index] || ''
}

function showImage(resetView: boolean) {
  const url = currentSource()
  workingUrl.value = url
  history.value = []
  clearMarks()
  if (resetView) {
    shape.value = 'box'
    angle.value = 0
    status.value = '字标斜着时，先把画面转到水平，再画框'
  }
  nextTick(() => loadUrl(url, resetView))
}

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    showImage(true)
  },
)

watch(
  () => props.index,
  () => {
    if (!props.modelValue) return
    showImage(false)
  },
)

function errorText(e: unknown): string {
  const err = e as { response?: { data?: { message?: string } }; message?: string }
  return err.response?.data?.message || err.message || '擦除失败'
}

function clearMarks() {
  boxes.value = []
  limits.value = []
  contours.value = []
  shapes.value = []
  labels.value = []
  drag = null
}

function canvas(): HTMLCanvasElement | undefined {
  return canvasRef.value
}

function wrap(): HTMLDivElement | undefined {
  return wrapRef.value
}

function point(ev: PointerEvent): number[] {
  const el = canvas()
  if (!el) return [0, 0]
  const rect = el.getBoundingClientRect()
  const x = Math.round((ev.clientX - rect.left) * (el.width / rect.width))
  const y = Math.round((ev.clientY - rect.top) * (el.height / rect.height))
  return [Math.max(0, Math.min(el.width, x)), Math.max(0, Math.min(el.height, y))]
}

function viewSize(width: number, height: number, deg: number) {
  const rad = (deg * Math.PI) / 180
  const c = Math.abs(Math.cos(rad))
  const s = Math.abs(Math.sin(rad))
  return {
    width: Math.max(1, Math.ceil(width * c + height * s)),
    height: Math.max(1, Math.ceil(width * s + height * c)),
  }
}

function viewToImage(vx: number, vy: number): number[] {
  if (!image || Math.abs(angle.value) < 0.05) return [vx, vy]
  const w = image.naturalWidth
  const h = image.naturalHeight
  const view = viewSize(w, h, angle.value)
  const rad = (angle.value * Math.PI) / 180
  const cos = Math.cos(rad)
  const sin = Math.sin(rad)
  const dx = vx - view.width / 2
  const dy = vy - view.height / 2
  return [dx * cos + dy * sin + w / 2, -dx * sin + dy * cos + h / 2]
}

function imageToView(ix: number, iy: number): number[] {
  if (!image || Math.abs(angle.value) < 0.05) return [ix, iy]
  const w = image.naturalWidth
  const h = image.naturalHeight
  const view = viewSize(w, h, angle.value)
  const rad = (angle.value * Math.PI) / 180
  const cos = Math.cos(rad)
  const sin = Math.sin(rad)
  const rx = ix - w / 2
  const ry = iy - h / 2
  return [view.width / 2 + rx * cos - ry * sin, view.height / 2 + rx * sin + ry * cos]
}

function clipPoly(points: number[][]): number[][] {
  if (!image) return points
  const w = image.naturalWidth - 1
  const h = image.naturalHeight - 1
  const inside = [
    (p: number[]) => p[0] >= 0,
    (p: number[]) => p[0] <= w,
    (p: number[]) => p[1] >= 0,
    (p: number[]) => p[1] <= h,
  ]
  const cross = [
    (a: number[], b: number[]) => [0, a[1] + ((b[1] - a[1]) * (0 - a[0])) / (b[0] - a[0])],
    (a: number[], b: number[]) => [w, a[1] + ((b[1] - a[1]) * (w - a[0])) / (b[0] - a[0])],
    (a: number[], b: number[]) => [a[0] + ((b[0] - a[0]) * (0 - a[1])) / (b[1] - a[1]), 0],
    (a: number[], b: number[]) => [a[0] + ((b[0] - a[0]) * (h - a[1])) / (b[1] - a[1]), h],
  ]
  let poly = points
  for (let edge = 0; edge < 4; edge++) {
    if (poly.length === 0) return []
    const next: number[][] = []
    for (let i = 0; i < poly.length; i++) {
      const cur = poly[i]
      const prev = poly[(i + poly.length - 1) % poly.length]
      const curIn = inside[edge](cur)
      const prevIn = inside[edge](prev)
      if (curIn) {
        if (!prevIn) next.push(cross[edge](prev, cur))
        next.push(cur)
      } else if (prevIn) {
        next.push(cross[edge](prev, cur))
      }
    }
    poly = next
  }
  return poly
}

function viewRectToLimit(x1: number, y1: number, x2: number, y2: number): Limit | null {
  const left = Math.min(x1, x2)
  const right = Math.max(x1, x2)
  const top = Math.min(y1, y2)
  const bottom = Math.max(y1, y2)
  if (Math.abs(angle.value) < 0.05) return { box: [left, top, right, bottom], polygon: null }
  const raw = [
    viewToImage(left, top),
    viewToImage(right, top),
    viewToImage(right, bottom),
    viewToImage(left, bottom),
  ]
  const polygon = clipPoly(raw).map(([x, y]) => [Math.round(x), Math.round(y)])
  if (polygon.length < 3) return null
  const xs = polygon.map((p) => p[0])
  const ys = polygon.map((p) => p[1])
  return {
    box: [Math.min(...xs), Math.min(...ys), Math.max(...xs), Math.max(...ys)],
    polygon,
  }
}

function paintImagePoly(ctx: CanvasRenderingContext2D, points: number[][], close: boolean) {
  const view = points.map(([x, y]) => imageToView(x, y))
  if (view.length < 2) return
  ctx.beginPath()
  ctx.moveTo(view[0][0], view[0][1])
  for (let i = 1; i < view.length; i++) ctx.lineTo(view[i][0], view[i][1])
  if (close) ctx.closePath()
  ctx.fill()
  ctx.stroke()
}

function draw() {
  const el = canvas()
  const box = wrap()
  if (!el) return
  const ctx = el.getContext('2d')
  if (!ctx) return
  ctx.setTransform(1, 0, 0, 1, 0, 0)
  ctx.clearRect(0, 0, el.width, el.height)
  ctx.fillStyle = '#eef2f5'
  ctx.fillRect(0, 0, el.width, el.height)
  if (image) {
    const rad = (angle.value * Math.PI) / 180
    ctx.imageSmoothingEnabled = Math.abs(angle.value) >= 0.05 || zoom < 8
    ctx.imageSmoothingQuality = 'high'
    ctx.save()
    ctx.translate(el.width / 2, el.height / 2)
    ctx.rotate(rad)
    ctx.drawImage(image, -image.naturalWidth / 2, -image.naturalHeight / 2)
    ctx.restore()
  }
  ctx.lineWidth = zoom > 0 ? 1 / zoom : 1
  ctx.fillStyle = 'rgba(198, 40, 40, 0.35)'
  ctx.strokeStyle = '#c62828'
  boxes.value.forEach((item, i) => {
    const outline = contours.value[i]
    if (outline && outline.length) {
      ctx.beginPath()
      outline.forEach((poly) => {
        const view = poly.map(([x, y]) => imageToView(x, y))
        ctx.moveTo(view[0][0], view[0][1])
        for (let p = 1; p < view.length; p++) ctx.lineTo(view[p][0], view[p][1])
        ctx.closePath()
      })
      ctx.fill('evenodd')
      ctx.stroke()
      return
    }
    const polygon = limits.value[i]?.polygon
    if (polygon && polygon.length >= 3) {
      paintImagePoly(ctx, polygon, true)
      return
    }
    const x1 = Math.min(item[0], item[2])
    const y1 = Math.min(item[1], item[3])
    const x2 = Math.max(item[0], item[2])
    const y2 = Math.max(item[1], item[3])
    paintImagePoly(ctx, [[x1, y1], [x2, y1], [x2, y2], [x1, y2]], true)
  })
  if (drag?.kind === 'box') {
    const x = Math.min(drag.start[0], drag.current[0])
    const y = Math.min(drag.start[1], drag.current[1])
    ctx.strokeRect(x, y, Math.abs(drag.current[0] - drag.start[0]), Math.abs(drag.current[1] - drag.start[1]))
  } else if (drag?.kind === 'lasso' && drag.points.length > 1) {
    ctx.beginPath()
    ctx.moveTo(drag.points[0][0], drag.points[0][1])
    for (let p = 1; p < drag.points.length; p++) ctx.lineTo(drag.points[p][0], drag.points[p][1])
    ctx.stroke()
  }
  if (zoom >= 8 && image && box && Math.abs(angle.value) < 0.05) {
    const x0 = Math.max(0, Math.floor(box.scrollLeft / zoom) - 1)
    const y0 = Math.max(0, Math.floor(box.scrollTop / zoom) - 1)
    const x1 = Math.min(el.width, Math.ceil((box.scrollLeft + box.clientWidth) / zoom) + 1)
    const y1 = Math.min(el.height, Math.ceil((box.scrollTop + box.clientHeight) / zoom) + 1)
    ctx.strokeStyle = 'rgba(30, 41, 59, 0.28)'
    ctx.beginPath()
    for (let x = x0; x <= x1; x++) {
      ctx.moveTo(x, y0)
      ctx.lineTo(x, y1)
    }
    for (let y = y0; y <= y1; y++) {
      ctx.moveTo(x0, y)
      ctx.lineTo(x1, y)
    }
    ctx.stroke()
  }
}

function setZoom(next: number, clientX?: number, clientY?: number) {
  const el = canvas()
  const box = wrap()
  if (!el || !box) return
  const rect = el.getBoundingClientRect()
  const wrapRect = box.getBoundingClientRect()
  const anchorX = clientX ?? wrapRect.left + box.clientWidth / 2
  const anchorY = clientY ?? wrapRect.top + box.clientHeight / 2
  const px = rect.width ? ((anchorX - rect.left) / rect.width) * el.width : el.width / 2
  const py = rect.height ? ((anchorY - rect.top) / rect.height) * el.height : el.height / 2
  zoom = Math.min(32, Math.max(0.2, next))
  const displayW = Math.max(1, Math.round(el.width * zoom))
  const displayH = Math.max(1, Math.round(el.height * zoom))
  el.style.width = `${displayW}px`
  el.style.height = `${displayH}px`
  zoomLabel.value = `${Math.round(zoom * 100)}%`
  void el.offsetWidth
  box.scrollLeft = (px / el.width) * el.offsetWidth - (anchorX - wrapRect.left)
  box.scrollTop = (py / el.height) * el.offsetHeight - (anchorY - wrapRect.top)
  draw()
}

function fitZoom() {
  const el = canvas()
  const box = wrap()
  if (!el || !box) return
  const next = Math.min(1, (box.clientWidth - 4) / el.width)
  if (pointer) setZoom(next, pointer.x, pointer.y)
  else setZoom(next)
}

function layoutCanvas(refit: boolean) {
  const el = canvas()
  if (!el || !image) return
  const size = viewSize(image.naturalWidth, image.naturalHeight, angle.value)
  if (el.width !== size.width || el.height !== size.height) {
    el.width = size.width
    el.height = size.height
  }
  if (refit) {
    nextTick(() => fitZoom())
    return
  }
  el.style.width = `${Math.max(1, Math.round(el.width * zoom))}px`
  el.style.height = `${Math.max(1, Math.round(el.height * zoom))}px`
  draw()
}

function setAngle(next: number | number[]) {
  const value = Array.isArray(next) ? next[0] : next
  const clamped = Math.max(-180, Math.min(180, Math.round(value * 10) / 10))
  if (clamped === angle.value) return
  drag = null
  limits.value.forEach((limit, i) => {
    if (shapes.value[i] !== 'box' || limit.polygon) return
    const x1 = Math.min(limit.box[0], limit.box[2])
    const y1 = Math.min(limit.box[1], limit.box[3])
    const x2 = Math.max(limit.box[0], limit.box[2])
    const y2 = Math.max(limit.box[1], limit.box[3])
    limit.polygon = [[x1, y1], [x2, y1], [x2, y2], [x1, y2]]
  })
  angle.value = clamped
  layoutCanvas(false)
}

function loadUrl(url: string, resetView: boolean) {
  if (!url) return
  const token = ++loadToken
  const img = new Image()
  img.onload = () => {
    if (token !== loadToken) return
    image = img
    const el = canvas()
    if (!el) return
    clearMarks()
    if (resetView) angle.value = 0
    const size = viewSize(img.naturalWidth, img.naturalHeight, angle.value)
    const same = !resetView && el.width === size.width && el.height === size.height
    layoutCanvas(!same)
  }
  img.src = url
}

function onPointerDown(ev: PointerEvent) {
  const el = canvas()
  const box = wrap()
  if (!el || !image || busy.value) return
  if (ev.button === 1 || spaceDown) {
    ev.preventDefault()
    pan = { x: ev.clientX, y: ev.clientY, left: box?.scrollLeft || 0, top: box?.scrollTop || 0 }
    el.setPointerCapture(ev.pointerId)
    return
  }
  if (ev.button !== 0) return
  gesture = { x: ev.clientX, y: ev.clientY, t: performance.now() }
  const p = point(ev)
  drag = shape.value === 'contour' ? { kind: 'lasso', points: [p] } : { kind: 'box', start: p, current: p }
  el.setPointerCapture(ev.pointerId)
}

function onPointerMove(ev: PointerEvent) {
  pointer = { x: ev.clientX, y: ev.clientY }
  const box = wrap()
  if (pan && box) {
    box.scrollLeft = pan.left - (ev.clientX - pan.x)
    box.scrollTop = pan.top - (ev.clientY - pan.y)
    return
  }
  if (!drag) return
  const p = point(ev)
  if (drag.kind === 'lasso') {
    const last = drag.points[drag.points.length - 1]
    if (Math.abs(p[0] - last[0]) + Math.abs(p[1] - last[1]) >= 1) drag.points.push(p)
  } else {
    drag.current = p
  }
  draw()
}

function onPointerUp(ev: PointerEvent) {
  if (pan) {
    pan = null
    gesture = null
    return
  }
  if (gesture) {
    const dx = ev.clientX - gesture.x
    const dy = ev.clientY - gesture.y
    const dt = performance.now() - gesture.t
    gesture = null
    if (dt < 420 && Math.abs(dx) >= 72 && Math.abs(dy) <= 40 && Math.abs(dx) > Math.abs(dy) * 1.6) {
      drag = null
      draw()
      void shiftImage(dx < 0 ? 1 : -1)
      return
    }
  }
  if (!drag) return
  const done = drag
  drag = null
  if (done.kind === 'box') {
    const [x1, y1] = done.start
    const [x2, y2] = done.current
    if (Math.abs(x2 - x1) >= 4 && Math.abs(y2 - y1) >= 4) {
      const limit = viewRectToLimit(x1, y1, x2, y2)
      if (limit) {
        boxes.value.push(limit.box)
        limits.value.push(limit)
        labels.value.push('手画')
        contours.value.push(null)
        shapes.value.push('box')
      }
    }
  } else if (done.points.length >= 3) {
    const xs = done.points.map((p) => p[0])
    const ys = done.points.map((p) => p[1])
    if (Math.max(...xs) - Math.min(...xs) >= 4 && Math.max(...ys) - Math.min(...ys) >= 4) {
      const polygon = clipPoly(done.points.map(([x, y]) => viewToImage(x, y))).map(([x, y]) => [Math.round(x), Math.round(y)])
      if (polygon.length >= 3) {
        const ixs = polygon.map((p) => p[0])
        const iys = polygon.map((p) => p[1])
        const box = [Math.min(...ixs), Math.min(...iys), Math.max(...ixs), Math.max(...iys)]
        boxes.value.push(box)
        limits.value.push({ box, polygon })
        labels.value.push('手画')
        contours.value.push(null)
        shapes.value.push('contour')
        snapAt(boxes.value.length - 1)
      }
    }
  }
  draw()
}

function replaceCurrent() {
  const url = workingUrl.value
  if (!url || history.value.length === 0 || busy.value) return
  history.value = []
  status.value = '已替换当前图，可继续左右滑'
  emit('replace', url)
}

async function shiftImage(delta: number) {
  if (busy.value || !props.sources.length) return
  const next = props.index + delta
  if (next < 0 || next >= props.sources.length) {
    ElMessage.info(delta > 0 ? '已经是最后一张' : '已经是第一张')
    return
  }
  if (history.value.length > 0) {
    try {
      await ElMessageBox.confirm('这张还有未替换的擦除，切换后会丢掉这次结果。', '切换图片', {
        confirmButtonText: '切换',
        cancelButtonText: '留在这张',
        type: 'warning',
      })
    } catch {
      return
    }
  }
  emit('update:index', next)
}

async function snapAt(index: number) {
  const limit = limits.value[index]
  if (!limit || !workingUrl.value) return
  try {
    const data = await snapProductImage({
      imageUrl: workingUrl.value,
      boxes: [limit.box],
      polygons: [limit.polygon],
    })
    const region = data.regions?.[0]
    if (!region?.contours?.length) {
      status.value = '圈内没有收紧，将按画出的范围擦除'
      return
    }
    contours.value[index] = region.contours
    draw()
  } catch (e) {
    status.value = errorText(e)
  }
}

async function runErase() {
  if (!workingUrl.value || boxes.value.length === 0) {
    ElMessage.warning('请先框选或圈选')
    return
  }
  busy.value = true
  try {
    const data = await eraseProductImage({
      imageUrl: workingUrl.value,
      boxes: limits.value.map((item) => item.box),
      polygons: limits.value.map((item) => item.polygon),
      shapes: shapes.value,
      scope: props.uploadContext?.scope,
      resource: props.uploadContext?.resource,
      productId: props.uploadContext?.productId,
      skuId: props.uploadContext?.skuId,
    })
    history.value.push(workingUrl.value)
    workingUrl.value = data.url
    status.value = '已擦除，可继续修，或替换当前图、添加保存'
    loadUrl(data.url, false)
  } catch (e) {
    ElMessage.error(errorText(e))
  } finally {
    busy.value = false
  }
}

function undoImage() {
  const previous = history.value.pop()
  if (!previous) {
    ElMessage.info('没有可撤销的擦除')
    return
  }
  workingUrl.value = previous
  status.value = history.value.length ? `已回到上一层，还可撤销 ${history.value.length} 层` : '已回到这张原图'
  loadUrl(previous, false)
}

function onWheel(ev: WheelEvent) {
  if (!image) return
  ev.preventDefault()
  setZoom(zoom * (ev.deltaY < 0 ? 1.2 : 1 / 1.2), ev.clientX, ev.clientY)
}

function onKeyDown(ev: KeyboardEvent) {
  const typing = ev.target instanceof HTMLElement && /^(INPUT|TEXTAREA|SELECT)$/.test(ev.target.tagName)
  if (!typing && (ev.ctrlKey || ev.metaKey) && ev.key === 'z' && !ev.shiftKey) {
    ev.preventDefault()
    undoImage()
    return
  }
  const onSlider = ev.target instanceof HTMLElement && !!ev.target.closest('.el-slider')
  if (!typing && !onSlider && ev.key === 'ArrowLeft') {
    ev.preventDefault()
    void shiftImage(-1)
    return
  }
  if (!typing && !onSlider && ev.key === 'ArrowRight') {
    ev.preventDefault()
    void shiftImage(1)
    return
  }
  if (typing || ev.code !== 'Space') return
  ev.preventDefault()
  spaceDown = true
}

function onKeyUp(ev: KeyboardEvent) {
  if (ev.code === 'Space') spaceDown = false
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      window.addEventListener('keydown', onKeyDown)
      window.addEventListener('keyup', onKeyUp)
    } else {
      window.removeEventListener('keydown', onKeyDown)
      window.removeEventListener('keyup', onKeyUp)
    }
  },
)

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeyDown)
  window.removeEventListener('keyup', onKeyUp)
})
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    title="擦除图片"
    width="1080px"
    append-to-body
    destroy-on-close
    class="image-erase-dialog"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="erase-layout">
      <div class="stage">
        <div class="toolbar">
          <el-button size="small" @click="pointer ? setZoom(zoom / 1.25, pointer.x, pointer.y) : setZoom(zoom / 1.25)">缩小</el-button>
          <el-button size="small" @click="pointer ? setZoom(zoom * 1.25, pointer.x, pointer.y) : setZoom(zoom * 1.25)">放大</el-button>
          <el-button size="small" @click="pointer ? setZoom(1, pointer.x, pointer.y) : setZoom(1)">100%</el-button>
          <el-button size="small" @click="fitZoom">适应</el-button>
          <span class="zoom">{{ zoomLabel }}</span>
          <span class="zoom">画面 {{ angle.toFixed(1) }}°</span>
          <span class="zoom">{{ pageLabel }}</span>
        </div>
        <div class="canvas-frame">
          <button
            v-if="sources.length > 1"
            type="button"
            class="nav-btn prev"
            title="上一张"
            :disabled="!canPrev || busy"
            @click="shiftImage(-1)"
          >
            ‹
          </button>
          <div ref="wrapRef" class="canvas-wrap" @wheel.prevent="onWheel">
            <canvas
              ref="canvasRef"
              width="960"
              height="640"
              @pointerdown="onPointerDown"
              @pointermove="onPointerMove"
              @pointerup="onPointerUp"
            />
          </div>
          <button
            v-if="sources.length > 1"
            type="button"
            class="nav-btn next"
            title="下一张"
            :disabled="!canNext || busy"
            @click="shiftImage(1)"
          >
            ›
          </button>
        </div>
      </div>
      <aside class="tool">
        <div class="tool-title">框选擦除</div>
        <el-select v-model="shape" :disabled="busy">
          <el-option label="框选，适合周边空白多" value="box" />
          <el-option label="圈选，画圈后按色差收紧" value="contour" />
        </el-select>
        <div class="rotate-label">画面角度</div>
        <el-slider
          :model-value="angle"
          :min="-180"
          :max="180"
          :step="0.1"
          :disabled="busy"
          @update:model-value="setAngle"
        />
        <div class="row">
          <el-button size="small" :disabled="busy" @click="setAngle(angle - 1)">左转</el-button>
          <el-button size="small" :disabled="busy" @click="setAngle(angle + 1)">右转</el-button>
          <el-button size="small" :disabled="busy || angle === 0" @click="setAngle(0)">复位</el-button>
        </div>
        <p class="hint">{{ status }} 向左滑打开下一张，向右滑打开上一张。先把斜着的字标转到水平，再画正的框。保存后的图片仍是原来的方向，选区以外的像素保持原图。</p>
        <ul>
          <li v-for="(_, i) in boxes" :key="i">
            <span>{{ shapes[i] === 'contour' ? '圈' : '框' }} · {{ labels[i] }}</span>
            <el-button size="small" text @click="boxes.splice(i, 1); limits.splice(i, 1); contours.splice(i, 1); shapes.splice(i, 1); labels.splice(i, 1); draw()">去掉</el-button>
          </li>
        </ul>
        <div class="row">
          <el-button :disabled="!history.length || busy" @click="undoImage">{{ undoLabel }}</el-button>
          <el-button :disabled="busy" @click="clearMarks(); draw()">清空选区</el-button>
        </div>
        <el-button type="primary" :loading="busy" @click="runErase">擦除</el-button>
        <el-button :disabled="!canCommit" @click="replaceCurrent">替换当前图</el-button>
        <el-button :disabled="!canCommit || !canAdd" @click="emit('add', workingUrl)">添加保存</el-button>
        <p class="hint">替换后留在这里，可以继续左右滑。点页面上的保存才会写进这件商品。</p>
      </aside>
    </div>
  </el-dialog>
</template>

<style scoped>
.erase-layout {
  display: flex;
  gap: 16px;
  min-height: 520px;
}

.stage {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.zoom {
  color: #606266;
  font-size: 13px;
}

.canvas-frame {
  position: relative;
  flex: 1;
  min-height: 460px;
  min-width: 0;
  display: flex;
}

.canvas-wrap {
  flex: 1;
  min-width: 0;
  min-height: 460px;
  max-height: 68vh;
  overflow: auto;
  background: #eef2f5;
  border-radius: 8px;
}

.nav-btn {
  position: absolute;
  top: 50%;
  z-index: 2;
  width: 36px;
  height: 72px;
  margin-top: -36px;
  border: none;
  border-radius: 8px;
  background: rgba(15, 23, 42, 0.55);
  color: #fff;
  font-size: 32px;
  line-height: 68px;
  cursor: pointer;
  padding: 0;
}

.nav-btn.prev {
  left: 8px;
}

.nav-btn.next {
  right: 8px;
}

.nav-btn:disabled {
  opacity: 0.35;
  cursor: default;
}

canvas {
  display: block;
  cursor: crosshair;
  touch-action: none;
  background: #eef2f5;
}

.tool {
  width: 240px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tool-title {
  font-weight: 600;
}

.tool :deep(.el-button) {
  margin-left: 0;
}

.rotate-label {
  font-size: 13px;
  color: #303133;
}

.row {
  display: flex;
  gap: 8px;
}

.row :deep(.el-button) {
  flex: 1;
}

ul {
  list-style: none;
  margin: 0;
  padding: 0;
  max-height: 160px;
  overflow: auto;
}

li {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  padding: 4px 0;
  border-bottom: 1px solid #f0f2f4;
}

.hint {
  margin: 0;
  font-size: 12px;
  line-height: 1.5;
  color: #909399;
}
</style>
