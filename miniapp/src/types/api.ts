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

export interface RegisterPhoneInput {
  phone: string
  code: string
  password: string
  nickname?: string
}

export interface LoginPhoneInput {
  phone: string
  password: string
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

export interface FamilyDetail extends FamilySummary {
  regionCode?: string | null
  description?: string | null
  searchable: boolean
  publicContactName?: string | null
  publicContactPhone?: string | null
  publicContactWechat?: string | null
  publicContactNote?: string | null
  publicContactVisible: boolean
  currentFounderMemberId?: number | null
  graphVersion: number
}

export interface PublicFamily {
  id: number | string
  familyName: string
  familySurname: string
  nativePlace?: string | null
  regionText?: string | null
  description?: string | null
  avatarUrl?: string | null
  publicContactName?: string | null
  publicContactPhone?: string | null
  publicContactWechat?: string | null
  publicContactNote?: string | null
  publicContactVisible: boolean
}

export interface PaginationQuery {
  page?: number
  pageSize?: number
}

export interface PaginatedResult<T> {
  items: T[]
  page: number
  pageSize: number
  total: number
}

export interface CreateVisitorMessageInput {
  visitorName?: string
  visitorPhone?: string
  visitorWechat?: string
  messageContent: string
}

export interface PublicVisitorMessage {
  messageId: number | string
  familyId: number | string
  visitorName?: string | null
  messageContent: string
  createdAt: string
  reviewedAt?: string | null
}

export interface VisitorMessageRecord extends PublicVisitorMessage {
  visitorPhone?: string | null
  visitorWechat?: string | null
  status: string
  updatedAt: string
}

export interface RejectInvitationInput {
  reason?: string
}

export interface Invitation {
  invitationId: number | string
  familyId: number | string
  familyName: string
  targetMemberId: number | string
  targetMemberName: string
  inviteChannel: string
  inviteMessage?: string | null
  familyRoleAfterAccept: string
  status: string
  expiredAt: string
  acceptedAt?: string | null
  rejectedAt?: string | null
  cancelledAt?: string | null
  createdAt: string
}

export interface CreateJoinRequestInput {
  applicantRealName?: string
  applicantMessage?: string
}

export interface CancelJoinRequestInput {
  cancelReason?: string
}

export interface JoinRequest {
  requestId: number | string
  familyId: number | string
  familyName?: string
  applicantUserId: number | string
  applicantRealName?: string | null
  applicantMessage?: string | null
  requestStatus: string
  handleComment?: string | null
  cancelledAt?: string | null
  createdAt: string
  updatedAt: string
}

export interface FamilyMember {
  memberId: number
  familyId: number
  name: string
  gender: string
  birthDate?: string | null
  birthYear?: number | null
  deathDate?: string | null
  deathYear?: number | null
  isAlive?: boolean | null
  avatarUrl?: string | null
  description?: string | null
  status: string
  userBindingPolicy: string
  boundUserId?: number | null
  boundFamilyRole?: string | null
  createdAt: string
  updatedAt: string
}

export interface TreeNode {
  memberId: number
  displayName: string
  surname?: string | null
  generationCharacter?: string | null
  gender: string
  memberType: string
  birthDate?: string | null
  deathDate?: string | null
  isLiving?: boolean | null
  userBindingState: string
  canExpand: boolean
  stopReason?: string | null
}

export interface TreeEdge {
  relationshipId: number
  fromMemberId: number
  toMemberId: number
  relationshipType: 'PARENT_CHILD' | 'SPOUSE'
  parentLinkType?: string | null
  relationNoteType?: string | null
  relationNote?: string | null
}

export interface TreeItem {
  memberId: number
  parentIds: number[]
  childrenIds: number[]
  spouseIds: number[]
}

export interface FamilyTreeResult {
  familyId: number
  treeMode: string
  graphVersion: number
  nodes: TreeNode[]
  edges: TreeEdge[]
  tree: TreeItem[]
}
