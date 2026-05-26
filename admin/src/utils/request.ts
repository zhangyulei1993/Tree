import axios, { type AxiosRequestConfig } from 'axios'

const instance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api',
  timeout: 20000
})

instance.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_token')

  if (token) {
    config.headers.Authorization = 'Bearer ' + token
  }

  return config
})

instance.interceptors.response.use(
  (res) => {
    const body = res.data

    if (body && typeof body.code !== 'undefined') {
      if (body.code === 0) {
        return body.data
      }

      if (body.code === 401) {
        localStorage.removeItem('admin_token')
        localStorage.removeItem('admin_user')
        window.location.href = '/login'
      }

      return Promise.reject(new Error(body.msg || 'request failed'))
    }

    return body
  },
  (err) => Promise.reject(err)
)

type RequestClient = {
  get<T>(url: string, config?: AxiosRequestConfig): Promise<T>
  post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>
  delete<T>(url: string, config?: AxiosRequestConfig): Promise<T>
}

export default instance as unknown as RequestClient