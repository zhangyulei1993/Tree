import type { FamilyMember } from '@/types/api'

function parseISODate(value: string): Date | null {
  const match = /^(\d{4})-(\d{2})-(\d{2})/.exec(value)
  if (!match) return null
  return new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3]))
}

function ageBetween(birth: Date, end: Date): number {
  let age = end.getFullYear() - birth.getFullYear()
  const monthDiff = end.getMonth() - birth.getMonth()
  if (monthDiff < 0 || (monthDiff === 0 && end.getDate() < birth.getDate())) {
    age -= 1
  }
  return Math.max(age, 0)
}

function resolveAgeEndDate(input: {
  deathDate?: string | null
  deathYear?: number | null
  isAlive?: boolean | null
}): Date {
  if (input.deathDate) {
    const death = parseISODate(input.deathDate)
    if (death) return death
  }
  if (input.isAlive === false && input.deathYear) {
    return new Date(input.deathYear, 11, 31)
  }
  return new Date()
}

export function formatMemberAge(input: {
  birthDate?: string | null
  birthYear?: number | null
  deathDate?: string | null
  deathYear?: number | null
  isAlive?: boolean | null
}): string {
  const endDate = resolveAgeEndDate(input)

  if (input.birthDate) {
    const birth = parseISODate(input.birthDate)
    if (birth) return `${ageBetween(birth, endDate)}岁`
  }

  if (input.birthYear) {
    const endYear =
      input.deathYear ??
      (input.isAlive === false ? endDate.getFullYear() : new Date().getFullYear())
    const age = endYear - input.birthYear
    if (age >= 0) return `约${age}岁`
  }

  return '未知'
}

export function formatMemberGender(gender: string): string {
  switch (gender) {
    case 'MALE':
      return '男'
    case 'FEMALE':
      return '女'
    default:
      return '未知'
  }
}

export function formatMemberLiving(isAlive?: boolean | null): string {
  if (isAlive === false) return '已故'
  if (isAlive === true) return '健在'
  return '未知'
}

export function formatTreeBindingState(state: string): string {
  switch (state) {
    case 'NOT_REQUIRED':
      return '无需绑定'
    case 'BOUND':
      return '已绑定'
    case 'UNBOUND':
      return '未绑定'
    case 'INVITING':
      return '邀请中'
    default:
      return '未知'
  }
}

export function formatMemberBindingNeed(member: Pick<FamilyMember, 'userBindingPolicy' | 'boundUserId'>): string {
  switch (member.userBindingPolicy) {
    case 'NOT_REQUIRED':
      return '无需绑定'
    case 'REQUIRED':
      return member.boundUserId ? '已绑定' : '待绑定'
    case 'OPTIONAL':
      return member.boundUserId ? '已绑定' : '可选绑定'
    default:
      return '未设置'
  }
}
