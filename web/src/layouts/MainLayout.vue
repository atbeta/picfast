<template>
	<n-layout style="min-height: 100vh">
		<n-layout-header
			bordered
			style="display: flex; align-items: center; justify-content: space-between; padding: 0 24px; height: 56px"
		>
			<div style="display: flex; align-items: center; gap: 16px">
				<router-link to="/" style="font-size: 18px; font-weight: bold; text-decoration: none; color: #1f2937"
					>ImageAPI</router-link
				>
				<n-menu mode="horizontal" :value="currentRoute" :options="menuOptions" @update:value="onMenuSelect" />
			</div>
			<div style="display: flex; align-items: center; gap: 12px">
				<n-tag v-if="user?.role === 'admin'" type="info" size="small">管理员</n-tag>
				<n-dropdown :options="userMenuOptions" @select="onUserMenuSelect">
					<n-button quaternary>{{ user?.name || user?.email }}</n-button>
				</n-dropdown>
			</div>
		</n-layout-header>
		<n-layout-content style="padding: 24px">
			<router-view />
		</n-layout-content>
	</n-layout>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NLayout, NLayoutHeader, NLayoutContent, NMenu, NButton, NDropdown, NTag } from 'naive-ui'
import { useUserStore } from '../stores/user'
import { logout } from '../api/auth'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const user = computed(() => userStore.user)

const routeMap: Record<string, string> = {
	upload: '/',
	images: '/images',
	albums: '/albums',
	admin: '/admin',
}

onMounted(async () => {
	if (!user.value) {
		try {
			await userStore.fetchProfile()
		} catch {
			/* */
		}
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
		{ label: '上传', key: 'upload' },
		{ label: '图片', key: 'images' },
		{ label: '相册', key: 'albums' },
	]
	if (user.value?.role === 'admin') {
		opts.push({ label: '管理', key: 'admin' })
	}
	return opts
})

function onMenuSelect(key: string) {
	const path = routeMap[key]
	if (path) router.push(path)
}

const userMenuOptions = [
	{ label: '个人资料', key: 'profile' },
	{ label: '退出登录', key: 'logout' },
]

function onUserMenuSelect(key: string) {
	if (key === 'logout') {
		logout().catch(() => {})
		userStore.clearTokens()
		router.push('/login')
	}
}
</script>
