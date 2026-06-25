<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Delete, Search } from '@element-plus/icons-vue'
import ShopListedProductsDialog from '../../components/platform/ShopListedProductsDialog.vue'
import {
  createPlatformShop,
  deletePlatformShop,
  fetchEnabledPlatformTypes,
  fetchPlatformShops,
  updatePlatformShop,
} from '../../api/platform'
import type { PlatformShop, PlatformShopType } from '../../types/platform'

const loading = ref(false)
const tableData = ref<PlatformShop[]>([])
const total = ref(0)
const platformTypes = ref<PlatformShopType[]>([])

const query = ref({
  keyword: '',
  platformTypeId: undefined as number | undefined,
  page: 1,
  pageSize: 10,
})

const dialogVisible = ref(false)
const saving = ref(false)
const editing = ref<Partial<PlatformShop>>({})
const listedProductsVisible = ref(false)
const listedProductsShop = ref<PlatformShop | null>(null)

async function loadTypes() {
  platformTypes.value = await fetchEnabledPlatformTypes()
}

async function loadData() {
  loading.value = true
  try {
    const data = await fetchPlatformShops({
      keyword: query.value.keyword || undefined,
      platformTypeId: query.value.platformTypeId,
      page: query.value.page,
      pageSize: query.value.pageSize,
    })
    tableData.value = data.list
    total.value = data.total
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

watch(query, loadData, { deep: true })

onMounted(async () => {
  await loadTypes()
  await loadData()
})

function handleAdd() {
  editing.value = {
    platformTypeId: query.value.platformTypeId || platformTypes.value[0]?.id,
    name: '',
    shopCode: '',
    externalShopId: '',
    status: 1,
    sort: 0,
    remark: '',
  }
  dialogVisible.value = true
}

function handleEdit(row: PlatformShop) {
  editing.value = { ...row }
  dialogVisible.value = true
}

async function handleSave() {
  if (!editing.value.platformTypeId) {
    ElMessage.warning('请选择店铺类型')
    return
  }
  if (!editing.value.name?.trim()) {
    ElMessage.warning('请填写店铺名称')
    return
  }
  saving.value = true
  try {
    if (editing.value.id) {
      await updatePlatformShop(editing.value.id, editing.value)
    } else {
      await createPlatformShop(editing.value)
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    await loadData()
    await loadTypes()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: PlatformShop) {
  try {
    await deletePlatformShop(row.id)
    ElMessage.success('已删除')
    await loadData()
    await loadTypes()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}

function onPageChange(page: number) {
  query.value.page = page
}

function typeLabel(id: number) {
  return platformTypes.value.find((t) => t.id === id)?.name || `#${id}`
}

function handleViewProducts(row: PlatformShop) {
  listedProductsShop.value = row
  listedProductsVisible.value = true
}
</script>

<template>
  <div class="platform-shop-page">
    <el-card class="filter-card">
      <el-form :inline="true" :model="query">
        <el-form-item label="关键字">
          <el-input
            v-model="query.keyword"
            placeholder="店铺名称 / 编码 / 平台店铺ID"
            clearable
            :prefix-icon="Search"
            class="keyword-input"
          />
        </el-form-item>
        <el-form-item label="店铺类型">
          <el-select v-model="query.platformTypeId" placeholder="全部" clearable style="width: 160px">
            <el-option v-for="t in platformTypes" :key="t.id" :label="t.name" :value="t.id">
              <div class="type-option">
                <el-image v-if="t.logo" :src="t.logo" class="type-option-logo" fit="cover" />
                <span v-else class="type-option-logo type-option-logo--text">{{ t.name.slice(0, 1) }}</span>
                <span>{{ t.name }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-loading="loading">
      <template #header>
        <span>店铺管理 <el-tag size="small" type="info">{{ total }} 家</el-tag></span>
        <el-button type="primary" :icon="Plus" @click="handleAdd">添加店铺</el-button>
      </template>

      <el-table :data="tableData" stripe border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="平台类型" width="140">
          <template #default="{ row }">
            <div class="type-cell">
              <el-image
                v-if="row.platformTypeLogo"
                :src="row.platformTypeLogo"
                class="type-logo"
                fit="cover"
              />
              <span v-else class="type-logo type-logo--text">
                {{ row.platformTypeName?.slice(0, 1) || '?' }}
              </span>
              <span>{{ row.platformTypeName || typeLabel(row.platformTypeId) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="店铺名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="shopCode" label="内部编码" width="120" show-overflow-tooltip />
        <el-table-column prop="externalShopId" label="平台店铺ID" width="140" show-overflow-tooltip />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status ? 'success' : 'info'" size="small">
              {{ row.status ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="70" align="center" />
        <el-table-column label="已铺货" width="90" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleViewProducts(row)">
              {{ row.listedProductCount ?? 0 }} 件
            </el-button>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="Edit" @click="handleEdit(row)">编辑</el-button>
            <el-popconfirm title="确定删除该店铺？" @confirm="handleDelete(row)">
              <template #reference>
                <el-button type="danger" link :icon="Delete">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
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

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑店铺' : '添加店铺'" width="520px">
      <el-form :model="editing" label-width="96px">
        <el-form-item label="店铺类型" required>
          <el-select v-model="editing.platformTypeId" placeholder="请选择平台类型" style="width: 100%">
            <el-option v-for="t in platformTypes" :key="t.id" :label="t.name" :value="t.id">
              <div class="type-option">
                <el-image v-if="t.logo" :src="t.logo" class="type-option-logo" fit="cover" />
                <span v-else class="type-option-logo type-option-logo--text">{{ t.name.slice(0, 1) }}</span>
                <span>{{ t.name }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="店铺名称" required>
          <el-input v-model="editing.name" placeholder="如：XX旗舰店" />
        </el-form-item>
        <el-form-item label="内部编码">
          <el-input v-model="editing.shopCode" placeholder="系统内标识，可选" />
        </el-form-item>
        <el-form-item label="平台店铺ID">
          <el-input v-model="editing.externalShopId" placeholder="平台侧店铺 ID，可选" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="editing.sort" :min="0" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="editing.status" :active-value="1" :inactive-value="0" />
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

    <ShopListedProductsDialog v-model="listedProductsVisible" :shop="listedProductsShop" />
  </div>
</template>

<style scoped>
.platform-shop-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.filter-card :deep(.el-card__body) {
  padding-bottom: 2px;
}

.keyword-input {
  width: 320px;
}

.type-cell,
.type-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.type-logo,
.type-option-logo {
  width: 28px;
  height: 28px;
  border-radius: 4px;
  flex-shrink: 0;
}

.type-logo--text,
.type-option-logo--text {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
  color: #606266;
  font-size: 13px;
  font-weight: 600;
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
