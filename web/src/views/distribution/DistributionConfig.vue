<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { fetchEnabledPlatformTypes } from '../../api/platform'
import {
  createDistributionConfig,
  deleteDistributionConfig,
  fetchDistributionConfigMeta,
  fetchDistributionConfigs,
  updateDistributionConfig,
  type DistributionConfig,
  type DistributionConfigMeta,
} from '../../api/distribution'
import type { PlatformShopType } from '../../types/platform'

const loading = ref(false)
const saving = ref(false)
const rows = ref<DistributionConfig[]>([])
const platformTypes = ref<PlatformShopType[]>([])
const meta = ref<DistributionConfigMeta>({
  shipModes: [],
  spotShipTimes: [],
  afterSalePolicies: [],
})

const dialogVisible = ref(false)
const editing = ref<DistributionConfig>(emptyForm())

const usedTypeIds = computed(() => new Set(rows.value.map((row) => row.platformTypeId)))

function emptyForm(): DistributionConfig {
  return {
    platformTypeId: 0,
    shipMode: 'spot',
    spotShipTime: 'next_day',
    freightTemplate: '',
    afterSalePolicy: '7day',
    enabled: 1,
    remark: '',
  }
}

function errorText(e: unknown, fallback: string) {
  const err = e as { response?: { data?: { message?: string } }; message?: string }
  return err.response?.data?.message || err.message || fallback
}

async function loadData() {
  loading.value = true
  try {
    rows.value = await fetchDistributionConfigs()
  } catch (e) {
    ElMessage.error(errorText(e, '加载失败'))
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    const [types, options] = await Promise.all([
      fetchEnabledPlatformTypes(),
      fetchDistributionConfigMeta(),
    ])
    platformTypes.value = types
    meta.value = options
  } catch (e) {
    ElMessage.error(errorText(e, '加载选项失败'))
  }
  await loadData()
})

function handleAdd() {
  const first = platformTypes.value.find((t) => !usedTypeIds.value.has(t.id))
  editing.value = { ...emptyForm(), platformTypeId: first?.id || 0 }
  dialogVisible.value = true
}

function handleEdit(row: DistributionConfig) {
  editing.value = {
    id: row.id,
    platformTypeId: row.platformTypeId,
    shipMode: row.shipMode,
    spotShipTime: row.spotShipTime,
    freightTemplate: row.freightTemplate || '',
    afterSalePolicy: row.afterSalePolicy,
    enabled: row.enabled,
    remark: row.remark || '',
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!editing.value.platformTypeId) {
    ElMessage.warning('请选择电商平台')
    return
  }
  saving.value = true
  try {
    const payload = {
      ...editing.value,
      freightTemplate: editing.value.freightTemplate.trim(),
      remark: editing.value.remark.trim(),
      enabled: editing.value.enabled ? 1 : 0,
    }
    if (editing.value.id) {
      await updateDistributionConfig(editing.value.id, payload)
    } else {
      await createDistributionConfig(payload)
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

async function handleDelete(row: DistributionConfig) {
  if (!row.id) return
  try {
    await ElMessageBox.confirm(`确定删除「${row.platformTypeName || '该平台'}」的铺货配置？`, '删除配置', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await deleteDistributionConfig(row.id)
    ElMessage.success('已删除')
    await loadData()
  } catch (e) {
    ElMessage.error(errorText(e, '删除失败'))
  }
}

function typeDisabled(id: number) {
  if (editing.value.id && editing.value.platformTypeId === id) return false
  return usedTypeIds.value.has(id)
}
</script>

<template>
  <div class="page">
    <el-alert
      class="tip"
      type="info"
      show-icon
      :closable="false"
      title="按电商平台保存发货模式、现货发货时间、运费模板和售后政策。同一平台只保留一条。当前商品铺货仍使用电商店铺页面上已经选好的默认值。"
    />

    <el-card v-loading="loading">
      <template #header>
        <span>铺货配置</span>
        <el-button type="primary" :icon="Plus" @click="handleAdd">新增配置</el-button>
      </template>
      <el-table :data="rows" stripe border>
        <el-table-column label="电商平台" width="140">
          <template #default="{ row }">{{ row.platformTypeName || '-' }}</template>
        </el-table-column>
        <el-table-column prop="shipModeLabel" label="发货模式" width="120" />
        <el-table-column prop="spotShipTimeLabel" label="现货发货时间" width="140" />
        <el-table-column label="运费模板" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.freightTemplate || '-' }}</template>
        </el-table-column>
        <el-table-column prop="afterSalePolicyLabel" label="售后政策" min-width="200" show-overflow-tooltip />
        <el-table-column label="启用" width="80" align="center">
          <template #default="{ row }">{{ row.enabled ? '是' : '否' }}</template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑铺货配置' : '新增铺货配置'" width="520px">
      <el-form label-width="120px">
        <el-form-item label="电商平台" required>
          <el-select v-model="editing.platformTypeId" placeholder="请选择" style="width: 100%">
            <el-option
              v-for="t in platformTypes"
              :key="t.id"
              :label="t.name"
              :value="t.id"
              :disabled="typeDisabled(t.id)"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="发货模式" required>
          <el-select v-model="editing.shipMode" style="width: 100%">
            <el-option v-for="opt in meta.shipModes" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="现货发货时间" required>
          <el-select v-model="editing.spotShipTime" style="width: 100%">
            <el-option v-for="opt in meta.spotShipTimes" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="运费模板">
          <el-input
            v-model="editing.freightTemplate"
            maxlength="128"
            placeholder="填写该平台运费模板名称，例如常用运费模板（杭州）"
          />
        </el-form-item>
        <el-form-item label="售后政策" required>
          <el-select v-model="editing.afterSalePolicy" style="width: 100%">
            <el-option v-for="opt in meta.afterSalePolicies" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="editing.enabled" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editing.remark" type="textarea" :rows="2" maxlength="256" />
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
.tip {
  margin-bottom: 16px;
}

:deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
