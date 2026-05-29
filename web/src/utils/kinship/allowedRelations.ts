import type { BaseRelation, KinshipContext, RelationOption } from './types'
import { BASE_RELATIONS, getSteps, MAX_KINSHIP_LEVEL, pathKey, RELATION_LABELS } from './helpers'
import { isTerminalPath } from './terminalRules'

function allSame(items: string[], value: string) {
  return items.length > 0 && items.every((item) => item === value)
}

function blockReason(pathLength: number, terminal: boolean, relation: BaseRelation) {
  if (pathLength >= MAX_KINSHIP_LEVEL) return '最多只能添加 5 层关系。'
  if (terminal) return '当前路径到此为止。'

  if (relation === 'parent') return '为避免从后代路径绕回祖辈，当前不允许选择父母。'
  if (relation === 'child') return '为避免从祖辈路径绕回后代，当前不允许选择子女。'
  if (relation === 'spouse') return '当前路径下配偶关系会造成姻亲回环，暂不建议继续选择。'
  if (relation === 'sibling') return '当前路径下兄弟姐妹关系会造成旁系回环，暂不建议继续选择。'
  return '当前路径在基础关系模式下不建议继续扩展。'
}

export function allowedRelations(context: KinshipContext): BaseRelation[] {
  const path = getSteps(context)
  const relations = path.map((item) => item.relation)
  const key = pathKey(path)

  if (path.length >= MAX_KINSHIP_LEVEL) return []
  if (isTerminalPath(path)) return []
  if (!path.length) return ['parent', 'child', 'sibling', 'spouse']

  if (allSame(relations, 'parent')) {
    if (path.length <= 3) return ['parent', 'sibling', 'spouse']
    return ['parent']
  }

  if (allSame(relations, 'child')) {
    if (path.length <= 3) return ['child', 'sibling', 'spouse']
    return ['child']
  }

  if (key === 'spouse') return ['parent', 'sibling', 'child']
  if (key === 'spouse/parent') return ['parent', 'sibling']
  if (key === 'spouse/sibling') return ['spouse', 'child']
  if (key === 'spouse/sibling/child') return ['child', 'spouse']
  if (key === 'spouse/child') return ['child', 'spouse']

  if (key === 'sibling') return ['spouse', 'child']
  if (key === 'sibling/child') return ['child', 'spouse']
  if (key === 'sibling/child/child') return ['child', 'spouse']

  if (key === 'parent/sibling') return ['spouse', 'child']
  if (key === 'parent/sibling/spouse') return ['child']
  if (key === 'parent/sibling/child') return ['child', 'spouse']
  if (key === 'parent/sibling/child/child') return ['spouse']

  if (key === 'parent/parent/sibling') return ['spouse']
  if (key === 'parent/parent/sibling/spouse') return []
  if (key === 'parent/parent/sibling/child') return ['spouse']

  const last = relations[relations.length - 1]

  if (last === 'parent') return ['parent', 'sibling', 'spouse']
  if (last === 'child') return ['child', 'spouse']
  if (last === 'sibling') return ['spouse', 'child']
  if (last === 'spouse') return ['parent', 'sibling', 'child']

  return []
}

export function relationOptionsFor(context: KinshipContext): RelationOption[] {
  const path = getSteps(context)
  const allowed = new Set(allowedRelations(context))
  const terminal = isTerminalPath(path)

  return BASE_RELATIONS.map((item) => {
    const disabled = !allowed.has(item.value)
    return {
      value: item.value,
      label: item.label,
      disabled,
      reason: disabled ? blockReason(path.length, terminal, item.value) : `可选择${RELATION_LABELS[item.value]}。`
    }
  })
}
