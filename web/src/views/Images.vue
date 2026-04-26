<template>
	<div>
		<div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px">
			<n-h2 style="margin: 0">我的图片</n-h2>
			<div style="display: flex; align-items: center; gap: 12px">
				<n-text depth="3">共 {{ total }} 张</n-text>
				<n-button v-if="!batchMode" size="small" @click="batchMode = true">批量管理</n-button>
				<n-button v-else size="small" @click="exitBatch">退出管理</n-button>
			</div>
		</div>

		<!-- Batch toolbar -->
		<div
			v-if="batchMode"
			style="
				display: flex;
				align-items: center;
				justify-content: space-between;
				margin-bottom: 16px;
				padding: 12px;
				background: #f0f9ff;
				border-radius: 8px;
			"
		>
			<n-checkbox v-model:checked="allSelected" @update:checked="toggleSelectAll">
				全选 ({{ selectedKeys.length }} / {{ images.length }})
			</n-checkbox>
			<n-button
				type="error"
				size="small"
				:disabled="selectedKeys.length === 0"
				:loading="batchDeleting"
				@click="batchDelete"
			>
				批量删除
			</n-button>
		</div>

		<n-spin :show="loading">
			<n-grid :cols="4" :x-gap="12" :y-gap="12">
				<n-gi v-for="img in images" :key="img.key">
					<n-card size="small" hoverable :style="{ cursor: batchMode ? 'default' : 'pointer', position: 'relative' }">
						<template #cover>
							<div
								style="
									height: 160px;
									overflow: hidden;
									background: #f5f5f5;
									display: flex;
									align-items: center;
									justify-content: center;
								"
								@click="batchMode ? toggleSelect(img.key) : showDetail(img)"
							>
								<n-image
									:src="imgSrc(img)"
									width="100%"
									height="160"
									object-fit="cover"
									preview-disabled
									fallback-src=""
								>
									<template #error>
										<n-text depth="3" style="font-size: 12px">无预览</n-text>
									</template>
								</n-image>
								<!-- Batch checkbox overlay -->
								<div v-if="batchMode" style="position: absolute; top: 8px; left: 8px; z-index: 1">
									<n-checkbox :checked="selectedKeys.includes(img.key)" @update:checked="() => toggleSelect(img.key)" />
								</div>
							</div>
						</template>
						<div
							style="display: flex; align-items: center; justify-content: space-between"
							@click="batchMode ? toggleSelect(img.key) : showDetail(img)"
						>
							<n-text
								depth="3"
								style="
									font-size: 12px;
									overflow: hidden;
									text-overflow: ellipsis;
									white-space: nowrap;
									max-width: 120px;
								"
								>{{ img.origin_name }}</n-text
							>
							<n-tag :type="img.permission === 1 ? 'success' : 'warning'" size="tiny">
								{{ img.permission === 1 ? '公开' : '私有' }}
							</n-tag>
						</div>
						<div style="display: flex; align-items: center; justify-content: space-between">
							<n-text depth="3" style="font-size: 12px"
								>{{ formatSize(img.size_bytes) }} · {{ img.width || 0 }}x{{ img.height || 0 }}</n-text
							>
							<n-tag v-if="img.strategy_name" size="tiny" type="info">
								{{ img.strategy_name }}
							</n-tag>
						</div>
					</n-card>
				</n-gi>
			</n-grid>
			<n-empty v-if="!loading && images.length === 0" description="暂无图片，去上传吧" style="margin-top: 48px" />
		</n-spin>

		<div style="display: flex; justify-content: center; margin-top: 24px" v-if="total > pageSize">
			<n-pagination v-model:page="page" :page-count="Math.ceil(total / pageSize)" @update:page="fetchImages" />
		</div>

		<n-modal v-model:show="detailVisible" preset="card" title="图片详情" style="width: 600px">
			<template v-if="detail">
				<div
					style="
						display: flex;
						justify-content: center;
						margin-bottom: 16px;
						background: #f9f9f9;
						border-radius: 8px;
						padding: 8px;
					"
				>
					<n-image :src="detail.url" :alt="detail.key" style="max-height: 300px" />
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
						<n-tag v-if="detail.strategy_name" size="small" type="info">
							{{ detail.strategy_name }} ({{ detail.strategy_type === 'local' ? '本地' : 'S3' }})
						</n-tag>
						<n-text v-else depth="3">未知</n-text>
					</n-descriptions-item>
				</n-descriptions>
				<div style="margin-top: 16px" v-if="detail.links">
					<n-input-group v-for="(val, fmt) in detail.links" :key="fmt" style="margin-bottom: 8px">
						<n-input :value="val" size="small" readonly />
						<n-button size="small" @click="copyText(val)">{{ fmt }}</n-button>
					</n-input-group>
				</div>
				<div style="display: flex; justify-content: flex-end; margin-top: 16px">
					<n-button type="error" @click="deleteDetail">删除图片</n-button>
				</div>
			</template>
		</n-modal>
	</div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
	NH2,
	NGrid,
	NGi,
	NCard,
	NImage,
	NText,
	NTag,
	NEmpty,
	NPagination,
	NSpin,
	NModal,
	NDescriptions,
	NDescriptionsItem,
	NSwitch,
	NInputGroup,
	NInput,
	NButton,
	NCheckbox,
	useMessage,
	useDialog,
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

