<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { fetchEnabledPlatformTypes } from '../../api/platform'
import {
  createDistributionShop,
  deleteDistributionShop,
  fetchDistributionShops,
  updateDistributionShop,
  type DistributionShop,
} from '../../api/distribution'
import type { PlatformShopType } from '../../types/platform'

const router = useRouter()
const loading = ref(false)
const saving = ref(false)
const tableData = ref<DistributionShop[]>([])
const total = ref(0)
const platformTypes = ref<PlatformShopType[]>([])
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const platformTypeId = ref<number | undefined>()

const dialogVisible = ref(false)
const editing = ref<{ id?: number; name: string; platformTypeId?: number; remark: string }>({
  name: '',
  remark: '',
})

function errorText(e: unknown, fallback: string) {
  const err = e as { response?: { data?: { message?: string } }; message?: string }
  return err.response?.data?.message || err.message || fallback
}

async function loadTypes() {
  platformTypes.value = await fetchEnabledPlatformTypes()
}

async function loadData() {
  loading.value = true
  try {
    const data = await fetchDistributionShops({
      keyword: keyword.value || undefined,
      platformTypeId: platformTypeId.value,
      page: page.value,
      pageSize: pageSize.value,
    })
    tableData.value = data.list
    total.value = data.total
  } catch (e) {
    ElMessage.error(errorText(e, '加载失败'))
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadTypes()
  await loadData()
})

function handleSearch() {
  page.value = 1
  loadData()
}

function handleAdd() {
  editing.value = {
    name: '',
    platformTypeId: platformTypeId.value || platformTypes.value[0]?.id,
    remark: '',
  }
  dialogVisible.value = true
}

function handleEdit(row: DistributionShop) {
  editing.value = {
    id: row.id,
    name: row.name,
    platformTypeId: row.platformTypeId,
    remark: row.remark || '',
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!editing.value.platformTypeId) {
    ElMessage.warning('请选择电商店铺类型')
    return
  }
  if (!editing.value.name.trim()) {
    ElMessage.warning('请填写店铺名称')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: editing.value.name.trim(),
      platformTypeId: editing.value.platformTypeId,
      remark: editing.value.remark.trim(),
    }
    if (editing.value.id) {
      await updateDistributionShop(editing.value.id, payload)
    } else {
      await createDistributionShop(payload)
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    await loadData()
  } catch (e) {
    ElMessage.error(errorText(e, '保存失败'))
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: DistributionShop) {
  try {
    await ElMessageBox.confirm(`确定删除铺货店铺「${row.name}」及其导入的商品？`, '删除店铺', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await deleteDistributionShop(row.id)
    ElMessage.success('已删除')
    await loadData()
  } catch (e) {
    ElMessage.error(errorText(e, '删除失败'))
  }
}

function openItems(row: DistributionShop) {
  router.push(`/distribution-shops/${row.id}`)
}
</script>

<template>
  <div class="page">
    <el-card class="filter-card">
      <el-form inline @submit.prevent="handleSearch">
        <el-form-item label="店铺名称">
          <el-input v-model="keyword" clearable placeholder="搜索店铺" :prefix-icon="Search" style="width: 220px" />
        </el-form-item>
        <el-form-item label="店铺类型">
          <el-select v-model="platformTypeId" clearable placeholder="全部" style="width: 160px">
            <el-option v-for="t in platformTypes" :key="t.id" :label="t.name" :value="t.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-loading="loading">
      <template #header>
        <span>铺货店铺 <el-tag size="small" type="info">{{ total }} 家</el-tag></span>
        <el-button type="primary" :icon="Plus" @click="handleAdd">新增铺货店铺</el-button>
      </template>
      <el-table :data="tableData" stripe border>
        <el-table-column prop="name" label="店铺名称" min-width="180" />
        <el-table-column label="电商店铺类型" width="160">
          <template #default="{ row }">{{ row.platformTypeName || '-' }}</template>
        </el-table-column>
        <el-table-column prop="itemCount" label="商品数" width="100" align="center" />
        <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openItems(row)">店铺商品</el-button>
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pager">
        <el-pagination
          background
          layout="total, prev, pager, next"
          :total="total"
          :page-size="pageSize"
          :current-page="page"
          @current-change="(p: number) => { page = p; loadData() }"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑铺货店铺' : '新增铺货店铺'" width="480px">
      <el-form label-width="110px">
        <el-form-item label="店铺名称" required>
          <el-input v-model="editing.name" maxlength="128" placeholder="请输入店铺名称" />
        </el-form-item>
        <el-form-item label="电商店铺类型" required>
          <el-select v-model="editing.platformTypeId" placeholder="请选择" style="width: 100%">
            <el-option v-for="t in platformTypes" :key="t.id" :label="t.name" :value="t.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editing.remark" type="textarea" :rows="2" maxlength="512" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.filter-card {
  margin-bottom: 16px;
}

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
