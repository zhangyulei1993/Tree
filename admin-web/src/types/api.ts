declare global {
  interface ImportMetaEnv {
    readonly VITE_API_MODE?: 'mock' | 'real'
    readonly VITE_API_BASE_URL?: string
  }

  interface ImportMeta {
    readonly env: ImportMetaEnv
  }
}

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export interface PageResult<T> {
  items: T[]
  page: number
  pageSize: number
  total: number
}

export type AdminRole = 'ROOT_ADMIN' | 'SUPER_ADMIN' | 'PLATFORM_ADMIN'

export interface AdminInfo {
  id: number
  username: string
  displayName?: string | null
  phone?: string | null
  email?: string | null
  role: AdminRole
  status: string
}

export interface AdminLoginResult {
  accessToken: string
  tokenType: string
  admin: AdminInfo
}

export interface ReviewRequest {
  reviewComment?: string
}

export interface DeleteVisitorMessageRequest {
  deleteReason?: string
}

export interface PublicApplication {
  applicationId: number
  familyId: number
  familyName?: string
  applicantUserId?: number
  applicantAdminId?: number
  status: string
  reason?: string
  reviewResult?: string
  reviewedByAdminId?: number
  reviewedAt?: string
  reviewComment?: string
  cancelledAt?: string
  cancelReason?: string
  createdAt: string
  updatedAt: string
}

export interface VisitorMessage {
  messageId: number
  familyId: number
  familyName?: string
  visitorName?: string
  visitorPhone?: string
  visitorWechat?: string
  messageContent: string
  status: string
  reviewedByAdminId?: number
  reviewedAt?: string
  reviewComment?: string
  deletedAt?: string
  createdAt: string
  updatedAt: string
}
