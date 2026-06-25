import { createRouter, createWebHistory } from 'vue-router'
import AdminLayout from '../layouts/AdminLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: AdminLayout,
      redirect: '/dashboard',
      children: [
        { path: 'dashboard', name: 'Dashboard', component: () => import('../views/Dashboard.vue'), meta: { title: '工作台' } },
        { path: 'products', name: 'ProductList', component: () => import('../views/product/ProductList.vue'), meta: { title: '商品列表' } },
        { path: 'products/drafts', name: 'ProductDrafts', component: () => import('../views/product/ProductDraftList.vue'), meta: { title: '商品草稿箱' } },
        { path: 'products/trash', name: 'ProductTrash', component: () => import('../views/product/ProductTrashList.vue'), meta: { title: '商品回收站' } },
        { path: 'products/create', name: 'ProductCreate', component: () => import('../views/product/ProductEdit.vue'), meta: { title: '添加商品' } },
        { path: 'products/:id/edit', name: 'ProductEdit', component: () => import('../views/product/ProductEdit.vue'), meta: { title: '编辑商品' } },
        { path: 'categories', name: 'CategoryList', component: () => import('../views/category/CategoryList.vue'), meta: { title: '商品分类' } },
        { path: 'brands', name: 'BrandList', component: () => import('../views/brand/BrandList.vue'), meta: { title: '品牌管理' } },
        { path: 'groups', name: 'ProductGroupList', component: () => import('../views/group/ProductGroupList.vue'), meta: { title: '商品分组' } },
        { path: 'platform-types', name: 'PlatformTypeList', component: () => import('../views/platform/PlatformTypeList.vue'), meta: { title: '店铺类型' } },
        { path: 'platform-shops', name: 'PlatformShopList', component: () => import('../views/platform/PlatformShopList.vue'), meta: { title: '店铺管理' } },
      ],
    },
  ],
})

export default router
