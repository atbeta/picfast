<template>
	<div>
		<h3 class="text-lg font-semibold text-[var(--color-text-primary)] mb-4">概览</h3>
		<div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
			<div class="bg-white rounded-lg border border-[var(--color-card-border)] p-5">
				<div class="text-xs font-medium text-[var(--color-text-tertiary)] uppercase tracking-wider mb-2">用户数</div>
				<div class="text-2xl font-bold text-[var(--color-text-primary)]">{{ stats.users }}</div>
			</div>
			<div class="bg-white rounded-lg border border-[var(--color-card-border)] p-5">
				<div class="text-xs font-medium text-[var(--color-text-tertiary)] uppercase tracking-wider mb-2">图片数</div>
				<div class="text-2xl font-bold text-[var(--color-text-primary)]">{{ stats.images }}</div>
			</div>
			<div class="bg-white rounded-lg border border-[var(--color-card-border)] p-5">
				<div class="text-xs font-medium text-[var(--color-text-tertiary)] uppercase tracking-wider mb-2">分组数</div>
				<div class="text-2xl font-bold text-[var(--color-text-primary)]">{{ stats.groups }}</div>
			</div>
			<div class="bg-white rounded-lg border border-[var(--color-card-border)] p-5">
				<div class="text-xs font-medium text-[var(--color-text-tertiary)] uppercase tracking-wider mb-2">存储策略</div>
				<div class="text-2xl font-bold text-[var(--color-text-primary)]">{{ stats.strategies }}</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminGetUsers, adminGetImages, adminGetGroups, adminGetStrategies } from '../../api/admin'

const stats = ref({ users: 0, images: 0, groups: 0, strategies: 0 })

onMounted(async () => {
	try {
		const [users, images, groups, strategies] = await Promise.all([
			adminGetUsers(), adminGetImages(), adminGetGroups(), adminGetStrategies(),
		])
		const ug = groups.data.data
		const sg = strategies.data.data
		stats.value = {
			users: users.data.data?.total || 0,
			images: images.data.data?.total || 0,
			groups: Array.isArray(ug) ? ug.length : 0,
			strategies: Array.isArray(sg) ? sg.length : 0,
		}
	} catch { /* */ }
})
</script>
