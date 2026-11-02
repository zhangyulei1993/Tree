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
  avatarUrl?: string | null
  status: string
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

export interface BindWechatPhoneInput {
  phoneCode: string
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
  grayAccessEnabled?: boolean
  usage: {
    ownedFamilies: number
    joinedFamilies: number
    membersPerOwnedFamily: Record<string, number>
    membersLimitsByFamily?: Record<string, number>
  }
}

export interface FamilySummary {
  id: number | string
  familyName: string
  familySurname: string
  nativePlace?: string | null
  regionCode?: string | null
  regionText?: string | null
  avatarUrl?: string | null
  status: string
  dissolutionCooldownUntil?: string | null
  dissolutionCooldownDays: number
  publicDisplayStatus: string
  publicDisplayEnabled: boolean
  role: string
}

export interface FamilyDetail extends FamilySummary {
  regionCode?: string | null
  description?: string | null
  dissolutionCompletedAt?: string | null
  searchable: boolean
  publicContactName?: string | null
  publicContactPhone?: string | null
  publicContactWechat?: string | null
  publicContactNote?: string | null
  publicContactVisible: boolean
  currentFounderMemberId?: number | null
  graphVersion: number
	  revision?: number
  memberCount?: number
  memberLimit?: number
}

export interface LeaveFamilyResult {
  familyId: number | string
  memberId: number | string
  status: 'LEFT'
}

export interface CreateFamilyInput {
  surname: string
  founderName: string
  founderGender: Gender
  founderBirthYear: number
  familyName?: string
  nativePlace?: string
  regionCode?: string
  regionText?: string
  description?: string
}

export interface UpdateFamilyInput {
	  expectedRevision?: number
  familyName?: string
  nativePlace?: string
  regionCode?: string
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

export interface PublicPlatformStats {
  familyCount: number
  memberCount: number
  relationshipCount: number
}

export interface ShowcaseFamilyItem {
  id: number
  displayName: string
  regionText: string
  summary: string
  coverImageUrl: string
  treeImageUrl: string
  memberCount: number
  relationshipCount: number
  rotationPinned: boolean
}

export interface ShowcaseFamilyTree {
  showcaseFamilyId: number
  nodes: Array<{
    memberId: number
    displayName: string
    identityText: string
    memberType: string
    gender: string
    birthYear?: number | null
  }>
  edges: TreeEdge[]
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
  inviteType?: string
  inviteChannel: string
  inviteMessage?: string | null
  familyRoleAfterAccept: string
  status: string
  expiredAt: string
  acceptedAt?: string | null
  acceptedByUserId?: number | string | null
  acceptedByDisplayName?: string
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
	portraitHairStyle?: PortraitHairStyle
	portraitGlassesStyle?: PortraitGlassesStyle
	portraitAvatarStyle?: PortraitAvatarStyle
  birthDate?: string | null
  birthYear?: number | null
  deathDate?: string | null
  deathYear?: number | null
  isAlive?: boolean | null
  avatarUrl?: string | null
  description?: string | null
  memorialVisible?: boolean
  status: string
  userBindingPolicy: string
	  revision?: number
	  profileSourceType?: string
	  profileVerificationStatus?: string
	  profileConfirmedAt?: string | null
	  profileConfirmedByUserId?: number | null
  boundUserId?: number | null
  boundFamilyRole?: string | null
  createdAt: string
  updatedAt: string
}

export type Gender = 'MALE' | 'FEMALE' | 'UNKNOWN'
export type PortraitHairStyle = 'AUTO' | 'SHORT' | 'BOB' | 'BRAIDS' | 'LOW_BUN' | 'WAVY' | 'LONG'
export type PortraitGlassesStyle = 'AUTO' | 'ON' | 'OFF'
export type PortraitAvatarStyle = 'AUTO' | `BASE_${number}` | `GLASSES_${number}`
export type UserBindingPolicy = 'OPTIONAL' | 'REQUIRED' | 'NOT_REQUIRED'

export interface CreateMemberInput {
  name: string
  gender?: Gender
	portraitHairStyle?: PortraitHairStyle
	portraitGlassesStyle?: PortraitGlassesStyle
	portraitAvatarStyle?: PortraitAvatarStyle
  birthDate?: string
  birthYear?: number
  deathDate?: string
  deathYear?: number
  isAlive?: boolean
  avatarUrl?: string
  description?: string
  memorialVisible?: boolean
  userBindingPolicy?: UserBindingPolicy
}

export type UpdateMemberInput = Partial<CreateMemberInput> & { expectedRevision?: number }

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
  coParentMemberId?: number
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
  coParentMemberId?: number
  memberId: number
  addType: RelationshipAddType
  memberType?: MemberType
  relationship: CreateRelationshipInput['relationship']
}

export interface UpdateRelationshipInput {
	  expectedRevision?: number
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
	  revision?: number
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

export interface QualityEligibility {
  eligible: boolean
  memberCount: number
  relationshipCount: number
  boundMemberCount: number
  profileCompleteCount: number
  unknownGenderCount: number
  profileComplete: boolean
  requirements: string[]
  checks: QualityEligibilityCheck[]
}

export interface QualityEligibilityCheck {
  key: string
  label: string
  current: string
  required: string
  met: boolean
}

export interface FamilyPublicStatus {
  familyId: number | string
  publicDisplayStatus: string
  publicDisplayEnabled: boolean
  publicAppliedAt?: string | null
  publicApprovedAt?: string | null
  publicEnabledAt?: string | null
  publicTakenDownAt?: string | null
}

export interface FamilyOperationLog {
  id: number | string
  operatorType: string
  operatorAdminId?: number | string | null
  operatorUserId?: number | string | null
  module: string
  action: string
  targetType?: string | null
  targetId?: number | string | null
  memberId?: number | string | null
  userId?: number | string | null
  result: string
  errorMessage?: string | null
  createdAt: string
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
  identityText?: string | null
  surname?: string | null
  generationCharacter?: string | null
  gender: string
	portraitHairStyle?: PortraitHairStyle
	portraitGlassesStyle?: PortraitGlassesStyle
	portraitAvatarStyle?: PortraitAvatarStyle
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
	  revision?: number
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
  grayAccessEnabled?: boolean
  nodes: TreeNode[]
  edges: TreeEdge[]
  tree: TreeItem[]
}
