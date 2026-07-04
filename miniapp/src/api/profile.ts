import { apiUrl, isRealApiMode, request, sessionTokenKey } from '@/api/client'
import { mockUser } from '@/mock/data'
import type { ApiResponse, UpdateProfileInput, UserInfo } from '@/types/api'

function mockUserInfo(): UserInfo {
  return {
    id: 'mock_user',
    phone: null,
    phoneVerified: false,
    phoneLoginEnabled: false,
    nickname: mockUser.nickname,
    avatarUrl: null,
    status: mockUser.status,
    passwordSet: false
  }
}

export async function getMe() {
  if (!isRealApiMode) return mockUserInfo()
  return request<UserInfo>('/users/me')
}

export async function updateProfile(input: UpdateProfileInput) {
  if (!isRealApiMode) {
    return {
      ...mockUserInfo(),
      nickname: input.nickname ?? mockUser.nickname
    } satisfies UserInfo
  }
  return request<UserInfo, UpdateProfileInput>('/users/me/profile', {
    method: 'PATCH',
    data: input
  })
}

export async function uploadAvatar(filePath: string) {
  if (!isRealApiMode) {
    return {
      ...mockUserInfo(),
      avatarUrl: filePath
    } satisfies UserInfo
  }

  const token = uni.getStorageSync(sessionTokenKey) as string
  return new Promise<UserInfo>((resolve, reject) => {
    uni.uploadFile({
      url: apiUrl('/users/me/avatar'),
      filePath,
      name: 'file',
      header: token ? { Authorization: `Bearer ${token}` } : {},
      success(response) {
        if (response.statusCode === 401) {
          reject(new Error('登录状态已失效'))
          return
        }
        if (response.statusCode < 200 || response.statusCode >= 300) {
          reject(new Error(`头像上传失败（HTTP ${response.statusCode}）`))
          return
        }
        try {
          const payload = JSON.parse(response.data) as ApiResponse<UserInfo>
          if (!payload || payload.code !== 0) {
            reject(new Error(payload?.message || '头像上传失败'))
            return
          }
          resolve(payload.data)
        } catch {
          reject(new Error('头像上传响应解析失败'))
        }
      },
      fail(error) {
        const errMsg = error.errMsg || ''
        if (/url not in domain list/i.test(errMsg)) {
          reject(new Error('头像上传域名未配置。请在微信公众平台为小程序添加 uploadFile 合法域名：https://tapi.bigbigboy.cn'))
          return
        }
        reject(new Error(errMsg || '头像上传失败'))
      }
    })
  })
}
