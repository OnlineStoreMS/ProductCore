import client, { unwrap, type PageData } from './client'

export interface DistributionShop {
  id: number
  platformTypeId: number
  platformTypeName?: string
  platformTypeLogo?: string
  name: string
  remark?: string
  itemCount: number
  createTime?: string
}

export interface DistributionItem {
  id: number
  shopId: number
  itemId: string
  title: string
  itemUrl: string
  picUrl: string
  price: string
  sales: string
  commentCount: string
  monthDeals: string
  monthConsign: string
  shipTime: string
  listedAt: string
  category: string
  tags: string
  sourceShopName: string
  sourceShopUrl: string
  collected: boolean
  productId?: number
}

export interface DistributionImportResult {
  created: number
  updated: number
  removed: number
  skipped: number
  total: number
  imageFailed: number
}

export async function fetchDistributionShops(params: {
  keyword?: string
  platformTypeId?: number
  page?: number
  pageSize?: number
}) {
  const res = await client.get('/distribution-shops', { params })
  return unwrap<PageData<DistributionShop>>(res)
}

export async function fetchDistributionShop(id: number) {
  const res = await client.get(`/distribution-shops/${id}`)
  return unwrap<DistributionShop>(res)
}

export async function createDistributionShop(data: { name: string; platformTypeId: number; remark?: string }) {
  const res = await client.post('/distribution-shops', data)
  return unwrap<DistributionShop>(res)
}

export async function updateDistributionShop(id: number, data: { name: string; platformTypeId: number; remark?: string }) {
  const res = await client.put(`/distribution-shops/${id}`, data)
  return unwrap<DistributionShop>(res)
}

export async function deleteDistributionShop(id: number) {
  await client.delete(`/distribution-shops/${id}`)
}

export async function fetchDistributionItems(shopId: number, params: {
  keyword?: string
  collected?: string
  sortBy?: string
  sortOrder?: string
  page?: number
  pageSize?: number
}) {
  const res = await client.get(`/distribution-shops/${shopId}/items`, { params })
  return unwrap<PageData<DistributionItem>>(res)
}

export interface DistributionConfigOption {
  value: string
  label: string
}

export interface DistributionConfigMeta {
  shipModes: DistributionConfigOption[]
  spotShipTimes: DistributionConfigOption[]
  afterSalePolicies: DistributionConfigOption[]
}

export interface DistributionConfig {
  id?: number
  platformTypeId: number
  platformTypeCode?: string
  platformTypeName?: string
  shipMode: string
  shipModeLabel?: string
  spotShipTime: string
  spotShipTimeLabel?: string
  freightTemplate: string
  afterSalePolicy: string
  afterSalePolicyLabel?: string
  enabled: number
  remark: string
}

export async function fetchDistributionConfigMeta() {
  const res = await client.get('/distribution-configs/meta')
  return unwrap<DistributionConfigMeta>(res)
}

export async function fetchDistributionConfigs() {
  const res = await client.get('/distribution-configs')
  return unwrap<DistributionConfig[]>(res)
}

export async function createDistributionConfig(data: DistributionConfig) {
  const res = await client.post('/distribution-configs', data)
  return unwrap<DistributionConfig>(res)
}

export async function updateDistributionConfig(id: number, data: DistributionConfig) {
  const res = await client.put(`/distribution-configs/${id}`, data)
  return unwrap<DistributionConfig>(res)
}

export async function deleteDistributionConfig(id: number) {
  await client.delete(`/distribution-configs/${id}`)
}

export async function importDistributionItems(shopId: number, file: File) {
  const form = new FormData()
  form.append('file', file)
  const res = await client.post(`/distribution-shops/${shopId}/items/import`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000,
  })
  return unwrap<DistributionImportResult>(res)
}
