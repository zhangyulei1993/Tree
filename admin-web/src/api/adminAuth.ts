import { apiClient, unwrapData } from '@/api/client'
import type { AdminInfo, AdminLoginResult } from '@/types/api'

export async function login(username: string, password: string): Promise<AdminLoginResult> {
  const response = await apiClient.post('/admin/auth/login', { username, password })
  return unwrapData<AdminLoginResult>(response)
}

export async function logout(): Promise<void> {
  await apiClient.post('/admin/auth/logout')
}

export async function getAdminMe(): Promise<AdminInfo> {
  const response = await apiClient.get('/admin/me')
  return unwrapData<AdminInfo>(response)
}
