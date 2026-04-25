<template>
  <div class="flex gap-6">
    <n-menu :value="currentRoute" :options="adminMenuOptions" style="width: 200px" />
    <div class="flex-1">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { NMenu } from 'naive-ui'

const route = useRoute()
const currentRoute = computed(() => {
  const path = route.path
  if (path.includes('/users')) return 'admin-users'
  if (path.includes('/groups')) return 'admin-groups'
  if (path.includes('/strategies')) return 'admin-strategies'
  if (path.includes('/images')) return 'admin-images'
  if (path.includes('/settings')) return 'admin-settings'
  return 'admin-dashboard'
})

const adminMenuOptions = [
  { label: 'Dashboard', key: 'admin-dashboard' },
  { label: 'Users', key: 'admin-users' },
  { label: 'Groups', key: 'admin-groups' },
  { label: 'Strategies', key: 'admin-strategies' },
  { label: 'Images', key: 'admin-images' },
  { label: 'Settings', key: 'admin-settings' },
]
</script>
