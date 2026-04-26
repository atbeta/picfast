<template>
	<div class="max-w-3xl">
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-lg font-semibold text-[var(--color-text-primary)]">API 令牌</h2>
			<n-button type="primary" size="small" @click="openCreateModal">创建令牌</n-button>
		</div>

		<n-alert v-if="tokens.length === 0 && !loading" type="info" title="暂无令牌" class="mb-4">
			你可以创建 API 令牌，用于第三方客户端或 MCP 集成访问你的图片库。
		</n-alert>

		<n-spin :show="loading">
			<div class="space-y-3">
				<div v-for="t in tokens" :key="t.id"
					class="bg-white rounded-lg border border-[var(--color-card-border)] p-4 flex items-center justify-between"
				>
					<div class="min-w-0">
						<div class="text-sm font-medium text-[var(--color-text-primary)]">{{ t.name }}</div>
						<div class="text-xs text-[var(--color-text-tertiary)] mt-0.5">
							创建于 {{ formatDate(t.created_at) }}
							<span v-if="isRealDate(t.expires_at)">· 过期于 {{ formatDate(t.expires_at) }}</span>
							<span v-else>· 永不过期</span>
							<span v-if="isRealDate(t.last_used_at)">· 上次使用 {{ formatDate(t.last_used_at) }}</span>
						</div>
						<div class="flex gap-1.5 mt-2">
							<n-tag v-for="scope in t.scopes" :key="scope" size="tiny" :bordered="false" type="info">{{ scope }}</n-tag>
						</div>
					</div>
					<n-popconfirm @positive-click="removeToken(t.id)">
						<template #trigger>
							<n-button text type="error" size="small">删除</n-button>
						</template>
						确认删除此令牌？
					</n-popconfirm>
				</div>
			</div>
		</n-spin>

		<n-modal v-model:show="showCreate" title="创建 API 令牌" preset="card" class="max-w-md">
			<n-form :model="form" :rules="rules" ref="formRef" label-placement="top">
				<n-form-item label="名称" path="name">
					<n-input v-model:value="form.name" placeholder="例如：MCP 客户端" />
				</n-form-item>
				<n-form-item label="有效期" path="expires_in">
					<n-select v-model:value="form.expires_in" :options="expireOptions" />
				</n-form-item>
				<n-form-item label="权限" path="scopes">
					<n-checkbox-group v-model:value="form.scopes">
						<div class="flex gap-4">
							<n-checkbox value="read">读取 (read)</n-checkbox>
							<n-checkbox value="write">写入 (write)</n-checkbox>
						</div>
					</n-checkbox-group>
				</n-form-item>
			</n-form>
			<n-alert v-if="newToken" type="warning" title="请立即复制令牌" class="mb-4">
				<div class="break-all font-mono text-xs bg-slate-50 p-3 rounded-lg mt-2 select-all">{{ newToken }}</div>
				<n-button text type="primary" size="small" class="mt-2" @click="copyToken">复制到剪贴板</n-button>
			</n-alert>
			<template #footer>
				<div class="flex justify-end gap-2">
					<n-button @click="showCreate = false">取消</n-button>
					<n-button type="primary" :loading="creating" @click="createToken">创建</n-button>
				</div>
			</template>
		</n-modal>
	</div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
	NButton, NTag, NSpin, NAlert, NModal, NForm, NFormItem, NInput, NSelect,
	NCheckbox, NCheckboxGroup, NPopconfirm, useMessage,
} from 'naive-ui'
import type { FormRules, FormInst } from 'naive-ui'
import { listApiTokens, createApiToken, deleteApiToken, type ApiToken } from '../api/api_tokens'

const message = useMessage()
const loading = ref(false)
const tokens = ref<ApiToken[]>([])
const showCreate = ref(false)
const creating = ref(false)
const newToken = ref('')
const formRef = ref<FormInst | null>(null)

const form = reactive({ name: '', expires_in: 'never', scopes: ['read', 'write'] })

const expireOptions = [
	{ label: '30 天', value: '30d' },
	{ label: '90 天', value: '90d' },
	{ label: '1 年', value: '1y' },
	{ label: '永不过期', value: 'never' },
]

const rules: FormRules = {
	name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
}

function openCreateModal() {
	form.name = ''
	form.expires_in = 'never'
	form.scopes = ['read', 'write']
	newToken.value = ''
	showCreate.value = true
}

onMounted(loadTokens)

async function loadTokens() {
	loading.value = true
	try { const { data } = await listApiTokens(); tokens.value = data.data || [] }
	catch { message.error('加载令牌失败') }
	finally { loading.value = false }
}

	async function createToken() {
	if (newToken.value) return
	try { await formRef.value?.validate() } catch { return }
	creating.value = true
	try {
		const { data } = await createApiToken({ name: form.name, expires_in: form.expires_in, scopes: form.scopes })
		newToken.value = data.data?.token || ''
		message.success('令牌创建成功')
		loadTokens()
	} catch { message.error('创建令牌失败') }
	finally { creating.value = false }
}

function isRealDate(s?: string) {
	if (!s) return false
	const d = new Date(s)
	return !isNaN(d.getTime()) && d.getFullYear() > 1
}

async function removeToken(id: number) {
	try { await deleteApiToken(id); message.success('令牌已删除'); loadTokens() }
	catch { message.error('删除失败') }
}

function copyToken() {
	if (!newToken.value) return
	navigator.clipboard.writeText(newToken.value).then(() => message.success('已复制'))
}

function formatDate(s?: string) {
	if (!s) return '-'
	return new Date(s).toLocaleString('zh-CN')
}
</script>
