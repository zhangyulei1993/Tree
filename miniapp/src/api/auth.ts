import { isRealApiMode, request } from '@/api/client'
import { mockUser } from '@/mock/data'
import type {
  CancelAccountByWechatInput,
  LoginResult,
  UserInfo
} from '@/types/api'

const clientType = 'WECHAT_MINI_PROGRAM'

function mockWechatLoginResult(withProfile = true): LoginResult {
  const user: UserInfo = {
    id: 'mock_user',
    phone: null,
    phoneVerified: false,
    nickname: withProfile ? mockUser.nickname : null,
    status: 'ACTIVE',
    passwordSet: false
  }
  return { accessToken: mockUser.token, tokenType: 'Bearer', user }
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
