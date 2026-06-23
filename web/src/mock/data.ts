import type { Brand, Category, Product, ProductGroup } from '../types/product'

export const brands: Brand[] = [
  { id: 1, name: 'Apple', logo: '', firstLetter: 'A', sort: 1, showStatus: 1, productCount: 12 },
  { id: 2, name: 'Nike', logo: '', firstLetter: 'N', sort: 2, showStatus: 1, productCount: 28 },
  { id: 3, name: '优衣库', logo: '', firstLetter: 'Y', sort: 3, showStatus: 1, productCount: 45 },
  { id: 4, name: '小米', logo: '', firstLetter: 'X', sort: 4, showStatus: 1, productCount: 36 },
  { id: 5, name: '华为', logo: '', firstLetter: 'H', sort: 5, showStatus: 0, productCount: 18 },
]

export const categories: Category[] = [
  {
    id: 1, parentId: 0, name: '数码电器', level: 0, productCount: 120, sort: 1, showStatus: 1,
    children: [
      {
        id: 11, parentId: 1, name: '手机通讯', level: 1, productCount: 45, sort: 1, showStatus: 1,
        children: [
          { id: 111, parentId: 11, name: '智能手机', level: 2, productCount: 30, sort: 1, showStatus: 1 },
          { id: 112, parentId: 11, name: '手机配件', level: 2, productCount: 15, sort: 2, showStatus: 1 },
        ],
      },
      { id: 12, parentId: 1, name: '电脑办公', level: 1, productCount: 35, sort: 2, showStatus: 1 },
    ],
  },
  {
    id: 2, parentId: 0, name: '服装鞋包', level: 0, productCount: 280, sort: 2, showStatus: 1,
    children: [
      { id: 21, parentId: 2, name: '男装', level: 1, productCount: 80, sort: 1, showStatus: 1 },
      { id: 22, parentId: 2, name: '女装', level: 1, productCount: 120, sort: 2, showStatus: 1 },
      { id: 23, parentId: 2, name: '运动鞋', level: 1, productCount: 80, sort: 3, showStatus: 1 },
    ],
  },
  {
    id: 3, parentId: 0, name: '家居家装', level: 0, productCount: 95, sort: 3, showStatus: 1,
    children: [
      { id: 31, parentId: 3, name: '家纺', level: 1, productCount: 40, sort: 1, showStatus: 1 },
      { id: 32, parentId: 3, name: '灯具', level: 1, productCount: 55, sort: 2, showStatus: 1 },
    ],
  },
]

export const productGroups: ProductGroup[] = [
  { id: 1, name: '春季上新', description: '2026 春季主推商品', productCount: 24, sort: 1, createTime: '2026-03-01' },
  { id: 2, name: '爆款精选', description: '各平台高转化商品', productCount: 18, sort: 2, createTime: '2026-01-15' },
  { id: 3, name: '待上架抖店', description: '已完善信息，待同步抖店', productCount: 8, sort: 3, createTime: '2026-05-20' },
  { id: 4, name: '闲鱼专供', description: '适合闲鱼渠道的二手/尾货', productCount: 12, sort: 4, createTime: '2026-04-10' },
]

