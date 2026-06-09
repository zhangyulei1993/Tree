import { apiClient, isRealApiMode } from '@/api/client'
import { currentUser } from '@/mock/data'
import type {
  ApiResponse,
  LoginPhoneInput,
  LoginResult,
  RegisterPhoneInput,
  SendCodeInput,
  SendCodeResult,
  UserInfo
} from '@/types/api'

function mockUser(): UserInfo {
  return {
    id: currentUser.id,
    phone: currentUser.maskedPhone,
    phoneVerified: currentUser.phoneVerified,
    nickname: currentUser.nickname,
    status: currentUser.status
  }
}

export async function sendCode(input: SendCodeInput): Promise<SendCodeResult> {
  if (!isRealApiMode) {
    return { expireSeconds: 300, cooldownSeconds: 60 }
  }
  const response = await apiClient.post<ApiResponse<SendCodeResult>>('/auth/send-code', input)
  return response.data.data
}

export async function registerPhone(input: RegisterPhoneInput): Promise<LoginResult> {
  if (!isRealApiMode) {
    return { accessToken: currentUser.token, tokenType: 'Bearer', user: mockUser() }
  }
  const response = await apiClient.post<ApiResponse<LoginResult>>('/auth/register-phone', input)
  return response.data.data
}

export async function loginPhone(input: LoginPhoneInput): Promise<LoginResult> {
  if (!isRealApiMode) {
    return { accessToken: currentUser.token, tokenType: 'Bearer', user: mockUser() }
  }
  const response = await apiClient.post<ApiResponse<LoginResult>>('/auth/login-phone', input)
  return response.data.data
}

export async function logoutUser(): Promise<void> {
  if (!isRealApiMode) return
  await apiClient.post<ApiResponse<{ status: string }>>('/auth/logout')
}
