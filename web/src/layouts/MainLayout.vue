<template>
	<div class="flex h-screen bg-[var(--color-content-bg)]">
		<aside class="w-[220px] flex flex-col border-r border-[var(--color-sidebar-border)] bg-[var(--color-sidebar-bg)] shrink-0">
			<router-link to="/" class="flex items-center gap-2 px-5 h-14 border-b border-[var(--color-sidebar-border)]">
				<span class="text-lg font-bold text-[var(--color-text-primary)] tracking-tight">PicFast</span>
			</router-link>

			<nav class="flex-1 py-3 px-3 space-y-0.5 overflow-y-auto">
				<template v-for="item in menuItems" :key="item.key">
					<router-link
						:to="item.path"
						class="flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-colors"
						:class="isActive(item.key) ? 'bg-[var(--color-sidebar-active)] text-[var(--color-sidebar-text-active)] font-medium' : 'text-[var(--color-sidebar-text)] hover:bg-[var(--color-card-border)]'"
					>
						<NIcon size="18"><component :is="item.icon" /></NIcon>
						{{ item.label }}
					</router-link>
				</template>

				<div v-if="user?.role === 'admin'" class="pt-3 mt-3 border-t border-[var(--color-sidebar-border)]">
					<div class="px-3 pb-1.5 text-xs font-medium text-[var(--color-text-tertiary)] uppercase tracking-wider">管理</div>
					<router-link
						v-for="item in adminItems"
						:key="item.key"
						:to="item.path"
						class="flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-colors"
						:class="isActive(item.key) ? 'bg-[var(--color-sidebar-active)] text-[var(--color-sidebar-text-active)] font-medium' : 'text-[var(--color-sidebar-text)] hover:bg-[var(--color-card-border)]'"
					>
						<NIcon size="18"><component :is="item.icon" /></NIcon>
						{{ item.label }}
					</router-link>
				</div>
			</nav>
		</aside>

		<div class="flex-1 flex flex-col min-w-0">
			<header class="flex items-center justify-between h-14 px-6 border-b border-[var(--color-border)] bg-white shrink-0">
				<div class="text-sm text-[var(--color-text-secondary)]">
					{{ currentPageTitle }}
				</div>
				<div class="flex items-center gap-3">
					<n-tag v-if="user?.role === 'admin'" size="small" :bordered="false" type="info">管理员</n-tag>
					<n-dropdown :options="userMenuOptions" @select="onUserMenuSelect">
						<button class="flex items-center gap-2 text-sm text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] transition-colors">
							<div class="w-7 h-7 rounded-full bg-[var(--color-primary)] flex items-center justify-center text-white text-xs font-medium">
								{{ (user?.name || user?.email || '?')[0].toUpperCase() }}
							</div>
						</button>
					</n-dropdown>
				</div>
			</header>

			<main class="flex-1 overflow-y-auto p-6">
				<router-view />
			</main>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, onMounted, h } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NIcon, NTag, NDropdown } from 'naive-ui'
import {
	CloudUploadOutline,
	ImagesOutline,
	FolderOutline,
	KeyOutline,
	PersonOutline,
	SettingsOutline,
	PeopleOutline,
	ShieldOutline,
	ServerOutline,
	BarChartOutline,
	LogOutOutline,
} from '@vicons/ionicons5'
import { useUserStore } from '../stores/user'
import { logout } from '../api/auth'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const user = computed(() => userStore.user)

const menuItems = [
	{ key: 'upload', label: '上传', path: '/', icon: CloudUploadOutline },
	{ key: 'images', label: '图片', path: '/images', icon: ImagesOutline },
	{ key: 'albums', label: '相册', path: '/albums', icon: FolderOutline },
	{ key: 'api-tokens', label: '令牌', path: '/api-tokens', icon: KeyOutline },
]

const adminItems = [
	{ key: 'admin-dashboard', label: '概览', path: '/admin', icon: BarChartOutline },
	{ key: 'admin-users', label: '用户', path: '/admin/users', icon: PeopleOutline },
	{ key: 'admin-groups', label: '分组', path: '/admin/groups', icon: ShieldOutline },
	{ key: 'admin-strategies', label: '存储策略', path: '/admin/strategies', icon: ServerOutline },
	{ key: 'admin-images', label: '图片管理', path: '/admin/images', icon: ImagesOutline },
	{ key: 'admin-settings', label: '设置', path: '/admin/settings', icon: SettingsOutline },
]

const pageTitles: Record<string, string> = {
	'upload': '上传图片',
	'images': '我的图片',
	'albums': '相册',
	'api-tokens': 'API 令牌',
	'settings': '个人设置',
	'admin-dashboard': '管理概览',
	'admin-users': '用户管理',
	'admin-groups': '分组管理',
	'admin-strategies': '存储策略',
	'admin-images': '图片管理',
	'admin-settings': '系统设置',
}

const currentPageTitle = computed(() => {
	const key = route.meta.pageKey as string
	return pageTitles[key || route.name as string] || ''
})

function isActive(key: string) {
	if (key === 'upload') return route.path === '/'
	return route.path.startsWith('/' + key.replace('-', '/'))
}

onMounted(async () => {
	if (!user.value) {
		try { await userStore.fetchProfile() } catch { /* */ }
	}
})

const renderIcon = (icon: any) => () => h(NIcon, { size: 16 }, { default: () => h(icon) })

const userMenuOptions = [
	{ label: '个人设置', key: 'settings', icon: renderIcon(PersonOutline) },
	{ label: '退出登录', key: 'logout', icon: renderIcon(LogOutOutline) },
]

function onUserMenuSelect(key: string) {
	if (key === 'logout') {
		logout().catch(() => {})
		userStore.clearTokens()
		router.push('/login')
	} else if (key === 'settings') {
		router.push('/settings')
	}
}
</script>
