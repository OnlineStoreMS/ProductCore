<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Delete, Search, Upload } from '@element-plus/icons-vue'
import { createBrand, deleteBrand, fetchBrands, updateBrand } from '../../api/product'
import { uploadImage } from '../../api/upload'
import type { Brand } from '../../types/product'

const tableData = ref<Brand[]>([])
const keyword = ref('')
const loading = ref(false)
const dialogVisible = ref(false)
const saving = ref(false)
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
  editing.value = { name: '', logo: '', firstLetter: '', sort: 0, showStatus: 1 }
  dialogVisible.value = true
}

function handleEdit(row: Brand) {
  editing.value = { ...row }
  dialogVisible.value = true
}

async function handleSave() {
  if (!editing.value.name?.trim()) {
    ElMessage.warning('请填写品牌名称')
    return
  }
  saving.value = true
  try {
    if (editing.value.id) {
      await updateBrand(editing.value.id, editing.value)
    } else {
      await createBrand(editing.value)
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
            <img v-if="row.logo" :src="row.logo" :alt="row.name" class="brand-logo" />
            <span v-else class="logo-placeholder">{{ row.firstLetter || row.name?.slice(0, 1) || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="品牌名称" min-width="160">
          <template #default="{ row }">
            <span>{{ row.name }}</span>
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
        <el-table-column label="操作" width="140">
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

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑品牌' : '添加品牌'" width="480px">
      <el-form :model="editing" label-width="80px">
        <el-form-item label="品牌名称" required>
          <el-input v-model="editing.name" placeholder="请输入品牌名称" />
        </el-form-item>
        <el-form-item label="品牌Logo">
          <div class="logo-field">
            <img v-if="editing.logo" :src="editing.logo" alt="品牌Logo" class="logo-preview" />
            <el-input v-model="editing.logo" placeholder="Logo URL 或上传图片" />
            <el-upload :show-file-list="false" accept="image/*,.svg,image/svg+xml" :before-upload="onLogoUpload">
              <el-button :icon="Upload">上传</el-button>
            </el-upload>
            <el-button v-if="editing.logo" link type="danger" @click="clearLogo">清除</el-button>
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

.logo-field {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.logo-field :deep(.el-input) {
  flex: 1;
}

:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
