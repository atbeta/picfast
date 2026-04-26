<template>
	<div style="display: flex; gap: 24px">
		<n-menu :value="currentRoute" :options="adminMenuOptions" @update:value="onMenuSelect" style="width: 200px" />
		<div style="flex: 1">
			<router-view />
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NMenu } from 'naive-ui'

const router = useRouter()
const route = useRoute()

const routeMap: Record<string, string> = {
	'admin-dashboard': '/admin',
	'admin-users': '/admin/users',
	'admin-groups': '/admin/groups',
	'admin-strategies': '/admin/strategies',
	'admin-images': '/admin/images',
	'admin-settings': '/admin/settings',
}

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
	{ label: '仪表盘', key: 'admin-dashboard' },
	{ label: '用户管理', key: 'admin-users' },
	{ label: '分组管理', key: 'admin-groups' },
	{ label: '存储策略', key: 'admin-strategies' },
	{ label: '图片管理', key: 'admin-images' },
	{ label: '系统设置', key: 'admin-settings' },
]

function onMenuSelect(key: string) {
	const path = routeMap[key]
	if (path) router.push(path)
}
</script>
