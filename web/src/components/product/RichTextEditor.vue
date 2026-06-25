<script setup lang="ts">
import '@wangeditor/editor/dist/css/style.css'
import { onBeforeUnmount, ref, shallowRef, watch } from 'vue'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import type { IDomEditor } from '@wangeditor/editor'

const model = defineModel<string>({ default: '' })
const emit = defineEmits<{ blur: [] }>()

const editorRef = shallowRef<IDomEditor>()
const editorWrapRef = ref<HTMLElement | null>(null)
let anchor: Comment | null = null

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
    <Toolbar :editor="editorRef" :default-config="toolbarConfig" mode="default" class="rich-toolbar" />
    <Editor
      v-model="model"
      :default-config="editorConfig"
      mode="default"
      class="rich-body"
      @on-created="handleCreated"
    />
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
