<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <n-h3 class="m-0">Groups</n-h3>
      <n-button type="primary" @click="openCreate">New Group</n-button>
    </div>
    <n-data-table :columns="columns" :data="groups" :loading="loading" />

    <n-modal v-model:show="showModal" preset="dialog" :title="editing ? 'Edit Group' : 'Create Group'"
      positive-text="Save" negative-text="Cancel" @positive-click="saveGroup">
      <n-form>
        <n-form-item label="Name">
          <n-input v-model:value="form.name" />
        </n-form-item>
        <n-form-item label="Is Default">
          <n-switch v-model:value="form.is_default" />
        </n-form-item>
        <n-form-item label="Max File Size (MB)">
          <n-input-number v-model:value="form.max_size" :min="0" class="w-full" />
        </n-form-item>
        <n-form-item label="Accepted Extensions">
          <n-input v-model:value="form.extensions" placeholder="jpg,png,gif,webp" />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, reactive, onMounted } from 'vue'
import { NH3, NDataTable, NButton, NTag, NSpace, NModal, NForm, NFormItem, NInput, NInputNumber, NSwitch, useMessage, type DataTableColumns } from 'naive-ui'
import { adminGetGroups, adminCreateGroup, adminUpdateGroup, adminDeleteGroup } from '../../api/admin'

const message = useMessage()
const groups = ref<any[]>([])
const loading = ref(false)
const showModal = ref(false)
const editing = ref<any>(null)
const form = reactive({ name: '', is_default: false, max_size: 10, extensions: 'jpg,jpeg,png,gif,webp,bmp' })

const columns: DataTableColumns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: 'Name', key: 'name' },
  { title: 'Default', key: 'is_default', width: 80, render: (row: any) => h(NTag, { type: row.is_default ? 'success' : 'default', size: 'small' }, () => row.is_default ? 'Yes' : 'No') },
  { title: 'Users', key: 'user_count', width: 70 },
  {
    title: 'Actions', key: 'actions', width: 160,
    render: (row: any) =>
      h(NSpace, { size: 'small' }, () => [
        h(NButton, { size: 'small', onClick: () => openEdit(row) }, () => 'Edit'),
        h(NButton, { size: 'small', type: 'error', onClick: () => deleteGroup(row), disabled: row.is_default }, () => 'Delete'),
      ]),
  },
]

onMounted(() => fetchGroups())

async function fetchGroups() {
  loading.value = true
  try {
    const res = await adminGetGroups()
    groups.value = res.data.data
  } catch {
    message.error('Failed to load groups')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  form.name = ''
  form.is_default = false
  form.max_size = 10
  form.extensions = 'jpg,jpeg,png,gif,webp,bmp'
  showModal.value = true
}

function openEdit(group: any) {
  editing.value = group
  form.name = group.name
  form.is_default = group.is_default
  const configs = group.configs || {}
  form.max_size = Math.round((configs.max_size || 10485760) / 1048576)
  form.extensions = (configs.extensions || []).join(',')
  showModal.value = true
}

async function saveGroup() {
  const configs = {
    max_size: form.max_size * 1048576,
    extensions: form.extensions.split(',').map(s => s.trim()).filter(Boolean),
  }
  try {
    if (editing.value) {
      await adminUpdateGroup(editing.value.id, { name: form.name, is_default: form.is_default, configs })
      message.success('Group updated')
    } else {
      await adminCreateGroup({ name: form.name, is_default: form.is_default, configs })
      message.success('Group created')
    }
    fetchGroups()
  } catch (err: any) {
    message.error(err.response?.data?.message || 'Failed to save group')
  }
  return true
}

async function deleteGroup(group: any) {
  try {
    await adminDeleteGroup(group.id)
    message.success('Group deleted')
    fetchGroups()
  } catch (err: any) {
    message.error(err.response?.data?.message || 'Failed to delete group')
  }
}
</script>
