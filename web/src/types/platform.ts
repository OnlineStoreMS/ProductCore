export interface PlatformShopType {
  id: number
  code: string
  name: string
  logo?: string
  sort: number
  enabled: 0 | 1
  isBuiltin?: 0 | 1
  remark?: string
  shopCount?: number
  createTime?: string
}

export interface PlatformShop {
  id: number
  platformTypeId: number
  platformTypeName?: string
  platformTypeLogo?: string
  name: string
  sourceChannel?: string
  shopCode?: string
  externalShopId?: string
  status: 0 | 1
  remark?: string
  sort: number
  listedProductCount?: number
  createTime?: string
  updateTime?: string
}

export interface PlatformShopQuery {
  keyword?: string
  platformTypeId?: number
  status?: number
  page?: number
  pageSize?: number
}
