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

export interface FounderTransferRequest {
  requestId: number
  familyId: number
  familyName?: string
  fromMemberId: number
  fromUserId: number
  toMemberId: number
  toUserId: number
  requestStatus: string
  requestReason?: string
  reviewResult?: string
  reviewedByAdminId?: number
  reviewedAt?: string
  reviewComment?: string
  completedAt?: string
  cancelledAt?: string
  cancelReason?: string
  createdAt: string
  updatedAt: string
}

export interface DissolutionRequest {
  requestId: number
  familyId: number
  familyName?: string
  requesterMemberId: number
  requesterUserId: number
  requestStatus: string
  requestReason?: string
  reviewResult?: string
  reviewedByAdminId?: number
  reviewedAt?: string
  reviewComment?: string
  completedAt?: string
  cancelledAt?: string
  cancelReason?: string
  createdAt: string
  updatedAt: string
}

export interface RestoreFamilyRequest {
  searchable?: boolean
}

export interface RestoreFamilyResult {
  familyId: number
  status: string
  publicDisplayStatus: string
  searchable: boolean
  graphVersion: number
  restoredAt?: string
}

export interface TakeDownPublicFamilyRequest {
  reason?: string
}

export interface FamilyPublicStatus {
  familyId: number
  publicDisplayStatus: string
  publicAppliedAt?: string
  publicApprovedAt?: string
  publicTakenDownAt?: string
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

export interface ContentCategory {
  id: number
  key: string
  name: string
  description?: string | null
  sortOrder: number
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface ContentArticleSummary {
  id: number
  categoryId: number
  categoryKey: string
  categoryName: string
  title: string
  slug: string
  summary?: string | null
  coverUrl?: string | null
  authorName?: string | null
  source?: string | null
  status: string
  isFeatured: boolean
  sortOrder: number
  publishedAt?: string | null
  createdAt: string
  updatedAt: string
}

export interface ContentArticleDetail extends ContentArticleSummary {
  body: string
}

export interface ContentCategoryInput {
  key?: string
  name: string
  description?: string
  sortOrder?: number
  isActive?: boolean
}

export interface ContentArticleInput {
  categoryKey: string
  title: string
  slug: string
  summary?: string
  coverUrl?: string
  body: string
  authorName?: string
  source?: string
  status?: string
  isFeatured?: boolean
  sortOrder?: number
}
