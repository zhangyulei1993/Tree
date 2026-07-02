import type { RelGraph } from './relativeGraph'
import {
  childrenOf,
  genderOf,
  isParentOf,
  isSpouseOf,
  parentsOf,
  spousesOf
} from './relativeGraph'

type StepKind = 'parent' | 'child' | 'spouse'

function stepLabel(graph: RelGraph, fromId: number, toId: number, kind: StepKind): string {
  const targetGender = genderOf(graph, toId)
  if (kind === 'parent') {
    if (targetGender === 'MALE') return '父亲'
    if (targetGender === 'FEMALE') return '母亲'
    return '父母'
  }
  if (kind === 'child') {
    if (targetGender === 'MALE') return '儿子'
    if (targetGender === 'FEMALE') return '女儿'
    return '子女'
  }
  const fromGender = genderOf(graph, fromId)
  if (fromGender === 'MALE' && targetGender === 'FEMALE') return '妻子'
  if (fromGender === 'FEMALE' && targetGender === 'MALE') return '丈夫'
  return '配偶'
}

function inferStepKind(graph: RelGraph, fromId: number, toId: number): StepKind | null {
  if (isParentOf(graph, toId, fromId)) return 'parent'
  if (isParentOf(graph, fromId, toId)) return 'child'
  if (isSpouseOf(graph, fromId, toId)) return 'spouse'
  return null
}

/** 沿 parent/child/spouse 边的最短路径，用于无法解析规范称谓时的客观路径展示。 */
export function describeGraphObjectivePath(graph: RelGraph, meId: number, targetId: number): string {
  if (meId === targetId) return '我'

  const queue: Array<{ id: number; path: number[] }> = [{ id: meId, path: [meId] }]
  const visited = new Set<number>([meId])
  const maxDepth = 12

  while (queue.length > 0) {
    const current = queue.shift()!
    if (current.id === targetId) {
      const labels: string[] = ['我']
      for (let i = 0; i < current.path.length - 1; i += 1) {
        const fromId = current.path[i]
        const toId = current.path[i + 1]
        const kind = inferStepKind(graph, fromId, toId)
        labels.push(kind ? stepLabel(graph, fromId, toId, kind) : graph.nodeMap.get(toId)?.displayName || '成员')
      }
      return labels.join(' → ')
    }

    if (current.path.length - 1 >= maxDepth) continue
    const head = current.id

    for (const parentId of parentsOf(graph, head)) {
      if (visited.has(parentId)) continue
      visited.add(parentId)
      queue.push({ id: parentId, path: [...current.path, parentId] })
    }
    for (const childId of childrenOf(graph, head)) {
      if (visited.has(childId)) continue
      visited.add(childId)
      queue.push({ id: childId, path: [...current.path, childId] })
    }
    for (const spouseId of spousesOf(graph, head)) {
      if (visited.has(spouseId)) continue
      visited.add(spouseId)
      queue.push({ id: spouseId, path: [...current.path, spouseId] })
    }
  }

  const meName = graph.nodeMap.get(meId)?.displayName || '我'
  const targetName = graph.nodeMap.get(targetId)?.displayName || '成员'
  return `${meName} → ${targetName}`
}

export function isValidObjectivePathDescription(pathDescription: string | undefined): boolean {
  if (!pathDescription?.trim()) return false
  if (!pathDescription.includes('→') && pathDescription !== '我') return false
  const segments = pathDescription.split('→').map((part) => part.trim())
  if (segments.length < 1 || segments[0] !== '我') return false
  return segments.every((segment) => segment.length > 0)
}
