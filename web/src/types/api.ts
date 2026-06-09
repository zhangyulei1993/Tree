export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export interface UserInfo {
  id: number | string
  phone?: string | null
  phoneVerified: boolean
  nickname?: string | null
  status: string
}

export interface LoginResult {
  accessToken: string
  tokenType: string
  user: UserInfo
}

export interface SendCodeResult {
  expireSeconds: number
  cooldownSeconds: number
  devCode?: string
}

export interface LoginPhoneInput {
  phone: string
  password: string
  clientType: 'H5_WEB' | 'PC_WEB'
}

export interface RegisterPhoneInput extends LoginPhoneInput {
  code: string
  nickname?: string
}

export interface SendCodeInput {
  phone: string
  scene: 'REGISTER' | 'LOGIN'
  clientType: 'H5_WEB' | 'PC_WEB'
}

export interface FamilySummary {
  id: number | string
  familyName: string
  familySurname: string
  nativePlace?: string | null
  regionText?: string | null
  avatarUrl?: string | null
  status: string
  publicDisplayStatus: string
  role: string
}
