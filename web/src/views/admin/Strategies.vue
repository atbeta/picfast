<template>
	<div>
		<div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px">
			<n-h3 style="margin: 0">存储策略</n-h3>
			<n-button type="primary" @click="openCreate">新建策略</n-button>
		</div>
		<n-data-table :columns="columns" :data="strategies" :loading="loading" />

		<n-modal
			v-model:show="showModal"
			preset="dialog"
			:title="editing ? '编辑策略' : '新建策略'"
			positive-text="保存"
			negative-text="取消"
			@positive-click="saveStrategy"
			style="width: 550px"
		>
			<n-form label-placement="left" label-width="80">
				<n-form-item label="名称">
					<n-input v-model:value="form.name" placeholder="策略名称" />
				</n-form-item>
				<n-form-item label="类型">
					<n-select v-model:value="form.type" :options="typeOptions" :disabled="!!editing" />
				</n-form-item>
				<template v-if="form.type === 'local'">
					<n-form-item label="存储路径"
						><n-input v-model:value="form.localRoot" placeholder="/data/images"
					/></n-form-item>
				</template>
				<template v-if="form.type === 's3'">
					<n-form-item label="Endpoint"
						><n-input v-model:value="form.s3Endpoint" placeholder="https://s3.amazonaws.com"
					/></n-form-item>
					<n-form-item label="Region"><n-input v-model:value="form.s3Region" placeholder="us-east-1" /></n-form-item>
					<n-form-item label="Bucket"><n-input v-model:value="form.s3Bucket" /></n-form-item>
					<n-form-item label="Access Key"><n-input v-model:value="form.s3AccessKey" /></n-form-item>
					<n-form-item label="Secret Key"
						><n-input v-model:value="form.s3SecretKey" type="password" show-password-on="click"
					/></n-form-item>
				</template>
			</n-form>
		</n-modal>
	</div>
</template>

<script setup lang="ts">
import { ref, reactive, h, onMounted } from 'vue'
import {
	NH3,
	NDataTable,
	NButton,
	NSpace,
	NModal,
	NForm,
	NFormItem,
	NInput,
	NSelect,
	useMessage,
	type DataTableColumns,
} from 'naive-ui'
import { adminGetStrategies, adminCreateStrategy, adminUpdateStrategy, adminDeleteStrategy } from '../../api/admin'

const message = useMessage()
const strategies = ref<any[]>([])
const loading = ref(false)
const showModal = ref(false)
const editing = ref<any>(null)
const form = reactive({
	name: '',
	type: 'local' as string,
	localRoot: '/data/images',
	s3Endpoint: '',
	s3Region: 'us-east-1',
	s3Bucket: '',
	s3AccessKey: '',
	s3SecretKey: '',
})

const typeOptions = [
	{ label: '本地存储', value: 'local' },
	{ label: 'S3 兼容存储', value: 's3' },
]

const columns: DataTableColumns = [
	{ title: 'ID', key: 'id', width: 60 },
	{ title: '名称', key: 'name' },
	{ title: '类型', key: 'type', width: 100, render: (row: any) => (row.type === 'local' ? '本地存储' : 'S3 存储') },
	{
		title: '操作',
		key: 'actions',
		width: 160,
		render: (row: any) =>
			h(NSpace, { size: 'small' }, () => [
				h(NButton, { size: 'small', onClick: () => openEdit(row) }, () => '编辑'),
				h(NButton, { size: 'small', type: 'error', onClick: () => deleteStrategy(row) }, () => '删除'),
			]),
	},
]

onMounted(() => fetchStrategies())

async function fetchStrategies() {
	loading.value = true
	try {
		const res = await adminGetStrategies()
		const d = res.data.data
		strategies.value = Array.isArray(d) ? d : []
	} catch {
		message.error('加载策略失败')
	} finally {
		loading.value = false
	}
}

function openCreate() {
	editing.value = null
	form.name = ''
	form.type = 'local'
	form.localRoot = '/data/images'
	form.s3Endpoint = ''
	form.s3Region = 'us-east-1'
	form.s3Bucket = ''
	form.s3AccessKey = ''
	form.s3SecretKey = ''
	showModal.value = true
}

function openEdit(s: any) {
	editing.value = s
	form.name = s.name
	form.type = s.type
	const c = s.configs || {}
	if (s.type === 'local') {
		form.localRoot = c.root || ''
	} else {
		form.s3Endpoint = c.endpoint || ''
		form.s3Region = c.region || ''
		form.s3Bucket = c.bucket || ''
		form.s3AccessKey = c.access_key || ''
		form.s3SecretKey = c.secret_key || ''
	}
	showModal.value = true
}

async function saveStrategy() {
	if (!form.name) {
		message.warning('请输入名称')
		return false
	}
	let configs: Record<string, unknown> = {}
	if (form.type === 'local') {
		configs = { root: form.localRoot, url: '/i' }
	} else {
		configs = {
			endpoint: form.s3Endpoint,
			region: form.s3Region,
			bucket: form.s3Bucket,
			access_key: form.s3AccessKey,
			secret_key: form.s3SecretKey,
		}
	}
	try {
		if (editing.value) {
			await adminUpdateStrategy(editing.value.id, { name: form.name, configs })
			message.success('策略已更新')
		} else {
			await adminCreateStrategy({ name: form.name, strategy_type: form.type, configs })
			message.success('策略创建成功')
		}
		fetchStrategies()
	} catch (err: any) {
		message.error(err.response?.data?.message || '操作失败')
	}
	return true
}

async function deleteStrategy(s: any) {
	try {
		await adminDeleteStrategy(s.id)
		message.success('策略已删除')
		fetchStrategies()
	} catch (err: any) {
		message.error(err.response?.data?.message || '删除失败')
	}
}
</script>
