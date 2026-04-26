<template>
	<div>
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-lg font-semibold text-[var(--color-text-primary)]">我的图片</h2>
			<div class="flex items-center gap-3">
				<span class="text-sm text-[var(--color-text-tertiary)]">共 {{ total }} 张</span>
				<n-button v-if="!batchMode" size="small" quaternary @click="batchMode = true">批量管理</n-button>
				<n-button v-else size="small" quaternary @click="exitBatch">退出管理</n-button>
			</div>
		</div>

		<div v-if="batchMode" class="flex items-center justify-between mb-4 px-4 py-3 bg-blue-50/60 rounded-lg">
			<n-checkbox v-model:checked="allSelected" @update:checked="toggleSelectAll">
				全选 ({{ selectedKeys.length }} / {{ images.length }})
			</n-checkbox>
			<n-button type="error" size="small" :disabled="selectedKeys.length === 0" :loading="batchDeleting" @click="batchDelete">
				批量删除
			</n-button>
		</div>

		<n-spin :show="loading">
			<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
				<div v-for="img in images" :key="img.key"
					class="group bg-white rounded-lg border border-[var(--color-card-border)] overflow-hidden transition-shadow hover:shadow-md cursor-pointer"
					@click="batchMode ? toggleSelect(img.key) : showDetail(img)"
				>
					<div class="relative aspect-square bg-slate-50 flex items-center justify-center overflow-hidden">
						<img v-if="imgSrc(img)" :src="imgSrc(img)" class="w-full h-full object-cover" loading="lazy" />
						<span v-else class="text-xs text-[var(--color-text-tertiary)]">无预览</span>
						<div v-if="batchMode"
							class="absolute top-2 left-2 z-10"
							@click.stop
						>
							<n-checkbox :checked="selectedKeys.includes(img.key)" @update:checked="() => toggleSelect(img.key)" />
						</div>
						<div v-if="img.permission === 0"
							class="absolute top-2 right-2 bg-amber-500/80 text-white text-[10px] px-1.5 py-0.5 rounded"
						>私有</div>
					</div>
					<div class="px-2.5 py-2">
						<div class="flex items-center justify-between">
							<span class="text-xs text-[var(--color-text-secondary)] truncate max-w-[100px]">{{ img.origin_name }}</span>
						</div>
						<div class="flex items-center justify-between mt-0.5">
							<span class="text-[10px] text-[var(--color-text-tertiary)]">{{ formatSize(img.size_bytes) }}</span>
							<n-tag v-if="img.strategy_name" size="tiny" :bordered="false" type="info">{{ img.strategy_name }}</n-tag>
						</div>
					</div>
				</div>
			</div>
			<n-empty v-if="!loading && images.length === 0" description="暂无图片，去上传吧" class="mt-12" />
		</n-spin>

		<div class="flex justify-center mt-6" v-if="total > pageSize">
			<n-pagination v-model:page="page" :page-count="Math.ceil(total / pageSize)" @update:page="fetchImages" />
		</div>

		<n-modal v-model:show="detailVisible" preset="card" title="图片详情" class="max-w-xl">
			<template v-if="detail">
				<div class="flex justify-center mb-4 bg-slate-50 rounded-lg p-3">
					<img :src="toRelative(detail.url)" :alt="detail.key" class="max-h-72 object-contain" />
				</div>
				<n-descriptions bordered :column="2" label-placement="left" size="small">
					<n-descriptions-item label="Key">{{ detail.key }}</n-descriptions-item>
					<n-descriptions-item label="文件名">{{ detail.origin_name }}</n-descriptions-item>
					<n-descriptions-item label="大小">{{ formatSize(detail.size_bytes) }}</n-descriptions-item>
					<n-descriptions-item label="类型">{{ detail.mimetype }}</n-descriptions-item>
					<n-descriptions-item label="尺寸">{{ detail.width }}x{{ detail.height }}</n-descriptions-item>
					<n-descriptions-item label="权限">
						<n-switch :value="detail.permission === 1" @update:value="togglePermission">
							<template #checked>公开</template>
							<template #unchecked>私有</template>
						</n-switch>
					</n-descriptions-item>
					<n-descriptions-item label="存储策略">
						<n-tag v-if="detail.strategy_name" size="small" :bordered="false" type="info">
							{{ detail.strategy_name }} ({{ detail.strategy_type === 'local' ? '本地' : 'S3' }})
						</n-tag>
						<span v-else class="text-[var(--color-text-tertiary)]">未知</span>
					</n-descriptions-item>
				</n-descriptions>
				<div class="mt-4 space-y-2" v-if="detail.links">
					<n-input-group v-for="(val, fmt) in detail.links" :key="fmt">
						<n-input :value="val" size="small" readonly />
						<n-button size="small" @click="copyText(val)">{{ fmt }}</n-button>
					</n-input-group>
				</div>
				<div class="flex justify-end mt-4">
					<n-button type="error" size="small" quaternary @click="deleteDetail">删除图片</n-button>
				</div>
			</template>
		</n-modal>
	</div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
	NCard, NTag, NEmpty, NPagination, NSpin, NModal,
	NDescriptions, NDescriptionsItem, NSwitch, NInputGroup, NInput,
	NButton, NCheckbox, useMessage, useDialog,
} from 'naive-ui'
import { getImages, getImage, deleteImage, updateImage } from '../api/image'

