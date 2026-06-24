import type { TreeNode } from '@/types/api'

/** 从出生日期提取可比较的时间戳；缺失时排到最后 */
export function birthSortValue(node?: TreeNode): number {
  if (!node?.birthDate) return Number.MAX_SAFE_INTEGER
  const normalized = node.birthDate.replace(/\./g, '-').trim()
  const match = /^(\d{4})(?:-(\d{1,2}))?(?:-(\d{1,2}))?/.exec(normalized)
  if (!match) return Number.MAX_SAFE_INTEGER
  const year = Number(match[1])
  const month = match[2] ? Number(match[2]) : 1
  const day = match[3] ? Number(match[3]) : 1
  return year * 10000 + month * 100 + day
}

export function memberIdTieBreak(a: TreeNode, b: TreeNode): number {
  return a.memberId - b.memberId
}

/** 男性子女：长幼（生日 → memberId） */
export function compareMaleChildren(a: TreeNode, b: TreeNode): number {
  const byBirth = birthSortValue(a) - birthSortValue(b)
  if (byBirth !== 0) return byBirth
  return memberIdTieBreak(a, b)
}

function surnameOf(node: TreeNode): string {
  return (node.surname || node.displayName.trim().slice(0, 1) || '').trim()
}

/** 女性子女：姓氏/姓名 → memberId */
export function compareFemaleChildren(a: TreeNode, b: TreeNode): number {
  const bySurname = surnameOf(a).localeCompare(surnameOf(b), 'zh-CN')
  if (bySurname !== 0) return bySurname
  const byName = a.displayName.localeCompare(b.displayName, 'zh-CN')
  if (byName !== 0) return byName
  return memberIdTieBreak(a, b)
}

/** 配偶排序：稳定、不参与兄弟姐妹序 */
export function compareSpouses(a: TreeNode, b: TreeNode): number {
  const byBirth = birthSortValue(a) - birthSortValue(b)
  if (byBirth !== 0) return byBirth
  return memberIdTieBreak(a, b)
}

export function sortChildMemberIds(childIds: number[], nodeMap: Map<number, TreeNode>): number[] {
  const males: number[] = []
  const females: number[] = []
  const others: number[] = []

  for (const id of childIds) {
    const node = nodeMap.get(id)
    if (!node) continue
    if (node.gender === 'MALE') males.push(id)
    else if (node.gender === 'FEMALE') females.push(id)
    else others.push(id)
  }

  males.sort((a, b) => compareMaleChildren(nodeMap.get(a)!, nodeMap.get(b)!))
  females.sort((a, b) => compareFemaleChildren(nodeMap.get(a)!, nodeMap.get(b)!))
  others.sort((a, b) => memberIdTieBreak(nodeMap.get(a)!, nodeMap.get(b)!))

  return [...males, ...females, ...others]
}
