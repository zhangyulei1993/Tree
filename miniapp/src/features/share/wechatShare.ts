/** 微信小程序好友分享：普通小程序 / 文章 / 成员邀请 语义分离 */

export const MINI_PROGRAM_SHARE_TITLE = '家脉｜记录家族，连接亲人'

export const HOME_SHARE_PATH = '/pages/home/index'

export const DEFAULT_MINI_PROGRAM_SHARE_IMAGE = '/static/share/mini-program-default.jpg'

export const INVITATION_SHARE_IMAGE = '/static/share/family-invitation.jpg'

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
  return {
    title: MINI_PROGRAM_SHARE_TITLE,
    path: HOME_SHARE_PATH,
    imageUrl: DEFAULT_MINI_PROGRAM_SHARE_IMAGE
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
  return {
    title: `${familyName}｜公开家庭主页`,
    path: buildPublicFamilySharePath(input.id),
    imageUrl: DEFAULT_MINI_PROGRAM_SHARE_IMAGE
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
  const title = input.inviteType === 'JOIN_FAMILY_PENDING_MEMBER'
    ? `${input.familyName} 邀请你加入家庭`
    : `${input.familyName} 邀请你确认「${input.targetMemberName}」身份并加入家庭树`
  return {
    title,
    path: buildInviteSharePath(input.inviteToken),
    imageUrl: INVITATION_SHARE_IMAGE
  }
}
