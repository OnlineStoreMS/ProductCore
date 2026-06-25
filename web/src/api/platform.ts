import client, { unwrap, type PageData } from './client'
import type { ProductQuery } from './product'
import type { PlatformShop, PlatformShopQuery, PlatformShopType } from '../types/platform'
import type { Product } from '../types/product'

export async function fetchPlatformTypes(keyword?: string) {
  const res = await client.get('/platform-types', { params: { keyword } })
  return unwrap<PlatformShopType[]>(res)
}

export async function fetchEnabledPlatformTypes() {
  const res = await client.get('/platform-types/enabled')
  return unwrap<PlatformShopType[]>(res)
}

export async function createPlatformType(data: Partial<PlatformShopType>) {
  const res = await client.post('/platform-types', data)
  return unwrap<PlatformShopType>(res)
}

export async function updatePlatformType(id: number, data: Partial<PlatformShopType>) {
  const res = await client.put(`/platform-types/${id}`, data)
  return unwrap<PlatformShopType>(res)
}

export async function deletePlatformType(id: number) {
  await client.delete(`/platform-types/${id}`)
}

export async function fetchPlatformShops(query: PlatformShopQuery = {}) {
  const res = await client.get('/platform-shops', { params: query })
  return unwrap<PageData<PlatformShop>>(res)
}

export async function createPlatformShop(data: Partial<PlatformShop>) {
  const res = await client.post('/platform-shops', data)
  return unwrap<PlatformShop>(res)
}

export async function updatePlatformShop(id: number, data: Partial<PlatformShop>) {
  const res = await client.put(`/platform-shops/${id}`, data)
  return unwrap<PlatformShop>(res)
}

export async function deletePlatformShop(id: number) {
  await client.delete(`/platform-shops/${id}`)
}

export async function fetchShopListedProducts(shopId: number, query: ProductQuery = {}) {
  const res = await client.get(`/platform-shops/${shopId}/products`, { params: query })
  return unwrap<PageData<Product>>(res)
}
