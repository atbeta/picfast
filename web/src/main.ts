import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import { routes } from './router'
import { useUserStore } from './stores/user'
import './style.css'

const pinia = createPinia()
const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const token = localStorage.getItem('token')

  // 需要认证但无 token
  if (to.meta.requiresAuth && !token) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  // 需要管理员权限
  if (to.meta.requiresAdmin) {
    if (!token) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }
    const userStore = useUserStore()
    // 若用户信息未加载，尝试获取
    if (!userStore.user) {
      try {
        await userStore.fetchProfile()
      } catch {
        userStore.clearTokens()
        return { name: 'login', query: { redirect: to.fullPath } }
      }
    }
    if (!userStore.isAdmin()) {
      return { name: 'upload' }
    }
  }
})

const app = createApp(App)
app.use(pinia)
app.use(router)
app.mount('#app')
