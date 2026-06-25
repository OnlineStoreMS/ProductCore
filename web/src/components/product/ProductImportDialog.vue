<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, type UploadFile, type UploadInstance } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { importProduct } from '../../api/product'

defineProps<{
  brands: { id: number; name: string }[]
  categoryOptions: { id: number; name: string }[]
}>()

const visible = defineModel<boolean>({ default: false })
const emit = defineEmits<{ success: [] }>()

const uploading = ref(false)
const uploadRef = ref<UploadInstance>()
const selectedFile = ref<File | null>(null)

const form = ref({
  name: '',
  subTitle: '',
  brandId: undefined as number | undefined,
  categoryId: undefined as number | undefined,
})

const canSubmit = computed(
  () =>
    !!form.value.name.trim() &&
    !!form.value.brandId &&
    !!form.value.categoryId &&
    !!selectedFile.value &&
    !uploading.value,
)

function resetForm() {
  form.value = { name: '', subTitle: '', brandId: undefined, categoryId: undefined }
  selectedFile.value = null
  uploadRef.value?.clearFiles()
}

function onClose() {
  if (!uploading.value) resetForm()
}

function onFileChange(file: UploadFile) {
  const raw = file.raw
  if (!raw) return
  if (!raw.name.toLowerCase().endsWith('.zip')) {
    ElMessage.warning('请上传 .zip 文件')
    uploadRef.value?.clearFiles()
    selectedFile.value = null
    return
  }
  selectedFile.value = raw
}

function onFileRemove() {
  selectedFile.value = null
}

async function handleImport() {
  if (!canSubmit.value || !selectedFile.value) return
  uploading.value = true
  try {
    const fd = new FormData()
    fd.append('name', form.value.name.trim())
    fd.append('subTitle', form.value.subTitle.trim())
    fd.append('brandId', String(form.value.brandId))
    fd.append('categoryId', String(form.value.categoryId))
    fd.append('file', selectedFile.value)
    await importProduct(fd)
    ElMessage.success('商品导入成功')
    visible.value = false
    resetForm()
    emit('success')
  } catch (e) {
    ElMessage.error((e as Error).message || '导入失败')
  } finally {
    uploading.value = false
  }
}
</script>

<template>
  <el-dialog
    v-model="visible"
    title="导入商品"
    width="560px"
    destroy-on-close
    :close-on-click-modal="!uploading"
    :close-on-press-escape="!uploading"
    @close="onClose"
  >
    <el-form label-width="88px" :disabled="uploading">
      <el-form-item label="商品名称" required>
        <el-input v-model="form.name" placeholder="请输入商品名称" maxlength="100" show-word-limit />
      </el-form-item>
      <el-form-item label="短标题">
        <el-input v-model="form.subTitle" placeholder="导购短标题" maxlength="255" show-word-limit />
      </el-form-item>
      <el-form-item label="品牌" required>
        <el-select v-model="form.brandId" placeholder="请选择品牌" clearable filterable style="width: 100%">
          <el-option v-for="b in brands" :key="b.id" :label="b.name" :value="b.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="分类" required>
        <el-select v-model="form.categoryId" placeholder="请选择分类" clearable filterable style="width: 100%">
          <el-option v-for="c in categoryOptions" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="导入文件" required>
        <el-upload
          ref="uploadRef"
          drag
          :auto-upload="false"
          :limit="1"
          accept=".zip"
          :disabled="uploading"
          @change="onFileChange"
          @remove="onFileRemove"
        >
          <el-icon class="upload-icon"><UploadFilled /></el-icon>
          <div class="el-upload__text">将 zip 文件拖到此处，或<em>点击上传</em></div>
          <template #tip>
            <div class="upload-tip">
              zip 内应包含一个文件夹，格式如「淘宝_商品ID_723068433145」，内含主图、SKU、详情图及可选视频文件夹
            </div>
          </template>
        </el-upload>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button :disabled="uploading" @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="uploading" :disabled="!canSubmit" @click="handleImport">
        开始导入
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.upload-icon {
  font-size: 48px;
  color: var(--el-color-primary);
  margin-bottom: 8px;
}
.upload-tip {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.5;
}
</style>
