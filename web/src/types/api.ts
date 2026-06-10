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

export interface PublicContactInput {
  name?: string
  phone?: string
  wechat?: string
  note?: string
  visible?: boolean
}

export interface CreateFamilyInput {
  surname: string
  familyName?: string
  nativePlace?: string
  regionText?: string
  description?: string
  publicContact?: PublicContactInput
}

export interface UpdateFamilyInput {
  familyName?: string
  nativePlace?: string
  regionCode?: string
  regionText?: string
  description?: string
  avatarUrl?: string
  searchable?: boolean
  publicContactName?: string
  publicContactPhone?: string
  publicContactWechat?: string
  publicContactNote?: string
  publicContactVisible?: boolean
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

export interface SubmitPublicApplicationInput {
  applicationReason?: string
}

export interface CancelPublicApplicationInput {
  cancelReason?: string
}

export interface PublicApplication {
  applicationId: number
  familyId: number
  familyName?: string
  applicantUserId?: number
  applicantAdminId?: number
  status: string
  reason?: string | null
  reviewResult?: string | null
  reviewedByAdminId?: number | null
  reviewedAt?: string | null
  reviewComment?: string | null
  cancelledAt?: string | null
  cancelReason?: string | null
  createdAt: string
  updatedAt: string
}

export interface PublicApplicationQuery extends PaginationQuery {
  status?: string
}

export interface FamilyPublicStatus {
  familyId: number
  publicDisplayStatus: string
  publicAppliedAt?: string | null
  publicApprovedAt?: string | null
  publicTakenDownAt?: string | null
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

export type InviteChannel = 'SHARE_LINK' | 'IN_APP'

export interface CreateInvitationInput {
  inviteChannel: InviteChannel
  targetUserId?: number
  inviteMessage?: string
  familyRoleAfterAccept?: 'MEMBER'
}

export interface InvitationActionInput {
  reason?: string
}

export interface Invitation {
  invitationId: number | string
  familyId: number | string
  familyName: string
  targetMemberId: number | string
  targetMemberName: string
  inviteChannel: InviteChannel | string
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
  approveMode?: string | null
  boundMemberId?: number | null
  createdMemberId?: number | null
  handleComment?: string | null
  handledAt?: string | null
  cancelledAt?: string | null
  createdAt: string
  updatedAt: string
  graphVersion?: number | null
}

export type Gender = 'MALE' | 'FEMALE' | 'UNKNOWN'
export type UserBindingPolicy = 'OPTIONAL' | 'REQUIRED' | 'NOT_REQUIRED'
export type FamilyRole = 'FOUNDER' | 'FAMILY_ADMIN' | 'MEMBER'

export interface FamilyMember {
  memberId: number
  familyId: number
  name: string
  gender: Gender | string
  birthDate?: string | null
  birthYear?: number | null
  deathDate?: string | null
  deathYear?: number | null
  isAlive?: boolean | null
  avatarUrl?: string | null
  description?: string | null
  status: string
  userBindingPolicy: UserBindingPolicy | string
  boundUserId?: number | null
  boundFamilyRole?: FamilyRole | string | null
  createdAt: string
  updatedAt: string
}

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

export interface NewRelationshipMemberInput {
  name: string
  gender?: Gender
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

export interface UpdateRelationshipInput {
  parentLinkType?: string
  relationNoteType?: string
  relationNote?: string
}

export interface CreatedRelationshipMember {
  memberId: number
  familyId: number
  name: string
  gender: string
  status: string
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
  createdMember?: CreatedRelationshipMember
  relationships: Relationship[]
  graphVersion: number
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
  relationshipType: RelationshipType
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
