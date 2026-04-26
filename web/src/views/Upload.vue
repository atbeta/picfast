<template>
	<div>
		<div class="mb-4">
			<h2 class="text-lg font-semibold text-[var(--color-text-primary)]">上传图片</h2>
		</div>

		<div class="mb-4 space-y-3">
			<div v-if="strategies.length > 1" class="flex items-center gap-3 text-sm text-[var(--color-text-secondary)]">
				<span>存储策略：</span>
				<n-select
					v-model:value="selectedStrategyId"
					:options="strategyOptions"
					class="w-48"
					size="small"
					@update:value="onStrategyChange"
				/>
			</div>
			<div v-else-if="strategies.length === 1" class="text-sm text-[var(--color-text-secondary)]">
				当前使用存储策略：<span class="font-medium text-[var(--color-text-primary)]">{{ strategies[0].name }}</span>（{{ strategies[0].strategy_type === 'local' ? '本地' : 'S3' }}）
			</div>
		</div>

		<n-upload multiple directory-dnd :custom-request="handleUpload" :show-file-list="false" accept="image/*">
			<n-upload-dragger>
				<div class="py-10 text-center">
					<n-icon size="40" class="text-[var(--color-text-tertiary)] mb-3"><CloudUploadOutline /></n-icon>
					<div class="text-sm text-[var(--color-text-primary)]">点击或拖拽文件到此区域上传</div>
					<div class="text-xs text-[var(--color-text-tertiary)] mt-1">支持 JPG、PNG、GIF、WebP、BMP、SVG 格式</div>
				</div>
			</n-upload-dragger>
		</n-upload>

		<div v-if="results.length" class="mt-6">
			<div class="flex items-center justify-between mb-3">
				<h3 class="text-sm font-medium text-[var(--color-text-primary)]">上传结果</h3>
				<n-button size="tiny" quaternary @click="results = []">清除</n-button>
			</div>
			<div class="space-y-3">
				<n-card v-for="(item, i) in results" :key="i" size="small">
					<div class="flex items-center gap-4">
						<template v-if="item.status === 'uploading'">
							<div class="flex-1 min-w-0 flex items-center gap-2">
								<span class="text-sm font-medium truncate">{{ item.origin_name || item.key }}</span>
								<span class="text-xs text-[var(--color-text-tertiary)]">{{ formatSize(item.size_bytes) }}</span>
								<span class="text-xs text-[var(--color-text-tertiary)] ml-auto">{{ item.progress }}%</span>
							</div>
							<div class="shrink-0 w-6 h-6 rounded-full border-2 border-[var(--color-border)] border-t-[var(--color-primary)] animate-spin"></div>
						</template>
						<template v-else>
							<img
								v-if="item.thumbnail_url"
								:src="toRelative(item.thumbnail_url)"
								class="w-16 h-16 object-cover rounded-lg shrink-0"
							/>
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2">
									<span class="text-sm font-medium truncate">{{ item.origin_name || item.key }}</span>
									<span class="text-xs text-[var(--color-text-tertiary)]">{{ formatSize(item.size_bytes) }}</span>
								</div>
								<n-input-group class="mt-1.5">
									<n-input :value="item.links?.url" size="small" readonly />
									<n-button size="small" @click="copyText(item.links?.url)">复制</n-button>
								</n-input-group>
								<div class="flex gap-1.5 mt-1.5">
									<n-button size="tiny" quaternary @click="copyText(item.links?.markdown)">Markdown</n-button>
									<n-button size="tiny" quaternary @click="copyText(item.links?.bbcode)">BBCode</n-button>
									<n-button size="tiny" quaternary @click="copyText(item.links?.html)">HTML</n-button>
								</div>
							</div>
							<n-tag size="small" :type="item.status === 'ok' ? 'success' : 'error'" :bordered="false">
								{{ item.status === 'ok' ? '成功' : item.message || '失败' }}
							</n-tag>
						</template>
					</div>
				</n-card>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { NIcon } from 'naive-ui'
import { CloudUploadOutline } from '@vicons/ionicons5'
import {
	NUpload, NUploadDragger, NCard, NInputGroup, NInput, NButton,
	NTag, NSelect, useMessage,
} from 'naive-ui'
import { uploadImage } from '../api/image'
import { getStrategies, type Strategy } from '../api/strategies'
import { useUserStore } from '../stores/user'

const message = useMessage()
const userStore = useUserStore()
const results = ref<any[]>([])
const strategies = ref<Strategy[]>([])
const selectedStrategyId = ref<number | null>(null)

const strategyOptions = computed(() =>
	strategies.value.map((s) => ({ label: `${s.name} (${s.strategy_type === 'local' ? '本地' : 'S3'})`, value: s.id })),
)

function onStrategyChange(val: number) {
	selectedStrategyId.value = val
	localStorage.setItem('default_strategy_id', String(val))
}

onMounted(async () => {
	try {
		const res = await getStrategies()
		strategies.value = res.data.data || []
		let initialId: number | null = null
		const userSettings = (userStore.user as any)?.settings
		const userDefault = userSettings?.default_strategy ? Number(userSettings.default_strategy) : null
		const saved = localStorage.getItem('default_strategy_id')
		if (userDefault && strategies.value.some((s) => s.id === userDefault)) {
			initialId = userDefault
		} else if (saved && strategies.value.some((s) => s.id === Number(saved))) {
			initialId = Number(saved)
		} else if (strategies.value.length > 0) {
			initialId = strategies.value[0].id
		}
		selectedStrategyId.value = initialId
	} catch { /* */ }
})

function formatSize(bytes: number) {
	if (!bytes) return '0 B'
	const k = 1024
	const sizes = ['B', 'KB', 'MB', 'GB']
	const i = Math.floor(Math.log(bytes) / Math.log(k))
	return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}

function copyText(text?: string) {
	if (!text) return
	navigator.clipboard.writeText(text).then(() => message.success('已复制')).catch(() => {})
}

function toRelative(url: string) {
	try { return new URL(url).pathname }
	catch { return url }
}

async function handleUpload({ file, onFinish, onError }: any) {
	const raw = file.file || file
	const resultItem = { key: file.name || 'unknown', origin_name: file.name, status: 'uploading', progress: 0, size_bytes: raw.size || 0 }
	results.value.unshift(resultItem)
	const index = results.value.indexOf(resultItem)

	try {
		const params: Record<string, string> = {}
		if (selectedStrategyId.value != null) params.strategy_id = String(selectedStrategyId.value)
		const res = await uploadImage(raw, params, (percent: number) => {
			resultItem.progress = percent
			results.value[index] = { ...resultItem }
		})
		const data = res.data.data
		results.value[index] = { ...data, status: 'ok', progress: 100 }
		message.success(`上传成功: ${data.key}`)
		onFinish()
	} catch (err: any) {
		const msg = err.response?.data?.message || '上传失败'
		results.value[index] = { ...resultItem, status: 'error', message: msg, progress: 0 }
		message.error(`上传失败: ${file.name || '文件'}`)
		onError()
	}
}
</script>
