<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { eraseProductImage, snapProductImage } from '../../api/ai'
import type { UploadContext } from '../../api/upload'

type Limit = { box: number[]; polygon: number[][] | null }
type Drag =
  | { kind: 'box'; start: number[]; current: number[] }
  | { kind: 'lasso'; points: number[][] }
  | null

const props = defineProps<{
  modelValue: boolean
  source: string
  uploadContext?: UploadContext
  canAdd: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  replace: [url: string]
  add: [url: string]
}>()

const canvasRef = ref<HTMLCanvasElement>()
const wrapRef = ref<HTMLDivElement>()
const shape = ref<'box' | 'contour'>('box')
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

const canCommit = computed(() => history.value.length > 0 && !!workingUrl.value && !busy.value)
const undoLabel = computed(() => (history.value.length ? `撤销（${history.value.length}）` : '撤销'))

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    workingUrl.value = props.source
    history.value = []
    clearMarks()
    shape.value = 'box'
    status.value = '在图上框选或圈选后点擦除'
    nextTick(() => loadUrl(props.source, true))
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

function draw() {
  const el = canvas()
  const box = wrap()
  if (!el) return
  const ctx = el.getContext('2d')
  if (!ctx) return
  ctx.clearRect(0, 0, el.width, el.height)
  if (image) ctx.drawImage(image, 0, 0)
  ctx.lineWidth = zoom > 0 ? 1 / zoom : 1
  ctx.fillStyle = 'rgba(198, 40, 40, 0.35)'
  ctx.strokeStyle = '#c62828'
  boxes.value.forEach((item, i) => {
    const outline = contours.value[i]
    if (outline && outline.length) {
      ctx.beginPath()
      outline.forEach((poly) => {
        ctx.moveTo(poly[0][0], poly[0][1])
        for (let p = 1; p < poly.length; p++) ctx.lineTo(poly[p][0], poly[p][1])
        ctx.closePath()
      })
      ctx.fill('evenodd')
      ctx.stroke()
      return
    }
    const polygon = limits.value[i]?.polygon
    if (shapes.value[i] === 'contour' && polygon) {
      ctx.beginPath()
      ctx.moveTo(polygon[0][0], polygon[0][1])
      for (let p = 1; p < polygon.length; p++) ctx.lineTo(polygon[p][0], polygon[p][1])
      ctx.closePath()
      ctx.fill()
      ctx.stroke()
      return
    }
    const x = Math.min(item[0], item[2])
    const y = Math.min(item[1], item[3])
    ctx.fillRect(x, y, Math.abs(item[2] - item[0]), Math.abs(item[3] - item[1]))
    ctx.strokeRect(x, y, Math.abs(item[2] - item[0]), Math.abs(item[3] - item[1]))
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
  if (zoom >= 8 && image && box) {
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

function loadUrl(url: string, resetView: boolean) {
  if (!url) return
  const img = new Image()
  img.onload = () => {
    image = img
    const el = canvas()
    if (!el) return
    clearMarks()
    if (resetView || el.width !== img.naturalWidth || el.height !== img.naturalHeight) {
      el.width = img.naturalWidth
      el.height = img.naturalHeight
      nextTick(() => fitZoom())
    } else {
      draw()
    }
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

function onPointerUp() {
  if (pan) {
    pan = null
    return
  }
  if (!drag) return
  const done = drag
  drag = null
  if (done.kind === 'box') {
    const [x1, y1] = done.start
    const [x2, y2] = done.current
    if (Math.abs(x2 - x1) >= 4 && Math.abs(y2 - y1) >= 4) {
      const box = [x1, y1, x2, y2]
      boxes.value.push(box)
      limits.value.push({ box, polygon: null })
      labels.value.push('手画')
      contours.value.push(null)
      shapes.value.push('box')
    }
  } else if (done.points.length >= 3) {
    const xs = done.points.map((p) => p[0])
    const ys = done.points.map((p) => p[1])
    const box = [Math.min(...xs), Math.min(...ys), Math.max(...xs), Math.max(...ys)]
    if (box[2] - box[0] >= 4 && box[3] - box[1] >= 4) {
      boxes.value.push(box)
      limits.value.push({ box, polygon: done.points })
      labels.value.push('手画')
      contours.value.push(null)
      shapes.value.push('contour')
      snapAt(boxes.value.length - 1)
    }
  }
  draw()
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
        </div>
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
      </div>
      <aside class="tool">
        <div class="tool-title">框选擦除</div>
        <el-select v-model="shape" :disabled="busy">
          <el-option label="框选，适合周边空白多" value="box" />
          <el-option label="圈选，画圈后按色差收紧" value="contour" />
        </el-select>
        <p class="hint">{{ status }} 选区以外的像素保持原图。鼠标停在哪，就以那里为中心放大。</p>
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
        <el-button :disabled="!canCommit" @click="emit('replace', workingUrl)">替换当前图</el-button>
        <el-button :disabled="!canCommit || !canAdd" @click="emit('add', workingUrl)">添加保存</el-button>
        <p class="hint">替换或添加后，点页面上的保存才会写进这件商品。</p>
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

.canvas-wrap {
  flex: 1;
  min-height: 460px;
  max-height: 68vh;
  overflow: auto;
  background: #eef2f5;
  border-radius: 8px;
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
