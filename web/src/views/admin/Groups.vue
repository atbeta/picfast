<template>
	<div>
		<div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px">
			<n-h3 style="margin: 0">分组管理</n-h3>
			<n-button type="primary" @click="openCreate">新建分组</n-button>
		</div>
		<n-data-table :columns="columns" :data="groups" :loading="loading" />

		<n-modal
			v-model:show="showModal"
			preset="dialog"
			:title="editing ? '编辑分组' : '新建分组'"
			positive-text="保存"
			negative-text="取消"
			@positive-click="saveGroup"
			style="width: 500px"
		>
			<n-form label-placement="left" label-width="100">
				<n-form-item label="名称">
					<n-input v-model:value="form.name" placeholder="分组名称" />
				</n-form-item>
				<n-form-item label="默认分组">
					<n-switch v-model:value="form.is_default" />
				</n-form-item>
				<n-form-item label="最大文件">
					<n-input-number v-model:value="form.max_size" :min="1" style="width: 100%">
						<template #suffix>MB</template>
					</n-input-number>
				</n-form-item>
				<n-form-item label="允许格式">
					<n-input v-model:value="form.extensions" placeholder="jpg,png,gif,webp" />
				</n-form-item>
				<n-form-item label="每日上限">
					<n-input-number v-model:value="form.limit_per_day" :min="0" style="width: 100%" />
				</n-form-item>
				<n-form-item label="每月上限">
					<n-input-number v-model:value="form.limit_per_month" :min="0" style="width: 100%" />
				</n-form-item>
				<n-divider style="margin: 8px 0">存储策略</n-divider>
				<n-form-item label="可用策略">
					<n-checkbox-group v-model:value="form.strategy_ids">
						<n-space item-style="display: flex; align-items: center">
							<n-checkbox v-for="s in allStrategies" :key="s.id" :value="s.id" :label="s.name" />
						</n-space>
					</n-checkbox-group>
				</n-form-item>
			</n-form>
		</n-modal>
	</div>
</template>

<script setup lang="ts">
import { ref, h, reactive, onMounted } from 'vue'
import {
	NH3,
	NDataTable,
	NButton,
	NTag,
	NSpace,
	NModal,
	NForm,
	NFormItem,
	NInput,
	NInputNumber,
	NSwitch,
	NDivider,
	NCheckboxGroup,
	NCheckbox,
	useMessage,
	type DataTableColumns,
} from 'naive-ui'
import { adminGetGroups, adminCreateGroup, adminUpdateGroup, adminDeleteGroup, adminSetGroupStrategies, adminGetStrategies } from '../../api/admin'

const message = useMessage()
const groups = ref<any[]>([])
const loading = ref(false)
const showModal = ref(false)
const editing = ref<any>(null)
const form = reactive({
	name: '',
	is_default: false,
	max_size: 5,
	extensions: 'jpg,jpeg,png,gif,webp,bmp,svg',
	limit_per_day: 300,
	limit_per_month: 9999,
	strategy_ids: [] as number[],
})

const allStrategies = ref<any[]>([])

const columns: DataTableColumns = [
	{ title: 'ID', key: 'id', width: 60 },
	{ title: '名称', key: 'name' },
	{
		title: '默认',
		key: 'is_default',
		width: 60,
		render: (row: any) =>
			h(NTag, { type: row.is_default ? 'success' : 'default', size: 'small' }, () => (row.is_default ? '是' : '否')),
	},
	{ title: '用户数', key: 'user_count', width: 70 },
	{ title: '最大文件', key: 'max_size', width: 80, render: (row: any) => formatSize(row.configs?.max_size || 0) },
	{ title: '每日上限', key: 'limit', width: 80, render: (row: any) => `${row.configs?.limit_per_day || '-'} 张` },
	{
		title: '策略',
		key: 'strategy_ids',
		width: 150,
		render: (row: any) =>
			h(NSpace, { size: 'small' }, () =>
				(row.strategy_ids || []).length > 0
					? row.strategy_ids.map((id: number) => {
							const s = allStrategies.value.find((s) => s.id === id)
							return h(NTag, { size: 'small', type: 'info' }, () => s ? s.name : id)
						})
					: [h(NTag, { size: 'small' }, () => '无')],
			),
	},
	{
		title: '操作',
		key: 'actions',
		width: 160,
		render: (row: any) =>
			h(NSpace, { size: 'small' }, () => [
				h(NButton, { size: 'small', onClick: () => openEdit(row) }, () => '编辑'),
				h(
					NButton,
					{ size: 'small', type: 'error', onClick: () => deleteGroup(row), disabled: row.is_default },
					() => '删除',
				),
			]),
	},
]

onMounted(() => { fetchGroups(); fetchStrategies() })

async function fetchStrategies() {
	try {
		const res = await adminGetStrategies()
		allStrategies.value = res.data.data || []
	} catch {
		/* ignore */
	}
}

async function fetchGroups() {
	loading.value = true
	try {
		const res = await adminGetGroups()
		const d = res.data.data
		groups.value = Array.isArray(d) ? d : []
	} catch {
		message.error('加载分组失败')
	} finally {
		loading.value = false
	}
}

function openCreate() {
	editing.value = null
	form.name = ''
	form.is_default = false
	form.max_size = 5
	form.extensions = 'jpg,jpeg,png,gif,webp,bmp,svg'
	form.limit_per_day = 300
	form.limit_per_month = 9999
	form.strategy_ids = []
	showModal.value = true
}

function openEdit(group: any) {
	editing.value = group
	form.name = group.name
	form.is_default = group.is_default
	const c = group.configs || {}
	form.max_size = Math.round((c.max_size || 5242880) / 1048576)
	form.extensions = (c.extensions || []).join(',')
	form.limit_per_day = c.limit_per_day || 300
	form.limit_per_month = c.limit_per_month || 9999
	form.strategy_ids = (group.strategy_ids || []).map(Number)
	showModal.value = true
}

async function saveGroup() {
	if (!form.name) {
		message.warning('请输入名称')
		return false
	}
	const configs = {
		max_size: form.max_size * 1048576,
		extensions: form.extensions
			.split(',')
			.map((s) => s.trim())
			.filter(Boolean),
		limit_per_day: form.limit_per_day,
		limit_per_month: form.limit_per_month,
	}
	try {
		if (editing.value) {
			await adminUpdateGroup(editing.value.id, { name: form.name, is_default: form.is_default, configs })
			await adminSetGroupStrategies(editing.value.id, form.strategy_ids)
			message.success('分组已更新')
		} else {
			await adminCreateGroup({ name: form.name, is_default: form.is_default, configs })
			message.success('分组创建成功')
		}
		fetchGroups()
	} catch (err: any) {
		message.error(err.response?.data?.message || '操作失败')
	}
	return true
}

async function deleteGroup(group: any) {
	try {
		await adminDeleteGroup(group.id)
		message.success('分组已删除')
		fetchGroups()
	} catch (err: any) {
		message.error(err.response?.data?.message || '删除失败')
	}
}

function formatSize(bytes: number) {
	if (!bytes) return '-'
	const mb = bytes / 1048576
	return mb >= 1024 ? (mb / 1024).toFixed(1) + ' GB' : mb.toFixed(1) + ' MB'
}
</script>
