import type { MemberType, RelationshipAddType } from '@/types/api'

export const parentRoleLabels = ['本家成员', '本家成员的配偶'] as const
export const parentRoleValues: MemberType[] = ['LINEAGE_MEMBER', 'SPOUSE']

export function isParentAddType(addType: RelationshipAddType): boolean {
  return addType === 'ADD_FATHER' || addType === 'ADD_MOTHER'
}

export function parentRoleHint(memberType: MemberType): string {
  if (memberType === 'SPOUSE') {
    return '选择「本家成员的配偶」时，基准成员必须已有相反性别的本家成员父母。'
  }
  return ''
}

export interface JoinLocationPlacement {
  parentLinkType?: string
  relationNoteType?: string
  relationNote?: string
}

export function buildApproveJoinLocation(input: {
  baseMemberId: number | string
  addType: RelationshipAddType
  parentMemberType?: MemberType
  placement?: JoinLocationPlacement
}) {
  const { baseMemberId, addType, parentMemberType, placement } = input
  const relationship = addType === 'ADD_SPOUSE'
    ? {
        relationshipType: 'SPOUSE' as const,
        relationNoteType: placement?.relationNoteType,
        relationNote: placement?.relationNote
      }
    : {
        relationshipType: 'PARENT_CHILD' as const,
        parentLinkType: placement?.parentLinkType || 'PRIMARY',
        relationNoteType: placement?.relationNoteType,
        relationNote: placement?.relationNote
      }

  const location: {
    baseMemberId: number | string
    addType: RelationshipAddType
    memberType?: MemberType
    relationship: typeof relationship
  } = { baseMemberId, addType, relationship }

  if (isParentAddType(addType)) {
    location.memberType = parentMemberType ?? 'LINEAGE_MEMBER'
  }

  return location
}
