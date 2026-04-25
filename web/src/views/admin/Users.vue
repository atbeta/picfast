<template>
  <div>
    <n-h3>User Management</n-h3>
    <n-data-table :columns="columns" :data="users" :loading="loading" :pagination="pagination" remote
      @update:page="fetchUsers" />
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
  { title: 'Name', key: 'name' },
  { title: 'Email', key: 'email' },
  { title: 'Role', key: 'role', width: 80, render: (row: any) => h(NTag, { type: row.role === 'admin' ? 'info' : 'default', size: 'small' }, () => row.role) },
  {
    title: 'Status', key: 'status', width: 80,
    render: (row: any) => h(NTag, { type: row.status === 1 ? 'success' : 'error', size: 'small' }, () => row.status === 1 ? 'Active' : 'Frozen'),
  },
  {
    title: 'Group', key: 'group_id', width: 80,
    render: (row: any) => row.group_name || row.group_id || '-',
  },
  {
    title: 'Used', key: 'used_capacity', width: 100,
    render: (row: any) => formatSize(row.used_capacity || 0),
  },
  {
    title: 'Actions', key: 'actions', width: 180,
    render: (row: any) =>
      h(NSpace, { size: 'small' }, () => [
        h(NButton, { size: 'small', onClick: () => toggleStatus(row) }, () => row.status === 1 ? 'Freeze' : 'Activate'),
        h(NButton, { size: 'small', type: 'error', onClick: () => confirmDelete(row) }, () => 'Delete'),
      ]),
  },
]

onMounted(() => fetchUsers())

async function fetchUsers(page = 1) {
  loading.value = true
  try {
    const res = await adminGetUsers({ page: String(page), page_size: '20' })
    users.value = res.data.data
    pagination.value.itemCount = res.data.pagination?.total || 0
    pagination.value.page = page
  } catch {
    message.error('Failed to load users')
  } finally {
    loading.value = false
  }
}

async function toggleStatus(user: any) {
  const newStatus = user.status === 1 ? 0 : 1
  try {
    await adminUpdateUser(user.id, { status: newStatus })
    message.success('User status updated')
    fetchUsers(pagination.value.page)
  } catch {
    message.error('Failed to update user')
  }
}

function confirmDelete(user: any) {
  dialog.warning({
    title: 'Delete User',
    content: `Delete user "${user.name || user.email}"? All their images will be deleted.`,
    positiveText: 'Delete',
    negativeText: 'Cancel',
    onPositiveClick: async () => {
      try {
        await adminDeleteUser(user.id)
        message.success('User deleted')
        fetchUsers(pagination.value.page)
      } catch {
        message.error('Failed to delete user')
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
