<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Delete } from '@element-plus/icons-vue'
import {
  addKeywordProducts,
  createKeyword,
  deleteKeyword,
  fetchKeywordProducts,
  fetchKeywords,
  fetchProducts,
  removeKeywordProduct,
  updateKeyword,
} from '../../api/product'
import type { Product, ProductKeyword } from '../../types/product'

const list = ref<ProductKeyword[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const editing = ref<Partial<ProductKeyword>>({})

const productsDialogVisible = ref(false)
/** view: 查看列表；add: 添加商品 */
const productsDialogMode = ref<'view' | 'add'>('view')
const currentKeyword = ref<ProductKeyword | null>(null)
const keywordProducts = ref<Product[]>([])
const productsLoading = ref(false)
const addProductIds = ref<number[]>([])
const productOptions = ref<Product[]>([])
const productSearchLoading = ref(false)

async function loadData() {
  loading.value = true
  try {
    list.value = await fetchKeywords()
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

function handleEdit(row: ProductKeyword) {
  editing.value = { ...row }
  dialogVisible.value = true
}

async function handleSave() {
  try {
    if (editing.value.id) {
      await updateKeyword(editing.value.id, editing.value)
    } else {
      await createKeyword(editing.value)
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

async function handleDelete(row: ProductKeyword) {
  try {
    await deleteKeyword(row.id)
    ElMessage.success('已删除')
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}

async function openProducts(row: ProductKeyword, mode: 'view' | 'add' = 'view') {
  currentKeyword.value = row
  productsDialogMode.value = mode
  productsDialogVisible.value = true
  addProductIds.value = []
  productOptions.value = []
  await loadKeywordProducts()
  if (mode === 'add') {
    await searchProducts('')
  }
}

async function loadKeywordProducts() {
  if (!currentKeyword.value) return
  productsLoading.value = true
  try {
    keywordProducts.value = await fetchKeywordProducts(currentKeyword.value.id)
  } catch (e) {
    ElMessage.error((e as Error).message || '加载商品失败')
    keywordProducts.value = []
  } finally {
    productsLoading.value = false
  }
}

async function searchProducts(query: string) {
  productSearchLoading.value = true
  try {
    const data = await fetchProducts({ keyword: query || undefined, page: 1, pageSize: 30 })
    const linked = new Set(keywordProducts.value.map((p) => p.id))
    productOptions.value = data.list.filter((p) => !linked.has(p.id))
  } catch {
    productOptions.value = []
  } finally {
    productSearchLoading.value = false
  }
}

async function handleAddProducts() {
  if (!currentKeyword.value || !addProductIds.value.length) return
  try {
    await addKeywordProducts(currentKeyword.value.id, addProductIds.value)
    ElMessage.success('已添加')
    addProductIds.value = []
    await loadKeywordProducts()
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '添加失败')
  }
}

async function handleRemoveProduct(product: Product) {
  if (!currentKeyword.value) return
  try {
    await removeKeywordProduct(currentKeyword.value.id, product.id)
    ElMessage.success('已移除')
    await loadKeywordProducts()
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '移除失败')
  }
}
</script>

<template>
  <div class="keyword-page">
    <el-card v-loading="loading">
      <template #header>
        <span>关键词管理</span>
        <el-button type="primary" :icon="Plus" @click="handleAdd">新建关键词</el-button>
      </template>

      <el-table :data="list" stripe border>
        <el-table-column prop="name" label="关键词" min-width="160" />
        <el-table-column prop="description" label="描述" min-width="220" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="90" />
        <el-table-column label="商品" width="100">
          <template #default="{ row }">
            <el-button type="primary" link @click="openProducts(row, 'view')">
              {{ row.productCount || 0 }}
            </el-button>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="170" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="Plus" @click="openProducts(row, 'add')">添加商品</el-button>
            <el-button type="primary" link :icon="Edit" @click="handleEdit(row)">编辑</el-button>
            <el-popconfirm title="确定删除该关键词？" @confirm="handleDelete(row)">
              <template #reference>
                <el-button type="danger" link :icon="Delete">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑关键词' : '新建关键词'" width="480px">
      <el-form :model="editing" label-width="90px">
        <el-form-item label="关键词" required>
          <el-input v-model="editing.name" maxlength="64" show-word-limit />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editing.description" type="textarea" :rows="3" maxlength="512" />
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

    <el-dialog
      v-model="productsDialogVisible"
      :title="
        currentKeyword
          ? productsDialogMode === 'add'
            ? `添加商品 · ${currentKeyword.name}`
            : `商品列表 · ${currentKeyword.name}`
          : '商品列表'
      "
      width="720px"
      destroy-on-close
    >
      <div v-if="productsDialogMode === 'add'" class="add-row">
        <el-select
          v-model="addProductIds"
          multiple
          filterable
          remote
          clearable
          reserve-keyword
          placeholder="搜索并选择商品"
          :remote-method="searchProducts"
          :loading="productSearchLoading"
          style="flex: 1"
        >
          <el-option
            v-for="p in productOptions"
            :key="p.id"
            :label="`${p.name} (#${p.id})`"
            :value="p.id"
          />
        </el-select>
        <el-button type="primary" :disabled="!addProductIds.length" @click="handleAddProducts">添加</el-button>
      </div>

      <el-table v-loading="productsLoading" :data="keywordProducts" stripe border max-height="420">
        <el-table-column label="商品" min-width="280">
          <template #default="{ row }">
            <div class="product-cell">
              <el-image :src="row.pic" class="thumb" fit="cover" />
              <div>
                <div>{{ row.name }}</div>
                <div class="sub">#{{ row.id }} · {{ row.materialCode || row.productSn || '-' }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" align="center">
          <template #default="{ row }">
            <el-popconfirm title="从该关键词移除？" @confirm="handleRemoveProduct(row)">
              <template #reference>
                <el-button type="danger" link>移除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!productsLoading && !keywordProducts.length" description="暂无关联商品" :image-size="72" />
    </el-dialog>
  </div>
</template>

<style scoped>
:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.add-row {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.product-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.thumb {
  width: 40px;
  height: 40px;
  border-radius: 4px;
  flex-shrink: 0;
}

.sub {
  color: #94a3b8;
  font-size: 12px;
  margin-top: 2px;
}
</style>
