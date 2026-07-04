export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export interface UserInfo {
  id: number | string
  phone?: string | null
  phoneVerified: boolean
  phoneLoginEnabled?: boolean
  nickname?: string | null
  avatarUrl?: string | null
  status: string
  passwordSet?: boolean
}

export interface UpdateProfileInput {
  nickname?: string
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

export interface BindPhoneCredentialInput {
  phone: string
  password: string
  confirmPassword: string
}

export interface ChangePhoneLoginPasswordInput {
  currentPassword: string
  newPassword: string
}

export interface BindPhoneInput {
  phone: string
  code: string
  password: string
}

export type SendCodeScene = 'REGISTER' | 'LOGIN' | 'BIND_PHONE' | 'CHANGE_PHONE_OLD' | 'CHANGE_PHONE_NEW' | 'CANCEL_ACCOUNT'

export interface ChangePhoneInput {
  oldPhoneCode: string
  newPhone: string
  newPhoneCode: string
}

export interface StatusResult {
  status: string
}

export interface CancelAccountByWechatInput {
  code: string
  cancelReason?: string
}

export interface CancelAccountInput {
  phoneCode: string
  cancelReason?: string
}

export interface UserCapabilities {
  trustTier: 'WECHAT_ONLY' | 'PHONE_BOUND'
  limits: {
    maxOwnedFamilies: number
    maxMembersPerOwnedFamily: number
    maxJoinedFamilies: number
  }
  usage: {
    ownedFamilies: number
    joinedFamilies: number
    membersPerOwnedFamily: Record<string, number>
  }
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

export interface LeaveFamilyResult {
  familyId: number | string
  memberId: number | string
  status: 'LEFT'
}

export interface CreateFamilyInput {
  surname: string
  founderGender: Gender
  familyName?: string
  nativePlace?: string
  regionText?: string
  description?: string
}

export interface UpdateFamilyInput {
  familyName?: string
  nativePlace?: string
  regionText?: string
  description?: string
  searchable?: boolean
  publicContactName?: string
  publicContactPhone?: string
  publicContactWechat?: string
  publicContactNote?: string
  publicContactVisible?: boolean
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

export interface PublicFamilyShowcaseItem {
  id: number | string
  familyName: string
  familySurname: string
  nativePlace?: string | null
  regionText?: string | null
  description?: string | null
  avatarUrl?: string | null
  publicApprovedAt?: string | null
  createdAt?: string
  updatedAt?: string
}

export interface ListPublicFamilyShowcaseQuery extends PaginationQuery {}

export interface PublicFamilyListItem extends PublicFamily {
  publicApprovedAt?: string | null
  createdAt?: string
  updatedAt?: string
}

export interface ListPublicFamiliesQuery extends PaginationQuery {
  keyword?: string
  familySurname?: string
  regionText?: string
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

export interface RejectInvitationInput {
  reason?: string
}

export interface CreateInvitationInput {
  inviteChannel: 'SHARE_LINK'
  inviteMessage?: string
  familyRoleAfterAccept: 'MEMBER'
}

export interface Invitation {
  invitationId: number | string
  familyId: number | string
  familyName: string
  targetMemberId: number | string
  targetMemberName: string
  inviterDisplayName?: string
  inviterRole?: string
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

export interface CreatedInvitation {
  invitation: Invitation
  inviteToken: string
}

export interface CreateJoinRequestInput {
  applicantRealName?: string
  applicantGender: Gender
  applicantMessage?: string
}

export interface ApproveJoinRequestInput {
  approveMode: 'BIND_EXISTING_MEMBER' | 'CREATE_NEW_MEMBER'
  memberId?: number | string
  newMember?: {
    name: string
    gender?: Gender
    birthDate?: string
    birthYear?: number
    isAlive?: boolean
    userBindingPolicy?: UserBindingPolicy
  }
  location?: {
    baseMemberId: number | string
    addType: RelationshipAddType
    memberType?: MemberType
    relationship: {
      relationshipType?: RelationshipType
      parentLinkType?: string
      relationNoteType?: string
      relationNote?: string
    }
  }
  handleComment?: string
}

export interface RejectJoinRequestInput {
  handleComment?: string
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
  applicantGender?: Gender | null
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

export type Gender = 'MALE' | 'FEMALE' | 'UNKNOWN'
export type UserBindingPolicy = 'OPTIONAL' | 'REQUIRED' | 'NOT_REQUIRED'

export interface CreateMemberInput {
  name: string
  gender?: Gender
  birthDate?: string
  birthYear?: number
  deathDate?: string
  deathYear?: number
  isAlive?: boolean
  avatarUrl?: string
  description?: string
  userBindingPolicy?: UserBindingPolicy
}

export type UpdateMemberInput = Partial<CreateMemberInput>

export type RelationshipAddType =
  | 'ADD_FATHER'
  | 'ADD_MOTHER'
  | 'ADD_CHILD'
  | 'ADD_SPOUSE'
  | 'ADD_SIBLING'

export type RelationshipType = 'PARENT_CHILD' | 'SPOUSE'

export type MemberType = 'LINEAGE_MEMBER' | 'SPOUSE' | 'EXTERNAL_MEMBER'

export interface NewRelationshipMemberInput {
  name: string
  gender?: Gender
  memberType?: MemberType
  birthDate?: string
  birthYear?: number
  deathDate?: string
  deathYear?: number
  isAlive?: boolean
  userBindingPolicy?: UserBindingPolicy
}

export interface CreateRelationshipInput {
  baseMemberId: number
  addType: RelationshipAddType
  newMember: NewRelationshipMemberInput
  relationship: {
    relationshipType?: RelationshipType
    parentLinkType?: string
    relationNoteType?: string
    relationNote?: string
  }
}

export interface PlaceExistingMemberInput {
  baseMemberId: number
  memberId: number
  addType: RelationshipAddType
  memberType?: MemberType
  relationship: CreateRelationshipInput['relationship']
}

export interface UpdateRelationshipInput {
  parentLinkType?: string
  relationNoteType?: string
  relationNote?: string
}

export interface Relationship {
  relationshipId: number
  familyId: number
  fromMemberId: number
  toMemberId: number
  relationshipType: RelationshipType
  parentLinkType?: string | null
  relationNoteType?: string | null
  relationNote?: string | null
  status: string
  createdAt: string
  updatedAt: string
  deletedAt?: string | null
}

export interface RelationshipMutationResult {
  createdMember?: {
    memberId: number
    familyId: number
    name: string
    gender: string
    status: string
  }
  relationships: Relationship[]
  graphVersion: number
}

export interface PublicApplication {
  applicationId: number | string
  familyId: number | string
  familyName?: string
  familyPublicDisplayStatus?: string
  status: string
  reason?: string | null
  reviewResult?: string | null
  reviewComment?: string | null
  cancelledAt?: string | null
  cancelReason?: string | null
  createdAt: string
  updatedAt: string
}

export interface FamilyPublicStatus {
  familyId: number | string
  publicDisplayStatus: string
  publicAppliedAt?: string | null
  publicApprovedAt?: string | null
  publicTakenDownAt?: string | null
}

export interface RoleChangeResult {
  familyId: number | string
  memberId: number | string
  userId: number | string
  familyRole: string
  graphVersion: number
  updatedAt: string
}

export interface FounderTransferRequest {
  requestId: number | string
  familyId: number | string
  fromMemberId: number | string
  fromUserId: number | string
  toMemberId: number | string
  toUserId: number | string
  requestStatus: string
  requestReason?: string | null
  reviewResult?: string | null
  reviewComment?: string | null
  cancelledAt?: string | null
  cancelReason?: string | null
  createdAt: string
  updatedAt: string
}

export interface DissolutionRequest {
  requestId: number | string
  familyId: number | string
  requesterMemberId: number | string
  requesterUserId: number | string
  requestStatus: string
  requestReason?: string | null
  reviewResult?: string | null
  reviewComment?: string | null
  cancelledAt?: string | null
  cancelReason?: string | null
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
