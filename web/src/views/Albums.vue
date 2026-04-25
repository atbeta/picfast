<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <n-h2 class="m-0">Albums</n-h2>
      <n-button type="primary" @click="showCreate = true">New Album</n-button>
    </div>

    <n-spin :show="loading">
      <n-grid :cols="3" :x-gap="12" :y-gap="12" responsive="screen">
        <n-gi v-for="album in albums" :key="album.id">
          <n-card size="small" :title="album.name">
            <template #header-extra>
              <n-dropdown :options="albumActions" @select="(k: string) => onAlbumAction(k, album)">
                <n-button quaternary size="small">···</n-button>
              </n-dropdown>
            </template>
            <n-text depth="3">{{ album.intro || 'No description' }}</n-text>
            <template #footer>
              <n-text depth="3" class="text-xs">{{ album.image_count || 0 }} images</n-text>
            </template>
          </n-card>
        </n-gi>
      </n-grid>
      <n-empty v-if="!loading && albums.length === 0" description="No albums yet" class="mt-12" />
    </n-spin>

    <n-modal v-model:show="showCreate" preset="dialog" title="Create Album" positive-text="Create" negative-text="Cancel"
      @positive-click="createAlbumFn">
      <n-form>
        <n-form-item label="Name">
          <n-input v-model:value="createForm.name" placeholder="Album name" />
        </n-form-item>
        <n-form-item label="Description">
          <n-input v-model:value="createForm.intro" type="textarea" placeholder="Optional description" />
        </n-form-item>
      </n-form>
    </n-modal>

    <n-modal v-model:show="showEdit" preset="dialog" title="Edit Album" positive-text="Save" negative-text="Cancel"
      @positive-click="updateAlbumFn">
      <n-form>
        <n-form-item label="Name">
          <n-input v-model:value="editForm.name" />
        </n-form-item>
        <n-form-item label="Description">
          <n-input v-model:value="editForm.intro" type="textarea" />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
  NH2, NGrid, NGi, NCard, NText, NEmpty, NSpin, NButton, NDropdown,
  NModal, NForm, NFormItem, NInput, useMessage, useDialog,
} from 'naive-ui'
import { getAlbums, createAlbum, updateAlbum, deleteAlbum } from '../api/album'

const message = useMessage()
const dialog = useDialog()
const albums = ref<any[]>([])
const loading = ref(false)
const showCreate = ref(false)
const showEdit = ref(false)
const createForm = reactive({ name: '', intro: '' })
const editForm = reactive({ id: 0, name: '', intro: '' })

const albumActions = [
  { label: 'Edit', key: 'edit' },
  { label: 'Delete', key: 'delete' },
]

onMounted(() => fetchAlbums())

async function fetchAlbums() {
  loading.value = true
  try {
    const res = await getAlbums(1, 100)
    albums.value = res.data.data
  } catch {
    message.error('Failed to load albums')
  } finally {
    loading.value = false
  }
}

async function createAlbumFn() {
  if (!createForm.name) { message.warning('Name is required'); return false }
  try {
    await createAlbum(createForm.name, createForm.intro)
    message.success('Album created')
    createForm.name = ''
    createForm.intro = ''
    fetchAlbums()
  } catch {
    message.error('Failed to create album')
  }
  return true
}

function onAlbumAction(key: string, album: any) {
  if (key === 'edit') {
    editForm.id = album.id
    editForm.name = album.name
    editForm.intro = album.intro || ''
    showEdit.value = true
  } else if (key === 'delete') {
    dialog.warning({
      title: 'Delete Album',
      content: `Delete "${album.name}"? Images will not be deleted.`,
      positiveText: 'Delete',
      negativeText: 'Cancel',
      onPositiveClick: async () => {
        try {
          await deleteAlbum(album.id)
          message.success('Album deleted')
          fetchAlbums()
        } catch {
          message.error('Failed to delete album')
        }
      },
    })
  }
}

async function updateAlbumFn() {
  try {
    await updateAlbum(editForm.id, { name: editForm.name, intro: editForm.intro })
    message.success('Album updated')
    fetchAlbums()
  } catch {
    message.error('Failed to update album')
  }
  return true
}
</script>
