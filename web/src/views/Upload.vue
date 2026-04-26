<template>
	<div>
		<n-h2>上传图片</n-h2>
		<n-upload multiple directory-dnd :custom-request="handleUpload" :show-file-list="false" accept="image/*">
			<n-upload-dragger>
				<div style="padding: 40px 0; text-align: center">
					<n-text style="display: block; font-size: 16px">点击或拖拽文件到此区域上传</n-text>
					<n-text depth="3" style="display: block; margin-top: 4px">支持 JPG、PNG、GIF、WebP、BMP、SVG 格式</n-text>
				</div>
			</n-upload-dragger>
		</n-upload>

		<div v-if="results.length" style="margin-top: 24px">
			<n-h3>上传结果</n-h3>
			<n-grid :cols="1" :x-gap="12" :y-gap="12">
				<n-gi v-for="(item, i) in results" :key="i">
					<n-card size="small">
						<div style="display: flex; align-items: center; gap: 16px">
							<n-image
								v-if="item.thumbnail_url"
								:src="item.thumbnail_url"
								width="80"
								height="80"
								object-fit="cover"
								preview-disabled
							/>
							<div style="flex: 1; min-width: 0">
								<n-text strong>{{ item.origin_name || item.key }}</n-text>
								<n-text depth="3" style="margin-left: 8px; font-size: 12px">{{ formatSize(item.size_bytes) }}</n-text>
								<n-input-group style="margin-top: 4px">
									<n-input :value="item.links?.url" size="small" readonly />
									<n-button size="small" @click="copyText(item.links?.url)">复制链接</n-button>
								</n-input-group>
								<div style="display: flex; gap: 8px; margin-top: 4px">
									<n-button size="tiny" @click="copyText(item.links?.markdown)">Markdown</n-button>
									<n-button size="tiny" @click="copyText(item.links?.bbcode)">BBCode</n-button>
									<n-button size="tiny" @click="copyText(item.links?.html)">HTML</n-button>
								</div>
							</div>
							<n-tag v-if="item.status !== 'uploading'" size="small" :type="item.status === 'ok' ? 'success' : 'error'">
								{{ item.status === 'ok' ? '上传成功' : item.message || '上传失败' }}
							</n-tag>
							<div v-else style="width: 60px">
								<n-progress type="circle" :percentage="item.progress" :show-indicator="false" :stroke-width="6" :width="40" />
							</div>
						</div>
					</n-card>
				</n-gi>
			</n-grid>
			<n-button style="margin-top: 16px" @click="results = []">清除结果</n-button>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import {
	NH2,
	NH3,
	NUpload,
	NUploadDragger,
	NText,
	NGrid,
	NGi,
	NCard,
	NImage,
	NInputGroup,
	NInput,
	NButton,
	NTag,
	NProgress,
	useMessage,
} from 'naive-ui'
import { uploadImage } from '../api/image'

const message = useMessage()
const results = ref<any[]>([])

function formatSize(bytes: number) {
	if (!bytes) return '0 B'
	const k = 1024
	const sizes = ['B', 'KB', 'MB', 'GB']
	const i = Math.floor(Math.log(bytes) / Math.log(k))
	return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}

function copyText(text?: string) {
	if (!text) return
	navigator.clipboard
		.writeText(text)
		.then(() => message.success('已复制'))
		.catch(() => {})
}

async function handleUpload({ file, onFinish, onError }: any) {
	const raw = file.file || file
	const resultItem = {
		key: file.name || 'unknown',
		origin_name: file.name,
		status: 'uploading',
		progress: 0,
		size_bytes: raw.size || 0,
	}
	results.value.unshift(resultItem)
	const index = results.value.indexOf(resultItem)

	try {
		const res = await uploadImage(raw, undefined, (percent: number) => {
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
