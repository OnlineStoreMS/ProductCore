<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SkuSpecEditor from '../../components/product/SkuSpecEditor.vue'
import {
  createProduct,
  fetchBrands,
  fetchCategoryTree,
  fetchGroups,
  fetchProduct,
  updateProduct,
} from '../../api/product'
import type { Category, ProductForm, SkuItem, SkuSpec } from '../../types/product'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const saving = ref(false)
const isEdit = computed(() => !!route.params.id)
const activeTab = ref('basic')

const brands = ref<{ id: number; name: string }[]>([])
const categories = ref<Category[]>([])
const productGroups = ref<{ id: number; name: string }[]>([])

const form = ref<ProductForm>({
  name: '',
  subTitle: '',
  productSn: '',
  brandId: undefined as unknown as number,
  categoryId: undefined as unknown as number,
  groupIds: [],
  pic: '',
  albumPics: [],
  price: 0,
  originalPrice: 0,
  stock: 0,
  unit: '件',
  weight: 0,
  publishStatus: 0,
  verifyStatus: 1,
  sort: 0,
  description: '',
  detailHtml: '',
  skuSpecs: [],
  skus: [],
})

const skuSpecs = ref<SkuSpec[]>([])
const skus = ref<SkuItem[]>([])

const categoryOptions = computed(() => {
  const flat: { id: number; name: string }[] = []
  function walk(list: Category[], prefix = '') {
    for (const c of list) {
      flat.push({ id: c.id, name: prefix + c.name })
      if (c.children?.length) walk(c.children, prefix + c.name + ' / ')
    }
  }
  walk(categories.value)
  return flat
})

async function loadMeta() {
  const [b, c, g] = await Promise.all([fetchBrands(), fetchCategoryTree(), fetchGroups()])
  brands.value = b
  categories.value = c
  productGroups.value = g
}

async function loadProduct() {
  if (!isEdit.value) return
  loading.value = true
  try {
    const id = Number(route.params.id)
    const product = await fetchProduct(id)
    form.value = {
      name: product.name,
      subTitle: product.subTitle,
      productSn: product.productSn,
      brandId: product.brandId,
      categoryId: product.categoryId,
      groupIds: product.groupIds || [],
      pic: product.pic,
      albumPics: product.albumPics || [],
      productVideo: product.productVideo,
      price: product.price,
      originalPrice: product.originalPrice,
      stock: product.stock,
      unit: product.unit,
      weight: product.weight,
      publishStatus: product.publishStatus,
      verifyStatus: product.verifyStatus,
      sort: product.sort,
      description: product.description,
      detailHtml: product.detailHtml,
      skuSpecs: product.skuSpecs || [],
      skus: product.skus || [],
    }
    skuSpecs.value = JSON.parse(JSON.stringify(product.skuSpecs || []))
    skus.value = JSON.parse(JSON.stringify(product.skus || []))
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadMeta()
  await loadProduct()
})

async function handleSave() {
  if (!form.value.name) {
    ElMessage.warning('请填写商品名称')
    activeTab.value = 'basic'
    return
  }
  form.value.skuSpecs = skuSpecs.value
  form.value.skus = skus.value
  if (skus.value.length) {
    const prices = skus.value.map((s) => s.price).filter(Boolean)
    form.value.price = prices.length ? Math.min(...prices) : form.value.price
    form.value.stock = skus.value.reduce((sum, s) => sum + (s.stock || 0), 0)
  }
  saving.value = true
  try {
    if (isEdit.value) {
      await updateProduct(Number(route.params.id), form.value)
      ElMessage.success('商品已更新')
    } else {
      await createProduct(form.value)
      ElMessage.success('商品已创建')
    }
    router.push('/products')
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  } finally {
    saving.value = false
  }
}

function addAlbumPic() {
  const seed = Date.now()
  form.value.albumPics.push(`https://picsum.photos/seed/${seed}/400/400`)
}
</script>

