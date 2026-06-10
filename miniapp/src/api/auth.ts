import { isRealApiMode, request } from '@/api/client'
import { mockUser } from '@/mock/data'
import type {
  LoginPhoneInput,
  LoginResult,
  RegisterPhoneInput,
  SendCodeResult,
  UserInfo
} from '@/types/api'

const clientType = 'WECHAT_MINI_PROGRAM'

function mockLoginResult(): LoginResult {
  const user: UserInfo = {
    id: 'mock_user',
    phone: mockUser.maskedPhone,
    phoneVerified: true,
    nickname: mockUser.nickname,
    status: mockUser.status
  }
  return { accessToken: mockUser.token, tokenType: 'Bearer', user }
}

export async function sendCode(phone: string, scene: 'REGISTER' | 'LOGIN') {
  if (!isRealApiMode) {
    return { expireSeconds: 300, cooldownSeconds: 60 } satisfies SendCodeResult
  }
  return request<SendCodeResult, { phone: string; scene: string; clientType: string }>(
    '/auth/send-code',
    {
      method: 'POST',
      public: true,
      data: { phone, scene, clientType }
    }
  )
}

export async function registerPhone(input: RegisterPhoneInput) {
  if (!isRealApiMode) return mockLoginResult()
  return request<LoginResult, RegisterPhoneInput & { clientType: string }>(
    '/auth/register-phone',
    {
      method: 'POST',
      public: true,
      data: { ...input, clientType }
    }
  )
}

export async function loginPhone(input: LoginPhoneInput) {
  if (!isRealApiMode) return mockLoginResult()
  return request<LoginResult, LoginPhoneInput & { clientType: string }>(
    '/auth/login-phone',
    {
      method: 'POST',
      public: true,
      data: { ...input, clientType }
    }
  )
}

export async function logoutUser() {
  if (!isRealApiMode) return
  await request<{ status: string }>('/auth/logout', { method: 'POST' })
}