export const products: Product[] = [
  {
    id: 1001,
    name: 'Air Max 270 运动鞋',
    subTitle: '经典气垫跑鞋，舒适透气',
    productSn: 'NK-AM270-001',
    brandId: 2,
    brandName: 'Nike',
    categoryId: 23,
    categoryName: '运动鞋',
    groupIds: [1, 2],
    pic: 'https://picsum.photos/seed/nike1/200/200',
    albumPics: [
      'https://picsum.photos/seed/nike1/400/400',
      'https://picsum.photos/seed/nike2/400/400',
      'https://picsum.photos/seed/nike3/400/400',
    ],
    price: 899,
    originalPrice: 1299,
    stock: 320,
    unit: '双',
    weight: 850,
    publishStatus: 1,
    verifyStatus: 1,
    sort: 100,
    sale: 156,
    description: 'Nike Air Max 270 经典款',
    detailHtml: '<p>商品详情内容...</p>',
    skuSpecs: [
      { name: '颜色', values: ['黑白', '全黑', '白红'] },
      { name: '尺码', values: ['40', '41', '42', '43'] },
    ],
    skus: [
      { skuCode: 'NK270-BW-40', specs: { 颜色: '黑白', 尺码: '40' }, price: 899, costPrice: 450, stock: 30, pic: 'https://picsum.photos/seed/nike1/100/100' },
      { skuCode: 'NK270-BW-41', specs: { 颜色: '黑白', 尺码: '41' }, price: 899, costPrice: 450, stock: 45, pic: 'https://picsum.photos/seed/nike1/100/100' },
      { skuCode: 'NK270-BK-42', specs: { 颜色: '全黑', 尺码: '42' }, price: 899, costPrice: 450, stock: 50, pic: 'https://picsum.photos/seed/nike2/100/100' },
    ],
    createTime: '2026-01-10 14:30:00',
    updateTime: '2026-06-01 09:15:00',
  },
  {
    id: 1002,
    name: 'iPhone 16 Pro 256GB',
    subTitle: 'A18 Pro 芯片，钛金属设计',
    productSn: 'AP-IP16P-256',
    brandId: 1,
    brandName: 'Apple',
    categoryId: 111,
    categoryName: '智能手机',
    groupIds: [2],
    pic: 'https://picsum.photos/seed/iphone/200/200',
    albumPics: ['https://picsum.photos/seed/iphone/400/400'],
    price: 8999,
    originalPrice: 9999,
    stock: 88,
    unit: '台',
    weight: 199,
    publishStatus: 1,
    verifyStatus: 1,
    sort: 200,
    sale: 42,
    description: 'Apple iPhone 16 Pro',
    detailHtml: '<p>详情...</p>',
    skuSpecs: [
      { name: '颜色', values: ['原色钛金属', '白色钛金属', '黑色钛金属'] },
      { name: '存储', values: ['256GB', '512GB'] },
    ],
    skus: [
      { skuCode: 'IP16P-NT-256', specs: { 颜色: '原色钛金属', 存储: '256GB' }, price: 8999, costPrice: 7200, stock: 30 },
      { skuCode: 'IP16P-WT-512', specs: { 颜色: '白色钛金属', 存储: '512GB' }, price: 10999, costPrice: 8800, stock: 20 },
    ],
    createTime: '2026-02-20 10:00:00',
    updateTime: '2026-05-28 16:40:00',
  },
  {
    id: 1003,
    name: '纯棉圆领 T 恤',
    subTitle: '100% 纯棉，多色可选',
    productSn: 'UQ-T001',
    brandId: 3,
    brandName: '优衣库',
    categoryId: 21,
    categoryName: '男装',
    groupIds: [1, 4],
    pic: 'https://picsum.photos/seed/tshirt/200/200',
    albumPics: ['https://picsum.photos/seed/tshirt/400/400'],
    price: 79,
    originalPrice: 99,
    stock: 1200,
    unit: '件',
    weight: 180,
    publishStatus: 0,
    verifyStatus: 1,
    sort: 50,
    sale: 890,
    description: '基础款 T 恤',
    detailHtml: '<p>详情...</p>',
    skuSpecs: [
      { name: '颜色', values: ['白色', '黑色', '藏青'] },
      { name: '尺码', values: ['S', 'M', 'L', 'XL'] },
    ],
    skus: [],
    createTime: '2025-12-01 08:00:00',
    updateTime: '2026-06-10 11:20:00',
  },
]

export const platformStats = [
  { name: '淘宝', count: 156, color: '#FF5000' },
  { name: '抖音', count: 89, color: '#000000' },
  { name: '小红书', count: 45, color: '#FE2C55' },
  { name: '闲鱼', count: 32, color: '#FFDA44' },
]
