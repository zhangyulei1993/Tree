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

export interface DashboardStats {
  users: number
  families: number
  pendingPublicApplications: number
  pendingFounderTransfers: number
  pendingDissolutions: number
}

export interface ManagedUser {
  id: number
  phone?: string | null
  nickname?: string | null
  realName?: string | null
  accountOrigin: string
  registerClient: string
  phoneVerified: boolean
  phoneLoginEnabled: boolean
  hasWechatLogin: boolean
  canUnbindPhoneLogin: boolean
  trustTier: 'WECHAT_ONLY' | 'PHONE_BOUND'
  loginMethod: string
  status: string
  lastLoginAt?: string | null
  createdAt: string
}

export interface UserFamilyLink {
  familyId: number
  familyName: string
  memberId: number
  memberName: string
  familyRole: string
}

export interface ManagedUserDetail { user: ManagedUser; families: UserFamilyLink[] }

export interface ManagedFamily {
  id: number
  familyName: string
  familySurname: string
  nativePlace?: string | null
  regionText?: string | null
  description?: string | null
  status: string
  searchable: boolean
  publicDisplayStatus: string
  currentFounderMemberId?: number | null
  graphVersion: number
  memberCount: number
  createdAt: string
}

export interface ManagedFamilyMember {
  id: number
  familyId: number
  displayName: string
  gender: string
  status: string
  isLiving?: boolean | null
  userBindingPolicy: string
  boundUserId?: number | null
  familyRole?: string | null
}

export interface ManagedAdmin {
  id: number
  username: string
  displayName?: string | null
  phone?: string | null
  email?: string | null
  role: AdminRole
  status: string
  lockedUntil?: string | null
  lastLoginAt?: string | null
  createdAt: string
}

export interface OperationLogRecord {
  id: number
  operatorType: string
  operatorAdminId?: number
  operatorUserId?: number
  operatorRole?: string
  module: string
  action: string
  targetType?: string
  targetId?: number
  familyId?: number
  memberId?: number
  userId?: number
  detailJson?: unknown
  result: string
  errorMessage?: string
  ip?: string
  createdAt: string
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
  familyStatus?: string
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

export interface FinalizeFamilyResult {
  familyId: number
  status: string
  publicDisplayStatus: string
  searchable: boolean
  graphVersion: number
  dissolutionCompletedAt?: string
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

export interface PublicApplication {
  applicationId: number
  familyId: number
  familyName?: string
  familyPublicDisplayStatus?: string
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
  contentType: 'INTERNAL' | 'WECHAT_OFFICIAL'
  title: string
  slug: string
  summary?: string | null
  coverUrl?: string | null
  authorName?: string | null
  source?: string | null
  externalUrl?: string | null
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
  contentType: 'INTERNAL' | 'WECHAT_OFFICIAL'
  title: string
  slug: string
  summary?: string
  coverUrl?: string
  body: string
  externalUrl?: string
  authorName?: string
  source?: string
  status?: string
  isFeatured?: boolean
  sortOrder?: number
}

export interface AccountQuotaConfig {
  trustTier: 'WECHAT_ONLY' | 'PHONE_BOUND'
  maxOwnedFamilies: number
  maxMembersPerOwnedFamily: number
  maxJoinedFamilies: number
  updatedByAdminId?: number | null
  updatedAt: string
}

export interface UpdateAccountQuotaInput {
  maxOwnedFamilies: number
  maxMembersPerOwnedFamily: number
  maxJoinedFamilies: number
}

export interface AccountQuotaImpactPreview extends UpdateAccountQuotaInput {
  trustTier: string
  affectedUsers: number
  affectedFamilies: number
}

export interface AccountFeatureOverrideItem {
  id: number
  featureKey: string
  phoneMask: string
  updatedAt: string
}
