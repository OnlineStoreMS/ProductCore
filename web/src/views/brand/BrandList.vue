<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Delete, Search, Upload, Document, Link } from '@element-plus/icons-vue'
import { createBrand, deleteBrand, fetchBrands, updateBrand } from '../../api/product'
import { uploadImage } from '../../api/upload'
import type { Brand, BrandAuthFile } from '../../types/product'

const tableData = ref<Brand[]>([])
const keyword = ref('')
const loading = ref(false)
const dialogVisible = ref(false)
const saving = ref(false)
const uploadingAuth = ref(false)
const editing = ref<Partial<Brand>>({})

async function loadData() {
  loading.value = true
  try {
    tableData.value = await fetchBrands(keyword.value || undefined)
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function handleAdd() {
  editing.value = {
    name: '',
    logo: '',
    logoBg: '',
    website: '',
    agent: '',
    authFiles: [],
    firstLetter: '',
    sort: 0,
    showStatus: 1,
  }
  dialogVisible.value = true
}

function logoBgStyle(bg?: string) {
  return bg ? { background: bg } : undefined
}

function onLogoBgInput(val: string) {
  editing.value.logoBg = val.trim()
}

function handleEdit(row: Brand) {
  editing.value = { ...row, authFiles: [...(row.authFiles || [])] }
  dialogVisible.value = true
}

async function handleSave() {
  if (!editing.value.name?.trim()) {
    ElMessage.warning('请填写品牌名称')
    return
  }
  saving.value = true
  try {
    const payload = {
      ...editing.value,
      website: editing.value.website?.trim() || '',
      agent: editing.value.agent?.trim() || '',
      authFiles: editing.value.authFiles || [],
    }
    if (editing.value.id) {
      await updateBrand(editing.value.id, payload)
    } else {
      await createBrand(payload)
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: Brand) {
  try {
    await deleteBrand(row.id)
    ElMessage.success('已删除')
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}

async function onLogoUpload(file: File) {
  try {
    editing.value.logo = await uploadImage(file, { scope: 'common', resource: 'brand_logo' })
    ElMessage.success('Logo 已上传')
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  }
  return false
}

function clearLogo() {
  editing.value.logo = ''
}

function websiteHref(url?: string) {
  const s = String(url || '').trim()
  if (!s) return ''
  if (/^https?:\/\//i.test(s)) return s
  return 'https://' + s
}

function isAuthImage(file: BrandAuthFile) {
  return /\.(png|jpe?g|jfif|webp|gif|svg)(\?|$)/i.test(file.url || file.name || '')
}

async function onAuthUpload(file: File) {
  uploadingAuth.value = true
  try {
    const url = await uploadImage(file, { scope: 'common', resource: 'brand_auth' })
    const next = [...(editing.value.authFiles || [])]
    next.push({ name: file.name, url })
    editing.value.authFiles = next
    ElMessage.success('授权文件已上传')
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    uploadingAuth.value = false
  }
  return false
}

function removeAuthFile(index: number) {
  editing.value.authFiles = (editing.value.authFiles || []).filter((_, i) => i !== index)
}
</script>

<template>
  <div class="brand-page">
    <el-card v-loading="loading">
      <template #header>
        <span>品牌管理</span>
        <el-button type="primary" :icon="Plus" @click="handleAdd">添加品牌</el-button>
      </template>

      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="搜索品牌"
          :prefix-icon="Search"
          clearable
          style="width: 240px"
          @change="loadData"
        />
      </div>

      <el-table :data="tableData" stripe border>
        <el-table-column label="品牌Logo" width="128" align="center">
          <template #default="{ row }">
            <img v-if="row.logo" :src="row.logo" :alt="row.name" class="brand-logo" :style="logoBgStyle(row.logoBg)" />
            <span v-else class="logo-placeholder">{{ row.firstLetter || row.name?.slice(0, 1) || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="品牌名称" min-width="140">
          <template #default="{ row }">
            <span>{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="品牌官网" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            <a v-if="row.website" :href="websiteHref(row.website)" target="_blank" rel="noopener noreferrer" class="site-link">
              {{ row.website }}
            </a>
            <span v-else class="muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="国内代理商" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <span>{{ row.agent || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="品牌授权" min-width="160">
          <template #default="{ row }">
            <div v-if="row.authFiles?.length" class="auth-cell">
              <a
                v-for="(file, i) in row.authFiles"
                :key="i"
                :href="file.url"
                target="_blank"
                rel="noopener noreferrer"
                class="auth-link"
              >
                {{ file.name || '附件' }}
              </a>
            </div>
            <span v-else class="muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="firstLetter" label="首字母" width="80" align="center" />
        <el-table-column prop="productCount" label="商品数" width="90" align="center" />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column label="显示" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.showStatus ? 'success' : 'info'" size="small">
              {{ row.showStatus ? '显示' : '隐藏' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="Edit" @click="handleEdit(row)">编辑</el-button>
            <el-popconfirm title="确定删除？" @confirm="handleDelete(row)">
              <template #reference>
                <el-button type="danger" link :icon="Delete">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑品牌' : '添加品牌'" width="620px">
      <el-form :model="editing" label-width="96px">
        <el-form-item label="品牌名称" required>
          <el-input v-model="editing.name" placeholder="请输入品牌名称" />
        </el-form-item>
        <el-form-item label="品牌Logo">
          <div class="logo-field">
            <img
              v-if="editing.logo"
              :src="editing.logo"
              alt="品牌Logo"
              class="logo-preview"
              :style="logoBgStyle(editing.logoBg)"
            />
            <el-input v-model="editing.logo" placeholder="Logo URL 或上传图片" />
            <el-upload :show-file-list="false" accept="image/*,.svg,image/svg+xml" :before-upload="onLogoUpload">
              <el-button :icon="Upload">上传</el-button>
            </el-upload>
            <el-button v-if="editing.logo" link type="danger" @click="clearLogo">清除</el-button>
          </div>
        </el-form-item>
        <el-form-item label="Logo背景">
          <div class="bg-field">
            <el-color-picker v-model="editing.logoBg" show-alpha />
            <el-input
              :model-value="editing.logoBg || ''"
              placeholder="#FFFFFF 或 rgba(...)，空为透明"
              @update:model-value="onLogoBgInput"
            />
            <el-button v-if="editing.logoBg" link type="danger" @click="editing.logoBg = ''">透明</el-button>
          </div>
        </el-form-item>
        <el-form-item label="品牌官网">
          <el-input v-model="editing.website" placeholder="如 https://www.example.com" :prefix-icon="Link" />
        </el-form-item>
        <el-form-item label="国内代理商">
          <el-input v-model="editing.agent" placeholder="国内代理商名称" />
        </el-form-item>
        <el-form-item label="品牌授权">
          <div class="auth-editor">
            <el-upload
              :show-file-list="false"
              accept="image/*,.svg,.pdf,application/pdf"
              multiple
              :before-upload="onAuthUpload"
            >
              <el-button :icon="Upload" :loading="uploadingAuth">上传图片 / PDF</el-button>
            </el-upload>
            <div v-if="editing.authFiles?.length" class="auth-list">
              <div v-for="(file, i) in editing.authFiles" :key="i" class="auth-item">
                <img v-if="isAuthImage(file)" :src="file.url" :alt="file.name" class="auth-thumb" />
                <el-icon v-else class="auth-doc"><Document /></el-icon>
                <a :href="file.url" target="_blank" rel="noopener noreferrer" class="auth-name">{{ file.name }}</a>
                <el-button link type="danger" @click="removeAuthFile(i)">删除</el-button>
              </div>
            </div>
            <p v-else class="auth-hint">可上传授权书图片或 PDF，支持多个附件</p>
          </div>
        </el-form-item>
        <el-form-item label="首字母">
          <el-input v-model="editing.firstLetter" maxlength="1" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="editing.sort" :min="0" />
        </el-form-item>
        <el-form-item label="是否显示">
          <el-switch v-model="editing.showStatus" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  margin-bottom: 16px;
}

.brand-logo,
.logo-preview {
  width: 96px;
  height: 48px;
  border-radius: 4px;
  flex-shrink: 0;
  background: #f7f7f7;
  border: 1px solid #f0f0f0;
  object-fit: contain;
}

.logo-placeholder {
  display: inline-flex;
  width: 96px;
  height: 48px;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
  border-radius: 4px;
  color: #909399;
  font-size: 16px;
  font-weight: 600;
}

.logo-field,
.bg-field {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.logo-field :deep(.el-input),
.bg-field :deep(.el-input) {
  flex: 1;
}

.muted {
  color: rgba(0, 0, 0, 0.35);
}

.site-link,
.auth-link,
.auth-name {
  color: #ff7700;
  text-decoration: none;
}

.site-link:hover,
.auth-link:hover,
.auth-name:hover {
  text-decoration: underline;
}

.auth-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.auth-editor {
  width: 100%;
}

.auth-hint {
  margin: 8px 0 0;
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
}

.auth-list {
  margin-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.auth-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  background: #fafafa;
  border-radius: 6px;
}

.auth-thumb {
  width: 40px;
  height: 40px;
  object-fit: contain;
  border-radius: 4px;
  background: #fff;
  border: 1px solid #eee;
  flex-shrink: 0;
}

.auth-doc {
  font-size: 22px;
  color: #909399;
  flex-shrink: 0;
}

.auth-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
}

:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
