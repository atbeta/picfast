import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../api/index'

export const useUserStore = defineStore('user', () => {
	const token = ref(localStorage.getItem('token') || '')
	const refreshToken = ref(localStorage.getItem('refresh_token') || '')
	const user = ref<Record<string, unknown> | null>(null)

	function setTokens(access: string, refresh: string) {
		token.value = access
		refreshToken.value = refresh
		localStorage.setItem('token', access)
		localStorage.setItem('refresh_token', refresh)
	}

	function clearTokens() {
		token.value = ''
		refreshToken.value = ''
		localStorage.removeItem('token')
		localStorage.removeItem('refresh_token')
		user.value = null
	}

	async function fetchProfile() {
		const res = await api.get('/users/me')
		user.value = res.data.data
		return user.value
	}

	function isAdmin() {
		return user.value?.role === 'admin'
	}

	return { token, refreshToken, user, setTokens, clearTokens, fetchProfile, isAdmin }
})
