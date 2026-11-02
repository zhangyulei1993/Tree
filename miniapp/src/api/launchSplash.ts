import { isRealApiMode, request } from '@/api/client'

export interface LaunchSplashConfig {
  enabled: boolean
  title: string
  tagline: string
  imageUrl: string
  dayImageUrl: string
  nightImageUrl: string
  durationMs: number
  skipEnabled: boolean
}

export const defaultLaunchSplashConfig: LaunchSplashConfig = {
  enabled: true,
  title: 'Tree',
  tagline: '让每一页家谱，都有来处',
  imageUrl: '',
  dayImageUrl: '',
  nightImageUrl: '',
  durationMs: 3600,
  skipEnabled: true
}

export async function loadLaunchSplashConfig() {
  if (!isRealApiMode) return defaultLaunchSplashConfig
  try {
    const config = await request<LaunchSplashConfig>('/launch-splash-config', { public: true })
    return { ...defaultLaunchSplashConfig, ...config }
  } catch {
    return defaultLaunchSplashConfig
  }
}
