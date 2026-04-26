<template>
	<div style="max-width: 900px; margin: 0 auto">
		<n-space vertical :size="24">
			<n-space align="center" justify="space-between">
				<h2 style="margin: 0">API 令牌</h2>
				<n-button type="primary" @click="showCreate = true">创建令牌</n-button>
			</n-space>

			<n-alert v-if="tokens.length === 0 && !loading" type="info" title="暂无令牌">
				你可以创建 API 令牌，用于第三方客户端或 MCP 集成访问你的图片库。
			</n-alert>

			<n-spin :show="loading">
				<n-list bordered>
					<n-list-item v-for="t in tokens" :key="t.id">
						<n-thing :title="t.name" :description="formatDate(t.created_at)">
							<template #header-extra>
								<n-space>
									<n-tag v-for="scope in t.scopes" :key="scope" size="small" type="info">{{ scope }}</n-tag>
									<n-popconfirm @positive-click="removeToken(t.id)">
										<template #trigger>
											<n-button text type="error" size="small">删除</n-button>
										</template>
										确认删除此令牌？
									</n-popconfirm>
								</n-space>
							</template>
							<template #description>
								<div style="color: #6b7280; font-size: 13px">
									创建于 {{ formatDate(t.created_at) }}
									<span v-if="t.expires_at">· 过期于 {{ formatDate(t.expires_at) }}</span>
									<span v-else>· 永不过期</span>
									<span v-if="t.last_used_at">· 上次使用 {{ formatDate(t.last_used_at) }}</span>
								</div>
							</template>
						</n-thing>
					</n-list-item>
				</n-list>
			</n-spin>
		</n-space>

		<!-- Create modal -->
		<n-modal v-model:show="showCreate" title="创建 API 令牌" preset="card" style="width: 480px">
			<n-form :model="form" :rules="rules" ref="formRef" label-placement="left" label-width="auto">
				<n-form-item label="名称" path="name">
					<n-input v-model:value="form.name" placeholder="例如：MCP 客户端" />
				</n-form-item>
				<n-form-item label="有效期" path="expires_in">
					<n-select v-model:value="form.expires_in" :options="expireOptions" />
				</n-form-item>
				<n-form-item label="权限" path="scopes">
					<n-checkbox-group v-model:value="form.scopes">
						<n-checkbox value="read">读取 (read)</n-checkbox>
						<n-checkbox value="write">写入 (write)</n-checkbox>
					</n-checkbox-group>
				</n-form-item>
			</n-form>
			<n-alert v-if="newToken" type="warning" title="请立即复制令牌" style="margin-bottom: 16px">
				<div style="word-break: break-all; font-family: monospace; background: #f3f4f6; padding: 8px; border-radius: 4px; margin-top: 8px">
					{{ newToken }}
				</div>
				<n-button text type="primary" size="small" style="margin-top: 8px" @click="copyToken">复制到剪贴板</n-button>
			</n-alert>
			<template #footer>
				<n-space justify="end">
					<n-button @click="showCreate = false">取消</n-button>
					<n-button type="primary" :loading="creating" @click="createToken">创建</n-button>
				</n-space>
			</template>
		</n-modal>
	</div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
	NSpace, NButton, NList, NListItem, NThing, NTag, NSpin, NAlert,
	NModal, NForm, NFormItem, NInput, NSelect, NCheckbox, NCheckboxGroup, NPopconfirm,
	useMessage,
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

onMounted(loadTokens)

async function loadTokens() {
	loading.value = true
	try {
		const { data } = await listApiTokens()
		tokens.value = data.data || []
	} catch (e) {
		message.error('加载令牌失败')
	} finally {
		loading.value = false
	}
}

async function createToken() {
	try {
		await formRef.value?.validate()
	} catch {
		return
	}
	creating.value = true
	try {
		const { data } = await createApiToken({
			name: form.name,
			expires_in: form.expires_in,
			scopes: form.scopes,
		})
		newToken.value = data.data?.token || ''
		message.success('令牌创建成功')
		loadTokens()
	} catch (e) {
		message.error('创建令牌失败')
	} finally {
		creating.value = false
	}
}

async function removeToken(id: number) {
	try {
		await deleteApiToken(id)
		message.success('令牌已删除')
		loadTokens()
	} catch {
		message.error('删除失败')
	}
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
