<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Delete, Search, Upload } from '@element-plus/icons-vue'
import {
  createPlatformType,
  deletePlatformType,
  fetchPlatformTypes,
  updatePlatformType,
} from '../../api/platform'
import { uploadImage } from '../../api/upload'
import type { PlatformShopType } from '../../types/platform'

const tableData = ref<PlatformShopType[]>([])
const keyword = ref('')
const loading = ref(false)
const dialogVisible = ref(false)
const saving = ref(false)
const editing = ref<Partial<PlatformShopType>>({})

async function loadData() {
  loading.value = true
  try {
    tableData.value = await fetchPlatformTypes(keyword.value || undefined)
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function handleAdd() {
  editing.value = { code: '', name: '', logo: '', sort: 0, enabled: 1, remark: '' }
  dialogVisible.value = true
}

function handleEdit(row: PlatformShopType) {
  editing.value = { ...row }
  dialogVisible.value = true
}

async function handleSave() {
  if (!editing.value.name?.trim()) {
    ElMessage.warning('请填写类型名称')
    return
  }
  if (!editing.value.id && !editing.value.code?.trim()) {
    ElMessage.warning('请填写类型编码')
    return
  }
  saving.value = true
  try {
    if (editing.value.id) {
      await updatePlatformType(editing.value.id, editing.value)
    } else {
      await createPlatformType(editing.value)
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

async function handleDelete(row: PlatformShopType) {
  try {
    await deletePlatformType(row.id)
    ElMessage.success('已删除')
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}

async function onLogoUpload(file: File) {
  try {
    editing.value.logo = await uploadImage(file, { scope: 'common', resource: 'platform_logo' })
    ElMessage.success('Logo 已上传')
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  }
  return false
}
</script>

<template>
  <div class="platform-type-page">
    <el-card v-loading="loading">
      <template #header>
        <span>店铺类型</span>
        <el-button type="primary" :icon="Plus" @click="handleAdd">添加类型</el-button>
      </template>

      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="搜索类型名称 / 编码"
          :prefix-icon="Search"
          clearable
          class="keyword-input"
          @change="loadData"
        />
      </div>

      <el-table :data="tableData" stripe border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="Logo" width="72" align="center">
          <template #default="{ row }">
            <el-image v-if="row.logo" :src="row.logo" class="type-logo" fit="cover" />
            <span v-else class="logo-placeholder">{{ row.name?.slice(0, 1) || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="类型名称" min-width="120" />
        <el-table-column prop="code" label="编码" width="120" />
        <el-table-column prop="shopCount" label="店铺数" width="80" align="center" />
        <el-table-column prop="sort" label="排序" width="70" align="center" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
              {{ row.enabled ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="来源" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.isBuiltin" size="small" type="warning">内置</el-tag>
            <el-tag v-else size="small" type="info">自定义</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="Edit" @click="handleEdit(row)">编辑</el-button>
            <el-popconfirm
              v-if="!row.isBuiltin"
              title="确定删除该类型？"
              @confirm="handleDelete(row)"
            >
              <template #reference>
                <el-button type="danger" link :icon="Delete">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑店铺类型' : '添加店铺类型'" width="480px">
      <el-form :model="editing" label-width="88px">
        <el-form-item label="类型名称" required>
          <el-input v-model="editing.name" placeholder="如：京东、快手小店" />
        </el-form-item>
        <el-form-item label="类型编码" required>
          <el-input
            v-model="editing.code"
            placeholder="小写字母/数字/下划线，如 jd_shop"
            :disabled="!!editing.id && !!editing.isBuiltin"
          />
        </el-form-item>
        <el-form-item label="Logo">
          <div class="logo-field">
            <el-image v-if="editing.logo" :src="editing.logo" class="logo-preview" fit="cover" />
            <el-input v-model="editing.logo" placeholder="Logo URL 或上传图片" />
            <el-upload :show-file-list="false" accept="image/*" :before-upload="onLogoUpload">
              <el-button :icon="Upload">上传</el-button>
            </el-upload>
          </div>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="editing.sort" :min="0" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="editing.enabled" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editing.remark" type="textarea" :rows="2" />
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

.keyword-input {
  width: 280px;
}

.type-logo,
.logo-preview {
  width: 40px;
  height: 40px;
  border-radius: 6px;
}

.logo-placeholder {
  display: inline-flex;
  width: 40px;
  height: 40px;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
  border-radius: 6px;
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
