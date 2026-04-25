import type { RouteRecordRaw } from 'vue-router'

export const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('./views/auth/Login.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('./views/auth/Register.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    component: () => import('./layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', name: 'upload', component: () => import('./views/Upload.vue') },
      { path: 'images', name: 'images', component: () => import('./views/Images.vue') },
      { path: 'albums', name: 'albums', component: () => import('./views/Albums.vue') },
      {
        path: 'admin',
        component: () => import('./layouts/AdminLayout.vue'),
        meta: { requiresAdmin: true },
        children: [
          { path: '', name: 'admin-dashboard', component: () => import('./views/admin/Dashboard.vue') },
          { path: 'users', name: 'admin-users', component: () => import('./views/admin/Users.vue') },
          { path: 'groups', name: 'admin-groups', component: () => import('./views/admin/Groups.vue') },
          { path: 'strategies', name: 'admin-strategies', component: () => import('./views/admin/Strategies.vue') },
          { path: 'images', name: 'admin-images', component: () => import('./views/admin/Images.vue') },
          { path: 'settings', name: 'admin-settings', component: () => import('./views/admin/Settings.vue') },
        ],
      },
    ],
  },
]
