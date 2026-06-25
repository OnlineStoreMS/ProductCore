<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Edit, Plus, Search } from '@element-plus/icons-vue'
import { deleteProduct, discardEditDraft, fetchDraftProducts } from '../../api/product'
import type { Product } from '../../types/product'

const router = useRouter()
const loading = ref(false)
const tableData = ref<Product[]>([])
const total = ref(0)
const selectedRows = ref<Product[]>([])

const query = ref({
  keyword: '',
  page: 1,
  pageSize: 10,
})

async function loadData() {
  loading.value = true
  try {
    const data = await fetchDraftProducts({
      keyword: query.value.keyword || undefined,
      page: query.value.page,
      pageSize: query.value.pageSize,
    })
    tableData.value = data.list
    total.value = data.total
    selectedRows.value = []
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

watch(query, loadData, { deep: true })

onMounted(loadData)

function handleCreate() {
  router.push('/products/create')
}

function handleEdit(row: Product) {
  router.push(`/products/${row.id}/edit`)
}

function onSelectionChange(rows: Product[]) {
  selectedRows.value = rows
}

async function handleDelete(row: Product) {
  try {
    if (row.hasEditDraft && !row.isDraft) {
      await discardEditDraft(row.id)
      ElMessage.success('已放弃未发布编辑')
    } else {
      await deleteProduct(row.id)
      ElMessage.success('已移入回收站')
    }
    await loadData()
  } catch (e) {
    ElMessage.error((e as Error).message || '操作失败')
  }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selectedRows.value.length} 项？`, '批量操作', {
      type: 'warning',
    })
    let success = 0
    let failed = 0
    for (const row of selectedRows.value) {
      try {
        if (row.hasEditDraft && !row.isDraft) {
          await discardEditDraft(row.id)
        } else {
          await deleteProduct(row.id)
        }
        success++
      } catch {
        failed++
      }
    }
    if (success) ElMessage.success(`已处理 ${success} 项`)
    if (failed) ElMessage.warning(`${failed} 项失败`)
    await loadData()
  } catch {
    /* cancelled */
  }
}

function onPageChange(page: number) {
  query.value.page = page
}

const emptyTip = computed(() =>
  query.value.keyword ? '没有匹配的草稿' : '草稿箱为空，点击「新建草稿」开始创建',
)
</script>

<template>
  <div class="product-drafts">
    <el-card class="filter-card">
      <el-form :inline="true" :model="query">
        <el-form-item label="关键字">
          <el-input
            v-model="query.keyword"
            class="keyword-input"
            placeholder="商品名称 / 资料编码 / 货号 / 商品ID"
            clearable
            :prefix-icon="Search"
          />
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>商品草稿箱 <el-tag size="small" type="info">{{ total }} 条</el-tag></span>
          <el-button type="primary" :icon="Plus" @click="handleCreate">新建草稿</el-button>
        </div>
      </template>

      <el-alert
        type="info"
        :closable="false"
        show-icon
        class="draft-tip"
        title="包含新建草稿与已发布商品的未发布编辑；编辑中自动保存，完成后点击「保存并完成」。"
      />

      <div v-if="selectedRows.length" class="batch-bar">
        <span>已选 {{ selectedRows.length }} 项</span>
        <el-button type="danger" :icon="Delete" @click="handleBatchDelete">批量删除</el-button>
      </div>

      <el-table
        v-if="tableData.length"
        :data="tableData"
        stripe
        border
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column label="商品信息" min-width="260">
          <template #default="{ row }">
            <div class="product-info">
              <el-image
                v-if="row.pic"
                :src="row.pic"
                class="pic"
                fit="cover"
                :preview-src-list="[row.pic]"
                preview-teleported
                hide-on-click-modal
              />
              <div v-else class="pic pic-empty">无图</div>
              <div>
                <div class="title">
                  {{ row.name || '未命名商品' }}
                  <el-tag v-if="row.hasEditDraft && !row.isDraft" size="small" type="warning" class="type-tag">
                    未发布编辑
                  </el-tag>
                  <el-tag v-else size="small" type="info" class="type-tag">新建草稿</el-tag>
                </div>
                <div class="meta">商品ID：{{ row.id }}</div>
                <div v-if="row.materialCode" class="meta">资料编码：{{ row.materialCode }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="brandName" label="品牌" width="90">
          <template #default="{ row }">{{ row.brandName || '无品牌' }}</template>
        </el-table-column>
        <el-table-column prop="categoryName" label="分类" width="100">
          <template #default="{ row }">{{ row.categoryName || '无分类' }}</template>
        </el-table-column>
        <el-table-column label="草稿保存时间" width="168">
          <template #default="{ row }">
            {{ row.draftSavedAt || row.updateTime || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="Edit" @click="handleEdit(row)">继续编辑</el-button>
            <el-popconfirm
              :title="row.hasEditDraft && !row.isDraft ? '确定放弃未发布编辑？' : '确定删除该草稿？'"
              @confirm="handleDelete(row)"
            >
              <template #reference>
                <el-button type="danger" link :icon="Delete">
                  {{ row.hasEditDraft && !row.isDraft ? '放弃' : '删除' }}
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-else-if="!loading" :description="emptyTip" :image-size="80">
        <el-button type="primary" :icon="Plus" @click="handleCreate">新建草稿</el-button>
      </el-empty>

      <div v-if="total > query.pageSize" class="pagination">
        <el-pagination
          background
          layout="total, prev, pager, next"
          :total="total"
          :current-page="query.page"
          :page-size="query.pageSize"
          @current-change="onPageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.product-drafts {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.filter-card :deep(.el-card__body) {
  padding-bottom: 2px;
}

.keyword-input {
  width: 380px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.draft-tip {
  margin-bottom: 12px;
}

.batch-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
  font-size: 13px;
}

.product-info {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.pic {
  width: 64px;
  height: 64px;
  border-radius: 6px;
  flex-shrink: 0;
}

.pic-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  color: #909399;
  font-size: 12px;
}

.title {
  font-weight: 500;
  font-size: 14px;
  line-height: 1.4;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}

.type-tag {
  vertical-align: middle;
}

.meta {
  color: #909399;
  font-size: 12px;
  margin-top: 2px;
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