<template>
  <div v-loading="loading" class="product-edit">
    <el-card>
      <template #header>
        <span>{{ isEdit ? '编辑商品' : '添加商品' }}</span>
        <div>
          <el-button @click="router.back()">取消</el-button>
          <el-button type="primary" :loading="saving" @click="handleSave">保存商品</el-button>
        </div>
      </template>

      <el-tabs v-model="activeTab" class="edit-tabs">
        <el-tab-pane label="基本信息" name="basic">
          <el-form :model="form" label-width="100px" class="edit-form">
            <el-row :gutter="24">
              <el-col :span="12">
                <el-form-item label="商品名称" required>
                  <el-input v-model="form.name" placeholder="请输入商品名称" maxlength="100" show-word-limit />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="副标题">
                  <el-input v-model="form.subTitle" placeholder="卖点 / 促销语" />
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="货号">
                  <el-input v-model="form.productSn" placeholder="内部货号" />
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="品牌">
                  <el-select v-model="form.brandId" placeholder="选择品牌" style="width: 100%">
                    <el-option v-for="b in brands" :key="b.id" :label="b.name" :value="b.id" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="分类">
                  <el-select v-model="form.categoryId" placeholder="选择分类" style="width: 100%">
                    <el-option v-for="c in categoryOptions" :key="c.id" :label="c.name" :value="c.id" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="商品分组">
                  <el-select v-model="form.groupIds" multiple placeholder="选择分组" style="width: 100%">
                    <el-option v-for="g in productGroups" :key="g.id" :label="g.name" :value="g.id" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item label="单位">
                  <el-input v-model="form.unit" placeholder="件/双/台" />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item label="重量(g)">
                  <el-input-number v-model="form.weight" :min="0" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item label="市场价">
                  <el-input-number v-model="form.originalPrice" :min="0" :precision="2" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item label="排序">
                  <el-input-number v-model="form.sort" :min="0" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="24">
                <el-form-item label="商品描述">
                  <el-input v-model="form.description" type="textarea" :rows="3" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="上架状态">
                  <el-switch v-model="form.publishStatus" :active-value="1" :inactive-value="0" active-text="上架" inactive-text="下架" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="规格与 SKU" name="sku">
          <SkuSpecEditor v-model="skuSpecs" v-model:skus="skus" />
        </el-tab-pane>

        <el-tab-pane label="商品图片" name="media">
          <el-form label-width="100px">
            <el-form-item label="商品主图">
              <el-input v-model="form.pic" placeholder="图片 URL" style="max-width: 480px" />
              <el-image v-if="form.pic" :src="form.pic" style="width: 120px; height: 120px; margin-top: 8px" fit="cover" />
            </el-form-item>
            <el-form-item label="商品相册">
              <div class="album-grid">
                <div v-for="(pic, i) in form.albumPics" :key="i" class="album-item">
                  <el-image :src="pic" fit="cover" class="album-pic" />
                  <el-button size="small" type="danger" circle class="remove-btn" @click="form.albumPics.splice(i, 1)">×</el-button>
                </div>
                <div class="album-add" @click="addAlbumPic">
                  <el-icon><Plus /></el-icon>
                  <span>添加</span>
                </div>
              </div>
            </el-form-item>
            <el-form-item label="商品视频">
              <el-input v-model="form.productVideo" placeholder="视频 URL" style="max-width: 400px" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="商品详情" name="detail">
          <el-form label-width="100px">
            <el-form-item label="详情 HTML">
              <el-input v-model="form.detailHtml" type="textarea" :rows="12" placeholder="商品详情 HTML" />
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<style scoped>
.product-edit :deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.edit-tabs {
  min-height: 480px;
}

.edit-form {
  max-width: 960px;
  padding-top: 8px;
}

.album-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.album-item {
  position: relative;
  width: 100px;
  height: 100px;
}

.album-pic {
  width: 100px;
  height: 100px;
  border-radius: 6px;
}

.remove-btn {
  position: absolute;
  top: -8px;
  right: -8px;
}

.album-add {
  width: 100px;
  height: 100px;
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #8c939d;
  gap: 4px;
  font-size: 12px;
}

.album-add:hover {
  border-color: #409eff;
  color: #409eff;
}
</style>
