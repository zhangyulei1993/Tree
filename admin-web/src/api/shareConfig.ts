import { apiClient, apiMode, unwrapData } from '@/api/client'
import type { ShareConfig } from '@/types/api'

const mockConfigs: ShareConfig[] = [
  { id: 1, configKey: 'HOME', titleTemplate: '家脉｜记录家族，连接亲人', imageUrl: 'https://tapi.bigbigboy.cn/api/static/content/share/mini-program-default.jpg', imageUrls: ['https://tapi.bigbigboy.cn/api/static/content/share/mini-program-default.jpg'], description: '首页分享', updatedAt: new Date().toISOString() },
  { id: 2, configKey: 'PUBLIC_FAMILY', titleTemplate: '{{familyName}}｜优质家庭案例', imageUrl: 'https://tapi.bigbigboy.cn/api/static/content/share/mini-program-default.jpg', imageUrls: ['https://tapi.bigbigboy.cn/api/static/content/share/mini-program-default.jpg'], description: '优质家庭案例分享', updatedAt: new Date().toISOString() },
  { id: 3, configKey: 'PUBLIC_FAMILY_TIMELINE', titleTemplate: '{{familyName}}｜优质家庭案例', imageUrl: 'https://tapi.bigbigboy.cn/api/static/content/share/mini-program-default.jpg', imageUrls: ['https://tapi.bigbigboy.cn/api/static/content/share/mini-program-default.jpg'], description: '优质家庭朋友圈分享', updatedAt: new Date().toISOString() },
  { id: 4, configKey: 'INVITATION_NODE', titleTemplate: '{{familyName}}邀请你确认「{{targetMemberName}}」身份并加入家庭树', imageUrl: 'https://tapi.bigbigboy.cn/api/static/content/share/family-invitation.jpg', imageUrls: ['https://tapi.bigbigboy.cn/api/static/content/share/family-invitation.jpg'], description: '节点绑定邀请', updatedAt: new Date().toISOString() },
  { id: 5, configKey: 'INVITATION_REWARD', titleTemplate: '邀请你帮{{familyName}}助力解锁名额', imageUrl: 'https://tapi.bigbigboy.cn/api/static/content/share/family-invitation-v2.jpg', imageUrls: ['https://tapi.bigbigboy.cn/api/static/content/share/family-invitation-v2.jpg'], description: '家庭邀请助力', updatedAt: new Date().toISOString() },
  { id: 6, configKey: 'INVITATION_DETAIL', titleTemplate: '{{familyName}}邀请你确认家庭成员身份', imageUrl: 'https://tapi.bigbigboy.cn/api/static/content/share/family-invitation-v2.jpg', imageUrls: ['https://tapi.bigbigboy.cn/api/static/content/share/family-invitation-v2.jpg'], description: '邀请确认页转发', updatedAt: new Date().toISOString() },
  { id: 7, configKey: 'CONTENT_ARTICLE', titleTemplate: '{{articleTitle}}｜家脉阅读', imageUrl: 'https://tapi.bigbigboy.cn/api/static/content/share/mini-program-default.jpg', imageUrls: ['https://tapi.bigbigboy.cn/api/static/content/share/mini-program-default.jpg'], description: '文章内容分享（文章有封面时优先使用文章封面）', updatedAt: new Date().toISOString() },
]

export async function listShareConfigs() {
  if (apiMode === 'mock') return mockConfigs
  const response = await apiClient.get('/admin/share-configs')
  return unwrapData<ShareConfig[]>(response)
}

export async function updateShareConfig(configKey: string, input: Pick<ShareConfig, 'titleTemplate' | 'imageUrl' | 'imageUrls' | 'description'>) {
  if (apiMode === 'mock') {
    const item = mockConfigs.find((config) => config.configKey === configKey)
    if (item) Object.assign(item, input, { updatedAt: new Date().toISOString() })
    return item
  }
  const response = await apiClient.put(`/admin/share-configs/${configKey}`, input)
  return unwrapData<ShareConfig>(response)
}
