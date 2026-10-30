<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  fetchAIImageModels,
  fetchAISetting,
  imageModelKey,
  imageModelLabel,
  updateAISetting,
  type AIImageModel,
  type AISetting,
} from '../../api/ai'

const loading = ref(false)
const saving = ref(false)
const models = ref<AIImageModel[]>([])
const selected = ref('')
const setting = ref<AISetting>({
  defaultImageModel: 'qwen-image-3.0',
  defaultImageConnectionId: 0,
})

const options = computed(() =>
  models.value.map((item) => ({
    value: imageModelKey(item),
    label: imageModelLabel(item),
  })),
)

function errorText(e: unknown, fallback: string) {
  const err = e as { response?: { data?: { message?: string } }; message?: string }
  return err.response?.data?.message || err.message || fallback
}

async function load() {
  loading.value = true
  try {
    const [list, cfg] = await Promise.all([fetchAIImageModels(), fetchAISetting()])
    models.value = list || []
    setting.value = cfg
    const matched =
      models.value.find(
        (item) =>
          item.modelId === cfg.defaultImageModel &&
          (!cfg.defaultImageConnectionId || item.connectionId === cfg.defaultImageConnectionId),
      ) || models.value.find((item) => item.modelId === cfg.defaultImageModel)
    selected.value = matched ? imageModelKey(matched) : ''
  } catch (e) {
    ElMessage.error(errorText(e, '加载失败'))
  } finally {
    loading.value = false
  }
}

async function save() {
  const model = models.value.find((item) => imageModelKey(item) === selected.value)
  if (!model) {
    ElMessage.warning('请选择默认图片模型')
    return
  }
  saving.value = true
  try {
    setting.value = await updateAISetting({
      defaultImageModel: model.modelId,
      defaultImageConnectionId: model.connectionId,
    })
    ElMessage.success('已保存')
  } catch (e) {
    ElMessage.error(errorText(e, '保存失败'))
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="ai-setting" v-loading="loading">
    <el-card shadow="never">
      <template #header>
        <div class="card-head">
          <span>AI 修图</span>
        </div>
      </template>
      <p class="hint">
        默认使用 AI 大模型管理中心里已启用的生图 / 图片编辑模型。可配置千问、Grok 及其他厂商。商品修图弹窗也会先带出这个默认值，仍可临时换成别的模型。
      </p>
      <el-form label-width="120px" class="form">
        <el-form-item label="默认图片模型">
          <el-select
            v-model="selected"
            filterable
            placeholder="请选择图片模型"
            style="width: 420px"
            :disabled="!options.length"
          >
            <el-option v-for="item in options" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" :disabled="!selected" @click="save">保存</el-button>
          <el-button @click="load">刷新</el-button>
        </el-form-item>
      </el-form>
      <el-empty v-if="!loading && !models.length" description="还没有可用的图片模型，请先在 AI 大模型管理中心启用" />
    </el-card>
  </div>
</template>

<style scoped>
.ai-setting {
  max-width: 760px;
}

.card-head {
  font-weight: 600;
}

.hint {
  margin: 0 0 20px;
  color: #606266;
  line-height: 1.6;
}

.form {
  max-width: 560px;
}
</style>
