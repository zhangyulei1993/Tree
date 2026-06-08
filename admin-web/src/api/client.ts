import axios from 'axios'

export const apiClient = axios.create({
  baseURL: '/api',
  timeout: 10000
})

apiClient.interceptors.request.use((config) => {
  const token = sessionStorage.getItem('tree_admin_mock_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export const apiMode = 'mock'
