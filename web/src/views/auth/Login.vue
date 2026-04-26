<template>
	<div class="flex items-center justify-center min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
		<div class="w-full max-w-sm">
			<div class="text-center mb-8">
				<h1 class="text-2xl font-bold text-[var(--color-text-primary)] tracking-tight">PicFast</h1>
				<p class="mt-1 text-sm text-[var(--color-text-secondary)]">登录到你的图床</p>
			</div>
			<n-card :bordered="true" class="shadow-sm">
				<n-form ref="formRef" :model="form" :rules="rules" @submit.prevent="onSubmit">
					<n-form-item label="邮箱" path="email">
						<n-input v-model:value="form.email" placeholder="your@email.com" />
					</n-form-item>
					<n-form-item label="密码" path="password">
						<n-input v-model:value="form.password" type="password" placeholder="请输入密码" />
					</n-form-item>
					<n-button type="primary" block :loading="loading" attr-type="submit">登录</n-button>
				</n-form>
				<p class="mt-4 text-center text-sm text-[var(--color-text-secondary)]">
					还没有账号？<router-link to="/register" class="text-[var(--color-primary)] hover:underline">注册</router-link>
				</p>
			</n-card>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui'
import { login } from '../../api/auth'
import { useUserStore } from '../../stores/user'

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()
const loading = ref(false)

const form = reactive({ email: '', password: '' })
const rules = {
	email: { required: true, message: '请输入邮箱', trigger: 'blur' },
	password: { required: true, message: '请输入密码', trigger: 'blur' },
}

async function onSubmit() {
	loading.value = true
	try {
		const res = await login(form.email, form.password)
		const { access_token = '', refresh_token = '' } = res.data.data
		userStore.setTokens(access_token, refresh_token)
		await userStore.fetchProfile()
		message.success('登录成功')
		router.push('/')
	} catch (err: any) {
		message.error(err.response?.data?.message || '登录失败')
	} finally {
		loading.value = false
	}
}
</script>
