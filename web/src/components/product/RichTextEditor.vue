<script setup lang="ts">
import '@wangeditor/editor/dist/css/style.css'
import { computed, onBeforeUnmount, ref, shallowRef, watch } from 'vue'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import type { IDomEditor } from '@wangeditor/editor'

const model = defineModel<string>({ default: '' })
const emit = defineEmits<{ blur: [] }>()

const editorRef = shallowRef<IDomEditor>()
const editorWrapRef = ref<HTMLElement | null>(null)
const phoneView = ref(false)
let anchor: Comment | null = null

const previewHtml = computed(() => {
  const html = (model.value || '').trim()
  if (!html || html === '<p><br></p>' || html === '<p></p>') return ''
  return html
})

function showPhone() {
  if (editorRef.value?.isFullScreen) editorRef.value.unFullScreen()
  phoneView.value = true
}

const toolbarConfig = {}
const editorConfig = { placeholder: '请输入商品详情...' }

function mountFullscreenToBody() {
  const el = editorWrapRef.value
  if (!el || el.parentElement === document.body) return

  anchor = document.createComment('rich-editor-anchor')
  el.parentElement?.insertBefore(anchor, el)
  document.body.appendChild(el)
  document.body.classList.add('rich-editor-fullscreen-active')
}

function restoreFromFullscreen() {
  const el = editorWrapRef.value
  if (!el || !anchor?.parentElement) return

  anchor.parentElement.insertBefore(el, anchor)
  anchor.remove()
  anchor = null
  document.body.classList.remove('rich-editor-fullscreen-active')
}

function handleCreated(editor: IDomEditor) {
  editorRef.value = editor
  editor.on('fullScreen', mountFullscreenToBody)
  editor.on('unFullScreen', restoreFromFullscreen)
  editor.on('blur', () => emit('blur'))
}


watch(
  () => model.value,
  (html) => {
    const editor = editorRef.value
    if (!editor || editor.isDestroyed) return
    const current = editor.getHtml()
    if ((html || '') !== current) {
      editor.setHtml(html || '')
    }
  },
)

onBeforeUnmount(() => {
  if (editorRef.value?.isFullScreen) {
    restoreFromFullscreen()
  }
  editorRef.value?.destroy()
})
</script>

<template>
  <div ref="editorWrapRef" class="rich-editor">
    <div class="rich-mode">
      <el-button size="small" :type="phoneView ? 'default' : 'primary'" @click="phoneView = false">编辑</el-button>
      <el-button size="small" :type="phoneView ? 'primary' : 'default'" @click="showPhone">手机浏览</el-button>
      <span v-if="phoneView" class="mode-hint">按 375 宽的手机屏幕查看，图片随屏幕宽度显示</span>
    </div>
    <div v-show="!phoneView">
      <Toolbar :editor="editorRef" :default-config="toolbarConfig" mode="default" class="rich-toolbar" />
      <Editor
        v-model="model"
        :default-config="editorConfig"
        mode="default"
        class="rich-body"
        @on-created="handleCreated"
      />
    </div>
    <div v-show="phoneView" class="phone-stage">
      <div class="phone">
        <div class="phone-screen">
          <div v-if="previewHtml" class="phone-html" v-html="previewHtml" />
          <div v-else class="phone-empty">暂无详情内容</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.rich-editor {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background: #fff;
}

.rich-toolbar {
  border-bottom: 1px solid #dcdfe6;
}

.rich-body {
  min-height: 320px;
  height: 320px;
}

.rich-mode {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-bottom: 1px solid #dcdfe6;
  background: #fafafa;
}

.mode-hint {
  color: #909399;
  font-size: 12px;
}

.phone-stage {
  display: flex;
  justify-content: center;
  padding: 20px 12px 24px;
  background: #eef1f4;
}

.phone {
  width: 375px;
  height: 700px;
  max-height: 72vh;
  padding: 10px;
  border-radius: 28px;
  background: #1f2430;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.18);
}

.phone-screen {
  height: 100%;
  overflow: auto;
  border-radius: 20px;
  background: #fff;
}

.phone-html {
  font-size: 14px;
  line-height: 1.6;
  color: #303133;
  word-break: break-word;
}

.phone-html :deep(img),
.phone-html :deep(video) {
  max-width: 100% !important;
  height: auto !important;
  display: block;
}

.phone-html :deep(p) {
  margin: 0 0 8px;
}

.phone-html :deep(table) {
  max-width: 100%;
  display: block;
  overflow-x: auto;
}

.phone-empty {
  padding: 48px 16px;
  text-align: center;
  color: #909399;
  font-size: 13px;
}
</style>

<style>
.w-e-full-screen-container {
  z-index: 10000 !important;
  background: #fff !important;
}

body.rich-editor-fullscreen-active {
  overflow: hidden;
}
</style>
