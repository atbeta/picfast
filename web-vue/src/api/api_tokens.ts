import api from './index'
import type { ApiResult } from './types'

export interface ApiToken {
	id: number
	name: string
	token?: string
	scopes: string[]
	last_used_at?: string
	expires_at?: string
	created_at: string
}

export function createApiToken(data: { name: string; expires_in: string; scopes?: string[] }) {
	return api.post<ApiResult<ApiToken>>('/api-tokens', data)
}

export function listApiTokens() {
	return api.get<ApiResult<ApiToken[]>>('/api-tokens')
}

export function deleteApiToken(id: number) {
	return api.delete(`/api-tokens/${id}`)
}
