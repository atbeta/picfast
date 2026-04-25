<template>
  <n-layout style="min-height: 100vh">
    <n-layout-header bordered class="flex items-center justify-between px-6 py-3">
      <div class="flex items-center gap-4">
        <router-link to="/" class="text-lg font-bold no-underline text-gray-800">ImageAPI</router-link>
        <n-menu mode="horizontal" :value="currentRoute" :options="menuOptions" />
      </div>
      <div class="flex items-center gap-3">
        <n-tag v-if="user?.role === 'admin'" type="info" size="small">Admin</n-tag>
        <n-dropdown :options="userMenuOptions" @select="onUserMenuSelect">
          <n-button quaternary>{{ user?.name || user?.email }}</n-button>
        </n-dropdown>
      </div>
    </n-layout-header>
    <n-layout-content class="p-6">
      <router-view />
    </n-layout-content>
  </n-layout>
</template>

<script setup lang="ts">
import { computed, h, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NLayout, NLayoutHeader, NLayoutContent, NMenu, NButton, NDropdown, NTag, useMessage } from 'naive-ui'
import { useUserStore } from '../stores/user'
import { logout } from '../api/auth'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const message = useMessage()
const user = computed(() => userStore.user)

onMounted(async () => {
  if (!user.value) {
    try { await userStore.fetchProfile() } catch { /* */ }
  }
})

const currentRoute = computed(() => {
  const path = route.path
  if (path.startsWith('/images')) return 'images'
  if (path.startsWith('/albums')) return 'albums'
  if (path.startsWith('/admin')) return 'admin'
  return 'upload'
})

const menuOptions = computed(() => {
  const opts = [
    { label: 'Upload', key: 'upload' },
    { label: 'Images', key: 'images' },
    { label: 'Albums', key: 'albums' },
  ]
  if (user.value?.role === 'admin') {
    opts.push({ label: 'Admin', key: 'admin' })
  }
  return opts
})

const userMenuOptions = [
  { label: 'Profile', key: 'profile' },
  { label: 'Logout', key: 'logout' },
]

function onUserMenuSelect(key: string) {
  if (key === 'logout') {
    logout().catch(() => {})
    userStore.clearTokens()
    router.push('/login')
  }
}
</script>
