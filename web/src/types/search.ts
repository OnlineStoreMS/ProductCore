import type { ListedShop } from '../types/product'

export interface SuperSearchItem {
  productId: number
  productName: string
  materialCode: string
  productSn: string
  productPic: string
  brandName: string
  categoryName: string
  publishStatus: number
  skuId: number
  skuCode: string
  specs: Record<string, string>
  specLabel: string
  price: number
  stock: number
  pic: string
  listedShops: ListedShop[]
  listedShopCount: number
}
