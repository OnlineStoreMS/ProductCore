import client, { unwrap, type PageData } from './client'
import type { Brand, Category, Product, ProductForm, ProductGroup } from '../types/product'

export interface ProductQuery {
  keyword?: string
  brandId?: number
  categoryId?: number
  groupId?: number
  publishStatus?: number
  page?: number
  pageSize?: number
}

export async function fetchProducts(query: ProductQuery = {}) {
  const res = await client.get('/products', { params: query })
  return unwrap<PageData<Product>>(res)
}

export async function fetchProduct(id: number) {
  const res = await client.get(`/products/${id}`)
  return unwrap<Product>(res)
}

export async function createProduct(data: ProductForm) {
  const res = await client.post('/products', data)
  return unwrap<Product>(res)
}

export async function updateProduct(id: number, data: ProductForm) {
  const res = await client.put(`/products/${id}`, data)
  return unwrap<Product>(res)
}

export async function deleteProduct(id: number) {
  await client.delete(`/products/${id}`)
}

export async function updateProductPublishStatus(id: number, publishStatus: 0 | 1) {
  await client.patch(`/products/${id}/publish-status`, { publishStatus })
}

export async function fetchBrands(keyword?: string) {
  const res = await client.get('/brands', { params: { keyword } })
  return unwrap<Brand[]>(res)
}

export async function createBrand(data: Partial<Brand>) {
  const res = await client.post('/brands', data)
  return unwrap<Brand>(res)
}

export async function updateBrand(id: number, data: Partial<Brand>) {
  const res = await client.put(`/brands/${id}`, data)
  return unwrap<Brand>(res)
}

export async function deleteBrand(id: number) {
  await client.delete(`/brands/${id}`)
}

export async function fetchCategoryTree() {
  const res = await client.get('/categories/tree')
  return unwrap<Category[]>(res)
}

export async function createCategory(data: Partial<Category>) {
  const res = await client.post('/categories', data)
  return unwrap<Category>(res)
}

export async function updateCategory(id: number, data: Partial<Category>) {
  const res = await client.put(`/categories/${id}`, data)
  return unwrap<Category>(res)
}

export async function deleteCategory(id: number) {
  await client.delete(`/categories/${id}`)
}

export async function fetchGroups() {
  const res = await client.get('/groups')
  return unwrap<ProductGroup[]>(res)
}

export async function fetchGroupProducts(groupId: number) {
  const res = await client.get(`/groups/${groupId}/products`)
  return unwrap<Product[]>(res)
}

export async function createGroup(data: Partial<ProductGroup>) {
  const res = await client.post('/groups', data)
  return unwrap<ProductGroup>(res)
}

export async function updateGroup(id: number, data: Partial<ProductGroup>) {
  const res = await client.put(`/groups/${id}`, data)
  return unwrap<ProductGroup>(res)
}

export async function deleteGroup(id: number) {
  await client.delete(`/groups/${id}`)
}

export async function fetchDashboardStats() {
  const [all, published, recent, brands, groups, skuList] = await Promise.all([
    fetchProducts({ page: 1, pageSize: 1 }),
    fetchProducts({ page: 1, pageSize: 1, publishStatus: 1 }),
    fetchProducts({ page: 1, pageSize: 5 }),
    fetchBrands(),
    fetchGroups(),
    fetchProducts({ page: 1, pageSize: 200 }),
  ])
  const skuTotal = skuList.list.reduce((s, p) => s + (p.skuCount || 0), 0)
  return {
    productTotal: all.total,
    publishedTotal: published.total,
    skuTotal,
    groupTotal: groups.length,
    recentProducts: recent.list,
    brands,
    groups,
  }
}
