import axios from 'axios'

const api = axios.create({
	baseURL: '/api/v1',
	headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
	const token = localStorage.getItem('token')
	if (token) {
		config.headers.Authorization = `Bearer ${token}`
	}
	return config
})

let isRefreshing = false
let refreshSubscribers: ((token: string) => void)[] = []

function onRefreshed(token: string) {
	refreshSubscribers.forEach((cb) => cb(token))
	refreshSubscribers = []
}

function addRefreshSubscriber(cb: (token: string) => void) {
	refreshSubscribers.push(cb)
}

api.interceptors.response.use(
	(res) => res,
	async (err) => {
		const originalRequest = err.config
		if (err.response?.status === 401 && originalRequest && !originalRequest._retry) {
			const refreshTokenVal = localStorage.getItem('refresh_token')
			if (!refreshTokenVal) {
				localStorage.removeItem('token')
				window.location.href = '/login'
				return Promise.reject(err)
			}

			originalRequest._retry = true

			if (isRefreshing) {
				return new Promise((resolve) => {
					addRefreshSubscriber((token: string) => {
						originalRequest.headers.Authorization = `Bearer ${token}`
						resolve(api(originalRequest))
					})
				})
			}

			isRefreshing = true
			try {
				const res = await api.post('/auth/refresh', { refresh_token: refreshTokenVal })
				const { access_token, refresh_token } = res.data.data
				localStorage.setItem('token', access_token)
				localStorage.setItem('refresh_token', refresh_token)
				api.defaults.headers.common['Authorization'] = `Bearer ${access_token}`
				onRefreshed(access_token)
				originalRequest.headers.Authorization = `Bearer ${access_token}`
				return api(originalRequest)
			} catch (refreshErr) {
				localStorage.removeItem('token')
				localStorage.removeItem('refresh_token')
				window.location.href = '/login'
				return Promise.reject(refreshErr)
			} finally {
				isRefreshing = false
			}
		}

		if (err.response?.status === 401) {
			localStorage.removeItem('token')
			window.location.href = '/login'
		}
		return Promise.reject(err)
	},
)

export default api
