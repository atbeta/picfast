import api from './index'
import type { ApiResult } from './types'

export interface Strategy {
	id: number
	name: string
	strategy_type: string
	configs: Record<string, any>
	created_at: string
	updated_at: string
}

export function getStrategies() {
	return api.get<ApiResult<Strategy[]>>('/strategies')
}
