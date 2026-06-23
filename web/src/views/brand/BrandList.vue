<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Delete, Search } from '@element-plus/icons-vue'
import { createBrand, deleteBrand, fetchBrands, updateBrand } from '../../api/product'
import type { Brand } from '../../types/product'

const tableData = ref<Brand[]>([])
const keyword = ref('')
const loading = ref(false)
const dialogVisible = ref(false)
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
  editing.value = { name: '', firstLetter: '', sort: 0, showStatus: 1 }
  dialogVisible.value = true
}

function handleEdit(row: Brand) {
  editing.value = { ...row }
  dialogVisible.value = true
}

async function handleSave() {
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
</script>

<template>
  <div class="brand-page">
    <el-card v-loading="loading">
      <template #header>
        <span>品牌管理</span>
        <el-button type="primary" :icon="Plus" @click="handleAdd">添加品牌</el-button>
      </template>

      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索品牌" :prefix-icon="Search" clearable style="width: 240px" @change="loadData" />
      </div>

      <el-table :data="tableData" stripe border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="品牌名称" min-width="160">
          <template #default="{ row }">
            <div class="brand-cell">
              <el-avatar :size="36" shape="square">{{ row.firstLetter }}</el-avatar>
              <span>{{ row.name }}</span>
            </div>
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

    <el-dialog v-model="dialogVisible" title="编辑品牌" width="420px">
      <el-form :model="editing" label-width="80px">
        <el-form-item label="品牌名称">
          <el-input v-model="editing.name" />
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
        <el-button type="primary" @click="handleSave">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  margin-bottom: 16px;
}

.brand-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
