<template>
	<div>
		<h3 class="text-lg font-semibold text-[var(--color-text-primary)] mb-4">用户管理</h3>
		<n-data-table
			:columns="columns"
			:data="users"
			:loading="loading"
			:pagination="pagination"
			remote
			@update:page="fetchUsers"
		/>
	</div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NH3, NDataTable, NTag, NButton, NSpace, useMessage, useDialog, type DataTableColumns } from 'naive-ui'
import { adminGetUsers, adminUpdateUser, adminDeleteUser } from '../../api/admin'

const message = useMessage()
const dialog = useDialog()
const users = ref<any[]>([])
const loading = ref(false)
const pagination = ref({ page: 1, pageSize: 20, itemCount: 0 })

const columns: DataTableColumns = [
	{ title: 'ID', key: 'id', width: 60 },
	{ title: '昵称', key: 'name' },
	{ title: '邮箱', key: 'email' },
	{
		title: '角色',
		key: 'role',
		width: 80,
		render: (row: any) =>
			h(NTag, { type: row.role === 'admin' ? 'info' : 'default', size: 'small' }, () =>
				row.role === 'admin' ? '管理员' : '用户',
			),
	},
	{
		title: '状态',
		key: 'status',
		width: 80,
		render: (row: any) =>
			h(NTag, { type: row.status === 1 ? 'success' : 'error', size: 'small' }, () =>
				row.status === 1 ? '正常' : '冻结',
			),
	},
	{ title: '已用容量', key: 'used_capacity', width: 100, render: (row: any) => formatSize(row.used_capacity || 0) },
	{ title: '图片数', key: 'image_num', width: 70 },
	{
		title: '操作',
		key: 'actions',
		width: 180,
		render: (row: any) =>
			h(NSpace, { size: 'small' }, () => [
				h(NButton, { size: 'small', onClick: () => toggleStatus(row) }, () => (row.status === 1 ? '冻结' : '激活')),
				h(NButton, { size: 'small', type: 'error', onClick: () => confirmDelete(row) }, () => '删除'),
			]),
	},
]

onMounted(() => fetchUsers())

async function fetchUsers(page = 1) {
	loading.value = true
	try {
		const res = await adminGetUsers(page, 20)
		const d = res.data.data
		users.value = d.items || []
		pagination.value.itemCount = d.total || 0
		pagination.value.page = page
	} catch {
		message.error('加载用户列表失败')
	} finally {
		loading.value = false
	}
}

async function toggleStatus(user: any) {
	const newStatus = user.status === 1 ? 0 : 1
	try {
		await adminUpdateUser(user.id, { status: newStatus })
		message.success(newStatus === 1 ? '已激活' : '已冻结')
		fetchUsers(pagination.value.page)
	} catch {
		message.error('操作失败')
	}
}

function confirmDelete(user: any) {
	dialog.warning({
		title: '确认删除',
		content: `确定要删除用户 "${user.name || user.email}" 吗？该用户的所有图片也将被删除。`,
		positiveText: '删除',
		negativeText: '取消',
		onPositiveClick: async () => {
			try {
				await adminDeleteUser(user.id)
				message.success('用户已删除')
				fetchUsers(pagination.value.page)
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
</script>
