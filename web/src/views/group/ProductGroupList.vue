<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Delete, View } from '@element-plus/icons-vue'
import {
  createGroup,
  deleteGroup,
  fetchGroupProducts,
  fetchGroups,
  updateGroup,
} from '../../api/product'
import type { Product, ProductGroup } from '../../types/product'

const tableData = ref<ProductGroup[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const editing = ref<Partial<ProductGroup>>({})
const popoverProducts = ref<Product[]>([])

async function loadData() {
  loading.value = true
  try {
    tableData.value = await fetchGroups()
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function handleAdd() {
  editing.value = { name: '', description: '', sort: 0 }
  dialogVisible.value = true
}

function handleEdit(row: ProductGroup) {
  editing.value = { ...row }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    if (editing.value.id) {
      await updateGroup(editing.value.id, editing.value)
    } else {
      await createGroup(editing.value)
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

async function handleDelete(row: ProductGroup) {
  try {
    await deleteGroup(row.id)
    ElMessage.success('已删除')
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}

async function loadGroupProducts(groupId: number) {
  try {
    popoverProducts.value = await fetchGroupProducts(groupId)
  } catch {
    popoverProducts.value = []
  }
}
</script>

<template>
  <div class="group-page">
    <el-card v-loading="loading">
      <template #header>
        <span>商品分组</span>
        <el-button type="primary" :icon="Plus" @click="handleAdd">新建分组</el-button>
      </template>

      <el-table :data="tableData" stripe border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="分组名称" min-width="140" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="productCount" label="商品数" width="90" align="center" />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="createTime" label="创建时间" width="160" />
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-popover placement="left" :width="320" trigger="click" @show="loadGroupProducts(row.id)">
              <template #reference>
                <el-button type="primary" link :icon="View">查看商品</el-button>
              </template>
              <div class="popover-products">
                <div v-for="p in popoverProducts" :key="p.id" class="pop-item">
                  <el-image :src="p.pic" class="pop-thumb" fit="cover" />
                  <span>{{ p.name }}</span>
                </div>
                <el-empty v-if="!popoverProducts.length" description="暂无商品" :image-size="60" />
              </div>
            </el-popover>
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

    <el-dialog v-model="dialogVisible" title="编辑分组" width="480px">
      <el-form :model="editing" label-width="80px">
        <el-form-item label="分组名称">
          <el-input v-model="editing.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editing.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="editing.sort" :min="0" />
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
.popover-products {
  max-height: 240px;
  overflow-y: auto;
}

.pop-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  border-bottom: 1px solid #f0f0f0;
  font-size: 13px;
}

.pop-thumb {
  width: 36px;
  height: 36px;
  border-radius: 4px;
  flex-shrink: 0;
}

:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
