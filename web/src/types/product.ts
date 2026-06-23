export interface Brand {
  id: number
  name: string
  logo?: string
  firstLetter: string
  sort: number
  showStatus: 0 | 1
  productCount: number
}

export interface Category {
  id: number
  parentId: number
  name: string
  level: number
  productCount: number
  sort: number
  showStatus: 0 | 1
  children?: Category[]
}

export interface ProductGroup {
  id: number
  name: string
  description: string
  productCount: number
  sort: number
  createTime: string
}

export interface SkuSpec {
  name: string
  values: string[]
}

export interface SkuItem {
  id?: number
  skuCode: string
  specs: Record<string, string>
  price: number
  costPrice: number
  stock: number
  pic?: string
}

export interface Product {
  id: number
  name: string
  subTitle: string
  productSn: string
  brandId: number
  brandName: string
  categoryId: number
  categoryName: string
  groupIds: number[]
  pic: string
  albumPics: string[]
  productVideo?: string
  price: number
  originalPrice: number
  stock: number
  unit: string
  weight: number
  publishStatus: 0 | 1
  verifyStatus: 0 | 1
  sort: number
  sale: number
  skuCount?: number
  description: string
  detailHtml: string
  skuSpecs: SkuSpec[]
  skus: SkuItem[]
  createTime: string
  updateTime: string
}

export type ProductForm = Omit<Product, 'id' | 'brandName' | 'categoryName' | 'createTime' | 'updateTime' | 'sale'> & {
  id?: number
}
