import { isRealApiMode, request } from '@/api/client'
import { mockUser } from '@/mock/data'
import type {
  BindPhoneInput,
  CancelAccountInput,
  ChangePhoneInput,
  LoginPhoneInput,
  LoginResult,
  RegisterPhoneInput,
  SendCodeResult,
  SendCodeScene,
  StatusResult,
  UserInfo
} from '@/types/api'

const clientType = 'WECHAT_MINI_PROGRAM'

function mockLoginResult(): LoginResult {
  const user: UserInfo = {
    id: 'mock_user',
    phone: mockUser.maskedPhone,
    phoneVerified: true,
    nickname: mockUser.nickname,
    status: mockUser.status,
    passwordSet: true
  }
  return { accessToken: mockUser.token, tokenType: 'Bearer', user }
}

function mockWechatPendingResult(): LoginResult {
  const user: UserInfo = {
    id: 'mock_user',
    phone: null,
    phoneVerified: false,
    nickname: mockUser.nickname,
    status: 'PENDING_BIND',
    passwordSet: false
  }
  return { accessToken: mockUser.token, tokenType: 'Bearer', user }
}

export async function sendCode(phone: string, scene: SendCodeScene) {
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

export async function wechatMiniLogin(code: string) {
  if (!isRealApiMode) return mockWechatPendingResult()
  return request<LoginResult, { code: string; clientType: string }>(
    '/auth/wechat-mini/login',
    {
      method: 'POST',
      public: true,
      data: { code, clientType }
    }
  )
}

export async function bindPhone(input: BindPhoneInput) {
  if (!isRealApiMode) return mockLoginResult()
  return request<LoginResult, BindPhoneInput>(
    '/auth/wechat-mini/bind-phone',
    {
      method: 'POST',
      data: input
    }
  )
}

export async function wechatPhoneLogin(phoneCode: string) {
  if (!isRealApiMode) return mockLoginResult()
  return request<LoginResult, { phoneCode: string; clientType: string }>(
    '/auth/wechat-mini/phone-login',
    {
      method: 'POST',
      public: true,
      data: { phoneCode, clientType }
    }
  )
}

export async function logoutUser() {
  if (!isRealApiMode) return
  await request<{ status: string }>('/auth/logout', { method: 'POST' })
}

export async function changePhone(input: ChangePhoneInput) {
  return request<StatusResult, ChangePhoneInput>('/auth/change-phone', { method: 'POST', data: input })
}

export async function cancelAccount(input: CancelAccountInput) {
  return request<{ status: string }, CancelAccountInput>('/auth/cancel-account', { method: 'POST', data: input })
}
