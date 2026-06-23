<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Delete } from '@element-plus/icons-vue'
import { createCategory, deleteCategory, fetchCategoryTree, updateCategory } from '../../api/product'
import type { Category } from '../../types/product'

const treeData = ref<Category[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const editing = ref<Partial<Category>>({})

async function loadData() {
  loading.value = true
  try {
    treeData.value = await fetchCategoryTree()
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function handleAdd(parent?: Category) {
  editing.value = { parentId: parent?.id ?? 0, name: '', sort: 0, showStatus: 1 }
  dialogVisible.value = true
}

function handleEdit(data: Category) {
  editing.value = { ...data }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    if (editing.value.id) {
      await updateCategory(editing.value.id, editing.value)
    } else {
      await createCategory(editing.value)
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

async function handleDelete(data: Category) {
  try {
    await deleteCategory(data.id)
    ElMessage.success('已删除')
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}
</script>

<template>
  <div class="category-page">
    <el-row :gutter="20">
      <el-col :span="10">
        <el-card v-loading="loading">
          <template #header>
            <span>分类树</span>
            <el-button type="primary" :icon="Plus" size="small" @click="handleAdd()">添加一级分类</el-button>
          </template>
          <el-tree
            :data="treeData"
            :props="{ label: 'name', children: 'children' }"
            default-expand-all
            node-key="id"
            highlight-current
          >
            <template #default="{ node, data }">
              <div class="tree-node">
                <span>{{ node.label }}</span>
                <span class="node-meta">
                  <el-tag size="small" type="info">{{ data.productCount }} 件</el-tag>
                  <el-button type="primary" link size="small" :icon="Plus" @click.stop="handleAdd(data)" />
                  <el-button type="primary" link size="small" :icon="Edit" @click.stop="handleEdit(data)" />
                  <el-popconfirm title="确定删除？" @confirm="handleDelete(data)">
                    <template #reference>
                      <el-button type="danger" link size="small" :icon="Delete" @click.stop />
                    </template>
                  </el-popconfirm>
                </span>
              </div>
            </template>
          </el-tree>
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card>
          <template #header>分类说明</template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="层级结构">支持多级分类，与电商类目结构对齐</el-descriptions-item>
            <el-descriptions-item label="平台映射">后期可配置各平台类目 ID 映射</el-descriptions-item>
            <el-descriptions-item label="属性模板">分类可绑定属性模板（规划中）</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="dialogVisible" title="编辑分类" width="420px">
      <el-form :model="editing" label-width="80px">
        <el-form-item label="分类名称">
          <el-input v-model="editing.name" />
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
.tree-node {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-right: 8px;
  font-size: 14px;
}

.node-meta {
  display: flex;
  align-items: center;
  gap: 4px;
}

:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
