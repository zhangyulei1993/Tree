import { isRealApiMode, request } from '@/api/client'
import { mockUser } from '@/mock/data'
import type { BindWechatPhoneInput, CancelAccountByWechatInput, LoginResult, UserInfo } from '@/types/api'

const clientType = 'WECHAT_MINI_PROGRAM'

function mockUserInfo(withProfile = true, phoneVerified = false): UserInfo {
  return {
    id: 'mock_user',
    phone: phoneVerified ? '13800000000' : null,
    phoneVerified,
    nickname: withProfile ? mockUser.nickname : null,
    status: 'ACTIVE',
  }
}

function mockWechatLoginResult(withProfile = true): LoginResult {
  return { accessToken: mockUser.token, tokenType: 'Bearer', user: mockUserInfo(withProfile) }
}

export async function wechatMiniLogin(code: string) {
  if (!isRealApiMode) return mockWechatLoginResult(true)
  return request<LoginResult, { code: string; clientType: string }>(
    '/auth/wechat-mini/login',
    {
      method: 'POST',
      public: true,
      data: { code, clientType }
    }
  )
}

export async function wechatMiniPhoneLogin(phoneCode: string, code: string) {
  if (!isRealApiMode) return mockWechatLoginResult(true)
  return request<LoginResult, { phoneCode: string; code: string; clientType: string }>(
    '/auth/wechat-mini/phone-login',
    {
      method: 'POST',
      public: true,
      data: { phoneCode, code, clientType }
    }
  )
}
export async function redeemAssistanceCode(phone: string, code: string) { return request<LoginResult, { phone: string; code: string }>('/auth/assistance/redeem', { method: 'POST', public: true, data: { phone, code } }) }

export async function bindWechatPhone(input: BindWechatPhoneInput) {
  if (!isRealApiMode) {
    return {
      accessToken: mockUser.token,
      tokenType: 'Bearer',
      user: {
        ...mockUserInfo(true, true, false),
        phone: '13800000000'
      }
    } satisfies LoginResult
  }
  return request<LoginResult, BindWechatPhoneInput & { clientType: string }>(
    '/auth/wechat-mini/bind-phone',
    {
      method: 'POST',
      data: { ...input, clientType }
    }
  )
}

export async function logoutUser() {
  if (!isRealApiMode) return
  await request<{ status: string }>('/auth/logout', { method: 'POST' })
}

export type AssistanceCodeStatus = { id: number; status: string; expiresAt: string; createdAt: string; lastUsedAt?: string }
export async function createAssistanceCode(validDays = 1) { return request<{ code: string; expiresAt: string; validDays: number; replacedPrevious: boolean }, { validDays: number }>('/users/me/assistance-codes', { method: 'POST', data: { validDays } }) }
export async function listAssistanceCodes() { return request<AssistanceCodeStatus[]>('/users/me/assistance-codes') }
export async function revokeAssistanceCode(codeId: number) { return request<{ status: string }>(`/users/me/assistance-codes/${codeId}/revoke`, { method: 'POST' }) }

export async function cancelAccountByWechat(input: CancelAccountByWechatInput) {
  return request<{ status: string }, CancelAccountByWechatInput>(
    '/auth/cancel-account/wechat-reauth',
    { method: 'POST', data: input }
  )
}
