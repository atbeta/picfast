<template>
	<div style="max-width: 600px; margin: 0 auto">
		<n-h2>个人设置</n-h2>

		<n-spin :show="loading">
			<n-form label-placement="left" label-width="100">
				<n-form-item label="显示名称">
					<n-input v-model:value="form.name" placeholder="你的昵称" />
				</n-form-item>

				<n-divider>安全</n-divider>

				<n-form-item label="修改密码">
					<n-input v-model:value="form.password" type="password" placeholder="留空则不修改" show-password-on="click" />
				</n-form-item>

				<n-divider>偏好</n-divider>

				<n-form-item label="默认策略">
					<n-select
						v-model:value="form.defaultStrategy"
						:options="strategyOptions"
						placeholder="跟随组默认"
						clearable
						style="width: 100%"
					/>
				</n-form-item>

				<n-form-item>
					<n-button type="primary" :loading="saving" @click="save">保存</n-button>
				</n-form-item>
			</n-form>
		</n-spin>
	</div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { NH2, NForm, NFormItem, NInput, NSelect, NButton, NDivider, NSpin, useMessage } from 'naive-ui'
import api from '../api/index'
import { getStrategies, type Strategy } from '../api/strategies'
import { useUserStore } from '../stores/user'

const message = useMessage()
const userStore = useUserStore()
const loading = ref(false)
const saving = ref(false)
const strategies = ref<Strategy[]>([])

const form = reactive({
	name: '',
	password: '',
	defaultStrategy: 0 as number,
})

const strategyOptions = computed(() =>
	[
		{ label: '跟随组默认', value: 0 },
		...strategies.value.map((s) => ({
			label: `${s.name} (${s.strategy_type === 'local' ? '本地' : 'S3'})`,
			value: s.id,
		})),
	],
)

onMounted(async () => {
	loading.value = true
	try {
		const [profileRes, stratRes] = await Promise.all([api.get('/users/me'), getStrategies()])
		const user = profileRes.data.data
		form.name = user.name || ''
		strategies.value = stratRes.data.data || []
		const settings = user.settings || {}
		if (settings.default_strategy) {
			form.defaultStrategy = Number(settings.default_strategy)
		}
	} catch {
		message.error('加载失败')
	} finally {
		loading.value = false
	}
})

async function save() {
	saving.value = true
	try {
		const data: Record<string, any> = { name: form.name }
		if (form.password) {
			data.password = form.password
		}
		if (form.defaultStrategy && form.defaultStrategy !== 0) {
			data.settings = { default_strategy: form.defaultStrategy }
		} else {
			data.settings = { default_strategy: null }
		}
		await api.put('/users/me', data)
		message.success('保存成功')
		userStore.fetchProfile()
	} catch (err: any) {
		message.error(err.response?.data?.message || '保存失败')
	} finally {
		saving.value = false
	}
}
</script>