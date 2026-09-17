import { isRealApiMode, request } from '@/api/client'

/** 微信小程序好友分享：普通小程序 / 文章 / 成员邀请 语义分离 */

export const MINI_PROGRAM_SHARE_TITLE = '家脉｜记录家族，连接亲人'

export const HOME_SHARE_PATH = '/pages/home/index'

const environment = (import.meta.env || {}) as ImportMetaEnv & {
  VITE_API_BASE_URL?: string
}

const shareImageBaseURL = /^https:\/\//i.test(environment.VITE_API_BASE_URL || '')
  ? String(environment.VITE_API_BASE_URL).replace(/\/$/, '')
  : 'https://tapi.bigbigboy.cn/api'

export const DEFAULT_MINI_PROGRAM_SHARE_IMAGE = `${shareImageBaseURL}/static/content/share/mini-program-default.jpg`

export const INVITATION_SHARE_IMAGE = `${shareImageBaseURL}/static/content/share/family-invitation.jpg`

type ShareConfigKey = 'HOME' | 'PUBLIC_FAMILY' | 'INVITATION_NODE' | 'INVITATION_PENDING_MEMBER'

interface RuntimeShareConfig {
  configKey: ShareConfigKey
  titleTemplate: string
  imageUrl: string
}

const defaultShareConfigs: Record<ShareConfigKey, RuntimeShareConfig> = {
  HOME: { configKey: 'HOME', titleTemplate: MINI_PROGRAM_SHARE_TITLE, imageUrl: DEFAULT_MINI_PROGRAM_SHARE_IMAGE },
  PUBLIC_FAMILY: { configKey: 'PUBLIC_FAMILY', titleTemplate: '{{familyName}}｜公开家庭主页', imageUrl: DEFAULT_MINI_PROGRAM_SHARE_IMAGE },
  INVITATION_NODE: { configKey: 'INVITATION_NODE', titleTemplate: '{{familyName}}邀请你确认「{{targetMemberName}}」身份并加入家庭树', imageUrl: INVITATION_SHARE_IMAGE },
  INVITATION_PENDING_MEMBER: { configKey: 'INVITATION_PENDING_MEMBER', titleTemplate: '{{familyName}}邀请你加入家庭（{{targetMemberName}}）', imageUrl: INVITATION_SHARE_IMAGE }
}

const runtimeShareConfigs = new Map<ShareConfigKey, RuntimeShareConfig>()
let lastShareConfigLoadedAt = 0
let shareConfigLoading: Promise<void> | null = null
const shareConfigCacheTtlMs = 30_000

export async function loadShareConfigs(options: { force?: boolean } = {}) {
  if (!isRealApiMode) return
  const isFresh = runtimeShareConfigs.size > 0 && Date.now() - lastShareConfigLoadedAt < shareConfigCacheTtlMs
  if (!options.force && isFresh) return
  if (shareConfigLoading) return shareConfigLoading
  shareConfigLoading = (async () => {
    try {
      const rows = await request<Array<{ configKey: ShareConfigKey; titleTemplate: string; imageUrl: string }>>('/share-configs', { public: true })
      const next = new Map<ShareConfigKey, RuntimeShareConfig>()
      rows.forEach((row) => {
        if (defaultShareConfigs[row.configKey] && row.titleTemplate && /^https:\/\//i.test(row.imageUrl)) {
          next.set(row.configKey, { configKey: row.configKey, titleTemplate: row.titleTemplate, imageUrl: row.imageUrl })
        }
      })
      if (next.size > 0) {
        runtimeShareConfigs.clear()
        next.forEach((value, key) => runtimeShareConfigs.set(key, value))
      }
      lastShareConfigLoadedAt = Date.now()
    } catch {
      // Keep bundled defaults when the configuration endpoint is unavailable.
    } finally {
      shareConfigLoading = null
    }
  })()
  return shareConfigLoading
}

function shareConfig(key: ShareConfigKey) {
  return runtimeShareConfigs.get(key) || defaultShareConfigs[key]
}

function renderTitle(template: string, values: Record<string, string>) {
  return template.replace(/\{\{(familyName|targetMemberName)\}\}/g, (_, key: string) => values[key] || '')
}

export interface WechatSharePayload {
  title: string
  path: string
  imageUrl: string
}

export interface WechatTimelineSharePayload {
  title: string
  query: string
  imageUrl: string
}

export function buildContentDetailSharePath(articleId: string | number): string {
  const normalized = String(articleId).trim()
  return `/pages/content/detail?id=${encodeURIComponent(normalized)}`
}

export function buildInviteSharePath(inviteToken: string): string {
  return `/pages/invite/detail?inviteToken=${encodeURIComponent(inviteToken.trim())}`
}

export function buildPublicFamilySharePath(familyId: string | number): string {
  const normalized = String(familyId).trim()
  return `/pages/family/public-profile?familyId=${encodeURIComponent(normalized)}`
}

export function resolveArticleShareImage(coverUrl?: string | null): string {
  const trimmed = coverUrl?.trim()
  if (trimmed && /^https:\/\//i.test(trimmed)) {
    return trimmed
  }
  return DEFAULT_MINI_PROGRAM_SHARE_IMAGE
}

export function buildHomeSharePayload(): WechatSharePayload {
  const config = shareConfig('HOME')
  return {
    title: config.titleTemplate,
    path: HOME_SHARE_PATH,
    imageUrl: config.imageUrl
  }
}

export function buildArticleSharePayload(input: {
  id: string | number
  title: string
  coverUrl?: string | null
}): WechatSharePayload {
  const title = input.title.trim() || '阅读精选'
  return {
    title,
    path: buildContentDetailSharePath(input.id),
    imageUrl: resolveArticleShareImage(input.coverUrl)
  }
}

export function buildPublicFamilySharePayload(input: {
  id: string | number
  familyName: string
}): WechatSharePayload {
  const familyName = input.familyName.trim() || '公开家庭'
  const config = shareConfig('PUBLIC_FAMILY')
  return {
    title: renderTitle(config.titleTemplate, { familyName }),
    path: buildPublicFamilySharePath(input.id),
    imageUrl: config.imageUrl
  }
}

export function buildPublicFamilyTimelinePayload(input: {
  id: string | number
  familyName: string
}): WechatTimelineSharePayload {
  const familyName = input.familyName.trim() || '公开家庭'
  return {
    title: `${familyName}｜公开家庭主页`,
    query: `familyId=${encodeURIComponent(String(input.id).trim())}`,
    imageUrl: DEFAULT_MINI_PROGRAM_SHARE_IMAGE
  }
}

export function buildInviteSharePayload(input: {
  inviteToken: string
  familyName: string
  targetMemberName: string
  inviteType?: string
}): WechatSharePayload {
  const config = shareConfig(input.inviteType === 'JOIN_FAMILY_PENDING_MEMBER' ? 'INVITATION_PENDING_MEMBER' : 'INVITATION_NODE')
  return {
    title: renderTitle(config.titleTemplate, { familyName: input.familyName.trim(), targetMemberName: input.targetMemberName.trim() }),
    path: buildInviteSharePath(input.inviteToken),
    imageUrl: config.imageUrl
  }
}
