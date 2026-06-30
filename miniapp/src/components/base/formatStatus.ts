export function familyStatusText(status?: string): string {
  switch (status) {
    case 'NORMAL':
      return '正常'
    case 'DISABLED':
      return '已停用'
    case 'DISSOLVED':
      return '已解散'
    case 'DISSOLUTION_PENDING':
      return '解散待审核'
    default:
      return '未知'
  }
}

export function roleText(role?: string): string {
  switch (role) {
    case 'FOUNDER':
      return '家庭创建者'
    case 'FAMILY_ADMIN':
      return '家庭管理员'
    case 'MEMBER':
      return '家庭成员'
    case 'ROOT_ADMIN':
      return '平台管理员'
    default:
      return '未知'
  }
}

export function publicStatusText(status?: string): string {
  switch (status) {
    case 'APPROVED':
      return '已通过'
    case 'PENDING':
      return '待审核'
    case 'REJECTED':
      return '已驳回'
    case 'PRIVATE':
      return '私密'
    case 'TAKEN_DOWN':
      return '已下架'
    default:
      return '未知'
  }
}

export function memberStatusText(status?: string): string {
  switch (status) {
    case 'ACTIVE':
      return '正常'
    case 'DELETED':
      return '已删除'
    default:
      return '未知'
  }
}

export function genderText(gender?: string): string {
  switch (gender) {
    case 'MALE':
    case 'male':
      return '男'
    case 'FEMALE':
    case 'female':
      return '女'
    default:
      return '未知'
  }
}

export function invitationStatusText(status?: string): string {
  switch (status) {
    case 'PENDING':
      return '待处理'
    case 'ACCEPTED':
      return '已接受'
    case 'REJECTED':
      return '已拒绝'
    case 'EXPIRED':
      return '已过期'
    case 'CANCELLED':
      return '已取消'
    default:
      return '未知'
  }
}

export function joinRequestStatusText(status?: string): string {
  switch (status) {
    case 'PENDING':
      return '审核中'
    case 'APPROVED':
      return '已通过'
    case 'REJECTED':
      return '已驳回'
    case 'CANCELLED':
      return '已取消'
    default:
      return '未知'
  }
}

export function accountStatusText(status?: string): string {
  switch (status) {
    case 'ACTIVE':
      return '正常'
    case 'DISABLED':
      return '已停用'
    case 'CANCELLED':
      return '已注销'
    case 'PENDING_PHONE_BIND':
      return '待绑定手机号'
    default:
      return '未知'
  }
}

export function relationTypeText(type?: string): string {
  switch (type) {
    case 'PARENT_CHILD':
      return '父母子女'
    case 'SPOUSE':
      return '配偶'
    default:
      return '亲属关系'
  }
}

export function bindingPolicyText(policy?: string): string {
  switch (policy) {
    case 'REQUIRED':
      return '需要绑定'
    case 'OPTIONAL':
      return '可选绑定'
    case 'NOT_REQUIRED':
      return '无需绑定'
    default:
      return '未知'
  }
}

export function statusTagTone(
  status?: string
): 'active' | 'pending' | 'danger' | 'muted' {
  const key = status?.toUpperCase() || ''
  if (['ACTIVE', 'NORMAL', 'APPROVED', 'ACCEPTED', 'PUBLISHED'].includes(key)) return 'active'
  if (['PENDING', 'PENDING_PHONE_BIND', 'DISSOLUTION_PENDING', 'DRAFT'].includes(key)) return 'pending'
  if (['DISABLED', 'REJECTED', 'CANCELLED', 'DELETED', 'DISSOLVED', 'EXPIRED', 'TAKEN_DOWN'].includes(key))
    return 'danger'
  return 'muted'
}
