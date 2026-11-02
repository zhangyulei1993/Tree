import { apiClient, apiMode, unwrapData } from '@/api/client'
import type { LaunchSplashConfig } from '@/types/api'

const mockConfig: LaunchSplashConfig = { id: 1, configKey: 'HOME', enabled: true, title: 'Tree', tagline: '让每一页家谱，都有来处', imageUrl: '', dayImageUrl: '', nightImageUrl: '', durationMs: 3600, skipEnabled: true, updatedAt: new Date().toISOString() }

export async function getLaunchSplashConfig() {
  if (apiMode === 'mock') return mockConfig
  return unwrapData<LaunchSplashConfig>(await apiClient.get('/admin/launch-splash-config'))
}

export async function updateLaunchSplashConfig(input: Omit<LaunchSplashConfig, 'id' | 'configKey' | 'updatedAt'>) {
  if (apiMode === 'mock') { Object.assign(mockConfig, input, { updatedAt: new Date().toISOString() }); return mockConfig }
  return unwrapData<LaunchSplashConfig>(await apiClient.put('/admin/launch-splash-config', input))
}