const message = useMessage()
const dialog = useDialog()
const images = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = 20
const total = ref(0)
const detailVisible = ref(false)
const detail = ref<any>(null)

const batchMode = ref(false)
const selectedKeys = ref<string[]>([])
const allSelected = ref(false)
const batchDeleting = ref(false)

function toggleSelect(key: string) {
	const idx = selectedKeys.value.indexOf(key)
	if (idx >= 0) selectedKeys.value.splice(idx, 1)
	else selectedKeys.value.push(key)
	allSelected.value = selectedKeys.value.length === images.value.length && images.value.length > 0
}

function toggleSelectAll(checked: boolean) {
	selectedKeys.value = checked ? images.value.map((img: any) => img.key) : []
}

function exitBatch() {
	batchMode.value = false
	selectedKeys.value = []
	allSelected.value = false
}

async function batchDelete() {
	if (selectedKeys.value.length === 0) return
	const count = selectedKeys.value.length
	dialog.warning({
		title: '确认批量删除',
		content: `确定要删除选中的 ${count} 张图片吗？此操作不可撤销。`,
		positiveText: '删除',
		negativeText: '取消',
		onPositiveClick: async () => {
			batchDeleting.value = true
			let success = 0, failed = 0
			for (const key of selectedKeys.value) {
				try { await deleteImage(key); success++ } catch { failed++ }
			}
			batchDeleting.value = false
			if (failed === 0) message.success(`成功删除 ${success} 张图片`)
			else message.warning(`删除完成：成功 ${success} 张，失败 ${failed} 张`)
			selectedKeys.value = []
			allSelected.value = false
			fetchImages()
		},
	})
}

onMounted(() => fetchImages())

function imgSrc(img: any) {
	const raw = img.thumbnail_url || img.url || ''
	return toRelative(raw)
}

function toRelative(url: string) {
	try { return new URL(url).pathname }
	catch { return url }
}

async function fetchImages() {
	loading.value = true
	try {
		const res = await getImages(page.value, pageSize)
		const d = res.data.data
		images.value = d.items || d
		total.value = d.total || 0
	} catch { message.error('加载图片失败') }
	finally { loading.value = false }
}

async function showDetail(img: any) {
	try {
		const res = await getImage(img.key)
		detail.value = res.data.data
		detailVisible.value = true
	} catch { message.error('加载图片详情失败') }
}

async function togglePermission(public_: boolean) {
	if (!detail.value) return
	try {
		await updateImage(detail.value.key, { permission: public_ ? 1 : 0 })
		detail.value.permission = public_ ? 1 : 0
		message.success(public_ ? '已设为公开' : '已设为私有')
		fetchImages()
	} catch { message.error('修改权限失败') }
}

async function deleteDetail() {
	if (!detail.value) return
	dialog.warning({
		title: '确认删除',
		content: `确定要删除图片 "${detail.value.origin_name}" 吗？此操作不可撤销。`,
		positiveText: '删除', negativeText: '取消',
		onPositiveClick: async () => {
			try { await deleteImage(detail.value.key); message.success('图片已删除'); detailVisible.value = false; fetchImages() }
			catch { message.error('删除失败') }
		},
	})
}

function formatSize(bytes: number) {
	if (!bytes) return '0 B'
	const k = 1024, sizes = ['B', 'KB', 'MB', 'GB']
	const i = Math.floor(Math.log(bytes) / Math.log(k))
	return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}

function copyText(text: string) {
	navigator.clipboard.writeText(text).then(() => message.success('已复制')).catch(() => {})
}
</script>