// Batch mode
const batchMode = ref(false)
const selectedKeys = ref<string[]>([])
const allSelected = ref(false)
const batchDeleting = ref(false)

function toggleSelect(key: string) {
	const idx = selectedKeys.value.indexOf(key)
	if (idx >= 0) {
		selectedKeys.value.splice(idx, 1)
	} else {
		selectedKeys.value.push(key)
	}
	allSelected.value = selectedKeys.value.length === images.value.length && images.value.length > 0
}

function toggleSelectAll(checked: boolean) {
	if (checked) {
		selectedKeys.value = images.value.map((img: any) => img.key)
	} else {
		selectedKeys.value = []
	}
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
			let success = 0
			let failed = 0
			for (const key of selectedKeys.value) {
				try {
					await deleteImage(key)
					success++
				} catch {
					failed++
				}
			}
			batchDeleting.value = false
			if (failed === 0) {
				message.success(`成功删除 ${success} 张图片`)
			} else {
				message.warning(`删除完成：成功 ${success} 张，失败 ${failed} 张`)
			}
			selectedKeys.value = []
			allSelected.value = false
			fetchImages()
		},
	})
}

onMounted(() => fetchImages())

function imgSrc(img: any) {
	return img.thumbnail_url || img.url || ''
}

async function fetchImages() {
	loading.value = true
	try {
		const res = await getImages(page.value, pageSize)
		const d = res.data.data
		images.value = d.items || d
		total.value = d.total || 0
	} catch {
		message.error('加载图片失败')
	} finally {
		loading.value = false
	}
}

async function showDetail(img: any) {
	try {
		const res = await getImage(img.key)
		detail.value = res.data.data
		detailVisible.value = true
	} catch {
		message.error('加载图片详情失败')
	}
}

async function togglePermission(public_: boolean) {
	if (!detail.value) return
	try {
		await updateImage(detail.value.key, { permission: public_ ? 1 : 0 })
		detail.value.permission = public_ ? 1 : 0
		message.success(public_ ? '已设为公开' : '已设为私有')
		fetchImages()
	} catch {
		message.error('修改权限失败')
	}
}

async function deleteDetail() {
	if (!detail.value) return
	dialog.warning({
		title: '确认删除',
		content: `确定要删除图片 "${detail.value.origin_name}" 吗？此操作不可撤销。`,
		positiveText: '删除',
		negativeText: '取消',
		onPositiveClick: async () => {
			try {
				await deleteImage(detail.value.key)
				message.success('图片已删除')
				detailVisible.value = false
				fetchImages()
			} catch {
				message.error('删除失败')
			}
		},
	})
}

function formatSize(bytes: number) {
	if (!bytes) return '0 B'
	const k = 1024
	const sizes = ['B', 'KB', 'MB', 'GB']
	const i = Math.floor(Math.log(bytes) / Math.log(k))
	return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}

function copyText(text: string) {
	navigator.clipboard
		.writeText(text)
		.then(() => message.success('已复制'))
		.catch(() => {})
}
</script>
