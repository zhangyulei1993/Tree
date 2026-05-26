import axios from 'axios'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api',
  timeout: 20000
})

request.interceptors.response.use(
  (res) => {
    const body = res.data

    if (body && typeof body.code !== 'undefined') {
      if (body.code === 0) {
        return body.data
      }

      return Promise.reject(new Error(body.msg || 'request failed'))
    }

    return body
  },
  (err) => Promise.reject(err)
)

export default request