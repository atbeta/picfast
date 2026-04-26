<template>
	<div>
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-lg font-semibold text-[var(--color-text-primary)]">相册</h2>
			<n-button type="primary" size="small" @click="openCreate">新建相册</n-button>
		</div>

		<n-spin :show="loading">
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
				<div v-for="album in albums" :key="album.id"
					class="bg-white rounded-lg border border-[var(--color-card-border)] p-4 transition-shadow hover:shadow-md"
				>
					<div class="flex items-center justify-between mb-2">
						<button class="text-sm font-medium text-[var(--color-text-primary)] hover:text-[var(--color-primary)] transition-colors text-left" @click="viewAlbumImages(album)">
							{{ album.name }}
						</button>
						<n-dropdown :options="albumActions" @select="(k: string) => onAlbumAction(k, album)">
							<n-button quaternary size="tiny" class="text-[var(--color-text-tertiary)]">···</n-button>
						</n-dropdown>
					</div>
					<div class="text-xs text-[var(--color-text-tertiary)]">{{ album.intro || '暂无描述' }}</div>
					<div class="text-xs text-[var(--color-text-tertiary)] mt-2 pt-2 border-t border-[var(--color-divider)]">
						{{ album.image_num || 0 }} 张图片
					</div>
				</div>
			</div>
			<n-empty v-if="!loading && albums.length === 0" description="暂无相册，点击上方按钮创建" class="mt-12" />
		</n-spin>

		<n-modal v-model:show="albumImagesVisible" preset="card" :title="currentAlbum?.name + ' - 图片列表'" class="max-w-2xl">
			<n-spin :show="albumImagesLoading">
				<div class="grid grid-cols-4 gap-2">
					<n-image v-for="img in albumImages" :key="img.key"
						:src="img.thumbnail_url || img.url" width="100%" height="100"
						object-fit="cover" preview-disabled class="rounded"
					/>
				</div>
				<n-empty v-if="!albumImagesLoading && albumImages.length === 0" description="相册内暂无图片" />
			</n-spin>
		</n-modal>

		<n-modal v-model:show="showCreate" preset="dialog" title="新建相册" positive-text="创建" negative-text="取消" @positive-click="createAlbumFn">
			<n-form label-placement="top">
				<n-form-item label="名称">
					<n-input v-model:value="createForm.name" placeholder="相册名称" />
				</n-form-item>
				<n-form-item label="描述">
					<n-input v-model:value="createForm.intro" type="textarea" placeholder="可选描述" :rows="3" />
				</n-form-item>
			</n-form>
		</n-modal>

		<n-modal v-model:show="showEdit" preset="dialog" title="编辑相册" positive-text="保存" negative-text="取消" @positive-click="updateAlbumFn">
			<n-form label-placement="top">
				<n-form-item label="名称">
					<n-input v-model:value="editForm.name" />
				</n-form-item>
				<n-form-item label="描述">
					<n-input v-model:value="editForm.intro" type="textarea" :rows="3" />
				</n-form-item>
			</n-form>
		</n-modal>
	</div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
	NCard, NEmpty, NSpin, NButton, NDropdown, NModal, NForm, NFormItem, NInput, NImage,
	useMessage, useDialog,
} from 'naive-ui'
import { getAlbums, createAlbum, updateAlbum, deleteAlbum } from '../api/album'
import { getImages } from '../api/image'

const message = useMessage()
const dialog = useDialog()
const albums = ref<any[]>([])
const loading = ref(false)
const showCreate = ref(false)
const showEdit = ref(false)
const createForm = reactive({ name: '', intro: '' })
const editForm = reactive({ id: 0, name: '', intro: '' })

const albumImagesVisible = ref(false)
const albumImagesLoading = ref(false)
const albumImages = ref<any[]>([])
const currentAlbum = ref<any>(null)

const albumActions = [
	{ label: '编辑', key: 'edit' },
	{ label: '删除', key: 'delete' },
]

onMounted(() => fetchAlbums())

async function fetchAlbums() {
	loading.value = true
	try { const res = await getAlbums(1, 100); const d = res.data.data; albums.value = d.items || d }
	catch { message.error('加载相册失败') }
	finally { loading.value = false }
}

function openCreate() { createForm.name = ''; createForm.intro = ''; showCreate.value = true }

async function createAlbumFn() {
	if (!createForm.name) { message.warning('请输入相册名称'); return false }
	try { await createAlbum(createForm.name, createForm.intro); message.success('相册创建成功'); fetchAlbums() }
	catch { message.error('创建失败') }
	return true
}

function onAlbumAction(key: string, album: any) {
	if (key === 'edit') {
		editForm.id = album.id; editForm.name = album.name; editForm.intro = album.intro || ''
		showEdit.value = true
	} else if (key === 'delete') {
		dialog.warning({
			title: '确认删除', content: `确定要删除相册 "${album.name}" 吗？图片不会被删除。`,
			positiveText: '删除', negativeText: '取消',
			onPositiveClick: async () => {
				try { await deleteAlbum(album.id); message.success('相册已删除'); fetchAlbums() }
				catch { message.error('删除失败') }
			},
		})
	}
}

async function updateAlbumFn() {
	if (!editForm.name) { message.warning('请输入相册名称'); return false }
	try { await updateAlbum(editForm.id, { name: editForm.name, intro: editForm.intro }); message.success('相册已更新'); fetchAlbums() }
	catch { message.error('更新失败') }
	return true
}

async function viewAlbumImages(album: any) {
	currentAlbum.value = album; albumImagesVisible.value = true; albumImagesLoading.value = true
	try {
		const res = await getImages(1, 50); const d = res.data.data; const all: any[] = d.items || d
		albumImages.value = all.filter((img: any) => img.album_id === album.id)
	} catch { albumImages.value = [] }
	finally { albumImagesLoading.value = false }
}
</script>
