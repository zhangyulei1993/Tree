import { isRealApiMode, request } from '@/api/client'
import { mockUser } from '@/mock/data'
import type {
  BindPhoneCredentialInput,
  CancelAccountByWechatInput,
  ChangePhoneLoginPasswordInput,
  LoginPhoneInput,
  LoginResult,
  UserInfo
} from '@/types/api'

const clientType = 'WECHAT_MINI_PROGRAM'

function mockUserInfo(withProfile = true, phoneLoginEnabled = false): UserInfo {
  return {
    id: 'mock_user',
    phone: phoneLoginEnabled ? '13800000000' : null,
    phoneVerified: phoneLoginEnabled,
    phoneLoginEnabled,
    nickname: withProfile ? mockUser.nickname : null,
    status: 'ACTIVE',
    passwordSet: phoneLoginEnabled
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

export async function loginPhone(input: LoginPhoneInput) {
  if (!isRealApiMode) {
    return {
      accessToken: mockUser.token,
      tokenType: 'Bearer',
      user: {
        ...mockUserInfo(true, true),
        phone: '13800000000'
      }
    } satisfies LoginResult
  }
  return request<LoginResult, LoginPhoneInput & { clientType: string }>(
    '/auth/login-phone',
    {
      method: 'POST',
      public: true,
      data: { ...input, clientType }
    }
  )
}

export async function bindPhoneCredential(input: BindPhoneCredentialInput) {
  if (!isRealApiMode) {
    return {
      accessToken: mockUser.token,
      tokenType: 'Bearer',
      user: {
        ...mockUserInfo(true, true),
        phone: '13800000000'
      }
    } satisfies LoginResult
  }
  return request<LoginResult, BindPhoneCredentialInput>(
    '/auth/wechat-mini/bind-phone-credential',
    { method: 'POST', data: input }
  )
}

export async function changePhoneLoginPassword(input: ChangePhoneLoginPasswordInput) {
  if (!isRealApiMode) {
    return {
      accessToken: mockUser.token,
      tokenType: 'Bearer',
      user: mockUserInfo(true, true)
    } satisfies LoginResult
  }
  return request<LoginResult, ChangePhoneLoginPasswordInput>(
    '/auth/wechat-mini/change-phone-login-password',
    { method: 'POST', data: input }
  )
}

export async function logoutUser() {
  if (!isRealApiMode) return
  await request<{ status: string }>('/auth/logout', { method: 'POST' })
}

export async function cancelAccountByWechat(input: CancelAccountByWechatInput) {
  return request<{ status: string }, CancelAccountByWechatInput>(
    '/auth/cancel-account/wechat-reauth',
    { method: 'POST', data: input }
  )
}
