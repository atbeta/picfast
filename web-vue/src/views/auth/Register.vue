<template>
	<div class="flex items-center justify-center min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
		<div class="w-full max-w-sm">
			<div class="text-center mb-8">
				<h1 class="text-2xl font-bold text-[var(--color-text-primary)] tracking-tight">PicFast</h1>
				<p class="mt-1 text-sm text-[var(--color-text-secondary)]">创建新账号</p>
			</div>
			<n-card :bordered="true" class="shadow-sm">
				<n-form ref="formRef" :model="form" :rules="rules" @submit.prevent="onSubmit">
					<n-form-item label="昵称" path="name">
						<n-input v-model:value="form.name" placeholder="你的昵称" />
					</n-form-item>
					<n-form-item label="邮箱" path="email">
						<n-input v-model:value="form.email" placeholder="your@email.com" />
					</n-form-item>
					<n-form-item label="密码" path="password">
						<n-input v-model:value="form.password" type="password" placeholder="至少 8 位" />
					</n-form-item>
					<n-button type="primary" block :loading="loading" attr-type="submit">注册</n-button>
				</n-form>
				<p class="mt-4 text-center text-sm text-[var(--color-text-secondary)]">
					已有账号？<router-link to="/login" class="text-[var(--color-primary)] hover:underline">登录</router-link>
				</p>
			</n-card>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui'
import { register } from '../../api/auth'
import { useUserStore } from '../../stores/user'

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()
const loading = ref(false)

const form = reactive({ name: '', email: '', password: '' })
const rules = {
	name: { required: true, message: '请输入昵称', trigger: 'blur' },
	email: { required: true, message: '请输入邮箱', trigger: 'blur' },
	password: { required: true, min: 8, message: '密码至少 8 位', trigger: 'blur' },
}

async function onSubmit() {
	loading.value = true
	try {
		const res = await register(form.email, form.password, form.name)
		const { access_token = '', refresh_token = '' } = res.data.data
		userStore.setTokens(access_token, refresh_token)
		await userStore.fetchProfile()
		message.success('注册成功')
		router.push('/')
	} catch (err: any) {
		message.error(err.response?.data?.message || '注册失败')
	} finally {
		loading.value = false
	}
}
</script>
