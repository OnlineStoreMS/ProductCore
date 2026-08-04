<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Delete, View } from '@element-plus/icons-vue'
import {
  createGroup,
  deleteGroup,
  fetchGroupProducts,
  fetchGroupTree,
  updateGroup,
} from '../../api/product'
import type { Product, ProductGroup } from '../../types/product'

const treeData = ref<ProductGroup[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const editing = ref<Partial<ProductGroup>>({})
const popoverProducts = ref<Product[]>([])

async function loadData() {
  loading.value = true
  try {
    treeData.value = await fetchGroupTree()
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function handleAdd(parent?: ProductGroup) {
  editing.value = {
    parentId: parent?.id ?? 0,
    name: '',
    description: '',
    sort: 0,
  }
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
        <el-button type="primary" :icon="Plus" @click="handleAdd()">新建一级分组</el-button>
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
            <span class="node-label">
              {{ node.label }}
              <span v-if="data.description" class="node-desc">{{ data.description }}</span>
            </span>
            <span class="node-meta">
              <el-tag size="small" type="info">{{ data.productCount }} 件</el-tag>
              <el-popover placement="left" :width="320" trigger="click" @show="loadGroupProducts(data.id)">
                <template #reference>
                  <el-button type="primary" link size="small" :icon="View" @click.stop>商品</el-button>
                </template>
                <div class="popover-products">
                  <div v-for="p in popoverProducts" :key="p.id" class="pop-item">
                    <el-image :src="p.pic" class="pop-thumb" fit="cover" />
                    <span>{{ p.name }}</span>
                  </div>
                  <el-empty v-if="!popoverProducts.length" description="暂无商品" :image-size="60" />
                </div>
              </el-popover>
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

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑分组' : '新建分组'" width="480px">
      <el-form :model="editing" label-width="90px">
        <el-form-item v-if="!editing.id && editing.parentId" label="上级分组">
          <el-tag type="info">子分组</el-tag>
        </el-form-item>
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
.tree-node {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-right: 8px;
  font-size: 14px;
  gap: 12px;
}

.node-label {
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
}

.node-desc {
  color: #94a3b8;
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 280px;
}

.node-meta {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

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

:deep(.el-tree-node__content) {
  height: 40px;
}
</style>
