import {
  buildUnsupportedCanonicalResult
} from '../kinship/canonicalKinshipTerms'
import { validateRegistryRelativeTitle } from '../kinship/officialCanonicalTitles'
import type { CanonicalKinshipResult } from '../kinship/canonicalTypes'
import { describeGraphObjectivePath } from './graphObjectivePath'
import type { RelGraph } from './relativeGraph'
import {
  bloodGenerationGap,
  childrenOf,
  findBloodAncestorPath,
  findBloodDescendantPath,
  genderOf,
  isParentOf,
  isSiblingOf,
  isSpouseOf,
  maternalCrossCount,
  parentsOf,
  siblingsOf,
  spousesOf
} from './relativeGraph'
import {
  compareReliableSeniority,
  compareSeniority,
  groupHasCompleteBirth,
  rankLabelInGroup,
  rankedSiblingCanonical,
  rankedSiblingTitle,
  sameParentSiblingGroup
} from './relativeRank'

type GrandparentLine =
  | 'paternalGrandfather'
  | 'paternalGrandmother'
  | 'maternalGrandfather'
  | 'maternalGrandmother'

function spouseTitle(graph: RelGraph, meId: number, targetId: number): string {
  const myGender = genderOf(graph, meId)
  const targetGender = genderOf(graph, targetId)
  if (myGender === 'MALE' && targetGender === 'FEMALE') return '妻子'
  if (myGender === 'FEMALE' && targetGender === 'MALE') return '丈夫'
  return '配偶'
}

function titleFromAncestorPath(graph: RelGraph, path: number[]): string {
  const depth = path.length - 1
  const targetId = path[path.length - 1]
  const targetGender = genderOf(graph, targetId)
  const isMale = targetGender === 'MALE'
  const isFemale = targetGender === 'FEMALE'

  if (depth === 1) {
    if (isMale) return '父亲'
    if (isFemale) return '母亲'
    return ''
  }

  if (depth === 2) {
    const via = genderOf(graph, path[1])
    if (via === 'MALE') return isMale ? '祖父' : isFemale ? '祖母' : ''
    if (via === 'FEMALE') return isMale ? '外祖父' : isFemale ? '外祖母' : ''
    return ''
  }

  if (depth === 3) {
    const p = genderOf(graph, path[1])
    const gp = genderOf(graph, path[2])
    if (p === 'MALE' && gp === 'MALE') return isMale ? '曾祖父' : isFemale ? '曾祖母' : ''
    if (p === 'MALE' && gp === 'FEMALE') return isMale ? '曾外祖父' : isFemale ? '曾外祖母' : ''
    if (p === 'FEMALE' && gp === 'MALE') return isMale ? '外曾祖父' : isFemale ? '外曾祖母' : ''
    if (p === 'FEMALE' && gp === 'FEMALE') return isMale ? '外曾外祖父' : isFemale ? '外曾外祖母' : ''
    return ''
  }

  if (depth === 4) {
    const prefix = maternalCrossCount(graph, path) === 0 ? '' : '外'
    return isMale ? `${prefix}高祖父` : isFemale ? `${prefix}高祖母` : ''
  }

  if (depth === 5) {
    const prefix = maternalCrossCount(graph, path) === 0 ? '' : '外'
    return isMale ? `${prefix}五世祖父` : isFemale ? `${prefix}五世祖母` : ''
  }

  return ''
}

function collectAncestorPaths(graph: RelGraph, meId: number, maxDepth = 5): number[][] {
  const result: number[][] = []
  const queue: number[][] = [[meId]]

  while (queue.length > 0) {
    const path = queue.shift()!
    const depth = path.length - 1
    if (depth >= maxDepth) continue

    const head = path[path.length - 1]
    for (const parentId of parentsOf(graph, head)) {
      if (path.includes(parentId)) continue
      const next = [...path, parentId]
      result.push(next)
      queue.push(next)
    }
  }

  return result
}

function replaceAncestorSuffix(title: string, suffix: string, replacement: string): string {
  if (!title.endsWith(suffix)) return ''
  return `${title.slice(0, -suffix.length)}${replacement}`
}

function distantAncestorSiblingTitle(
  graph: RelGraph,
  ancestorPath: number[],
  targetId: number
): string {
  const anchorId = ancestorPath[ancestorPath.length - 1]
  const anchorTitle = titleFromAncestorPath(graph, ancestorPath)
  const anchorGender = genderOf(graph, anchorId)
  const targetGender = genderOf(graph, targetId)

  if (anchorGender === 'MALE') {
    if (targetGender === 'FEMALE') return replaceAncestorSuffix(anchorTitle, '祖父', '姑祖母')
    if (targetGender === 'MALE') {
      const seniority = compareReliableSeniority(graph, anchorId, targetId)
      if (seniority === 'older') return replaceAncestorSuffix(anchorTitle, '祖父', '伯祖父')
      if (seniority === 'younger') return replaceAncestorSuffix(anchorTitle, '祖父', '叔祖父')
      return ''
    }
  }

  if (anchorGender === 'FEMALE') {
    if (targetGender === 'MALE') return replaceAncestorSuffix(anchorTitle, '祖母', '舅祖父')
    if (targetGender === 'FEMALE') return replaceAncestorSuffix(anchorTitle, '祖母', '姨祖母')
  }

  return ''
}

function distantAncestorSiblingSpouseTitle(
  graph: RelGraph,
  ancestorPath: number[],
  grandSiblingId: number
): string {
  return spouseTitleOfCollateral(distantAncestorSiblingTitle(graph, ancestorPath, grandSiblingId))
}

function titleFromDescendantPath(graph: RelGraph, path: number[]): string {
  const depth = path.length - 1
  const targetId = path[path.length - 1]
  const targetGender = genderOf(graph, targetId)
  const isMale = targetGender === 'MALE'
  const isFemale = targetGender === 'FEMALE'

  if (depth === 1) {
    if (isMale) return '儿子'
    if (isFemale) return '女儿'
    return ''
  }

  if (depth === 2) {
    const via = genderOf(graph, path[1])
    if (via === 'MALE') return isMale ? '孙子' : isFemale ? '孙女' : ''
    if (via === 'FEMALE') return isMale ? '外孙' : isFemale ? '外孙女' : ''
    return ''
  }

  if (depth === 3) {
    const via = genderOf(graph, path[1])
    if (via === 'MALE') return isMale ? '曾孙' : isFemale ? '曾孙女' : ''
    if (via === 'FEMALE') return isMale ? '外曾孙' : isFemale ? '外曾孙女' : ''
    return ''
  }

  if (depth === 4) {
    const via = genderOf(graph, path[1])
    if (via === 'MALE') return isMale ? '玄孙' : isFemale ? '玄孙女' : ''
    if (via === 'FEMALE') return isMale ? '外玄孙' : isFemale ? '外玄孙女' : ''
    return ''
  }

  return ''
}

function resolveGrandparentLine(parentGender: string, gpGender: string): GrandparentLine | null {
  if (parentGender === 'MALE' && gpGender === 'MALE') return 'paternalGrandfather'
  if (parentGender === 'MALE' && gpGender === 'FEMALE') return 'paternalGrandmother'
  if (parentGender === 'FEMALE' && gpGender === 'MALE') return 'maternalGrandfather'
  if (parentGender === 'FEMALE' && gpGender === 'FEMALE') return 'maternalGrandmother'
  return null
}

function paternalGrandfatherSiblingTitle(graph: RelGraph, anchorId: number, targetId: number): string {
  const targetGender = genderOf(graph, targetId)
  if (targetGender === 'FEMALE') {
    const sisters = sameParentSiblingGroup(graph, anchorId, 'FEMALE')
    const idx = sisters.indexOf(targetId)
    if (idx < 0) return '姑祖母'
    return '姑祖母'
  }
  if (targetGender === 'MALE') {
    const seniority = compareReliableSeniority(graph, anchorId, targetId)
    if (seniority === 'older') return '伯祖父'
    if (seniority === 'younger') return '叔祖父'
    return ''
  }
  return ''
}

function paternalGrandmotherSiblingTitle(graph: RelGraph, anchorId: number, targetId: number): string {
  const targetGender = genderOf(graph, targetId)
  const base = targetGender === 'MALE' ? '舅祖父' : '姨祖母'
  const group = sameParentSiblingGroup(graph, anchorId, targetGender)
  const idx = group.indexOf(targetId)
  if (idx < 0) return base
  return base
}

function maternalGrandfatherSiblingTitle(graph: RelGraph, anchorId: number, targetId: number): string {
  const targetGender = genderOf(graph, targetId)
  if (targetGender === 'FEMALE') {
    return '外姑祖母'
  }
  if (targetGender === 'MALE') {
    const seniority = compareReliableSeniority(graph, anchorId, targetId)
    if (seniority === 'older') return '外伯祖父'
    if (seniority === 'younger') return '外叔祖父'
    return ''
  }
  return ''
}

function maternalGrandmotherSiblingTitle(graph: RelGraph, anchorId: number, targetId: number): string {
  const targetGender = genderOf(graph, targetId)
  const base = targetGender === 'MALE' ? '外舅祖父' : '外姨祖母'
  const group = sameParentSiblingGroup(graph, anchorId, targetGender)
  const idx = group.indexOf(targetId)
  if (idx < 0) return base
  return base
}

function grandparentSiblingTitle(
  graph: RelGraph,
  anchorGpId: number,
  targetId: number,
  line: GrandparentLine
): string {
  switch (line) {
    case 'paternalGrandfather':
      return paternalGrandfatherSiblingTitle(graph, anchorGpId, targetId)
    case 'paternalGrandmother':
      return paternalGrandmotherSiblingTitle(graph, anchorGpId, targetId)
    case 'maternalGrandfather':
      return maternalGrandfatherSiblingTitle(graph, anchorGpId, targetId)
    case 'maternalGrandmother':
      return maternalGrandmotherSiblingTitle(graph, anchorGpId, targetId)
    default:
      return ''
  }
}

function spouseTitleOfCollateral(title: string): string {
  if (title.endsWith('伯祖父')) return title.replace(/伯祖父$/, '伯祖母')
  if (title.endsWith('叔祖父')) return title.replace(/叔祖父$/, '叔祖母')
  if (title.endsWith('姑祖母')) return title.replace(/姑祖母$/, '姑祖父')
  if (title.endsWith('舅祖父')) return title.replace(/舅祖父$/, '舅祖母')
  if (title.endsWith('姨祖母')) return title.replace(/姨祖母$/, '姨祖父')
  return ''
}

function grandparentSiblingSpouseTitle(
  graph: RelGraph,
  anchorGpId: number,
  grandSiblingId: number,
  line: GrandparentLine
): string {
  return spouseTitleOfCollateral(grandparentSiblingTitle(graph, anchorGpId, grandSiblingId, line))
}

function parentSiblingTitle(graph: RelGraph, parentId: number, targetId: number): string {
  const parentGender = genderOf(graph, parentId)
  const targetGender = genderOf(graph, targetId)

  if (parentGender === 'MALE') {
    if (targetGender === 'FEMALE') return '姑母'
    const seniority = compareReliableSeniority(graph, parentId, targetId)
    if (seniority === 'older') return '伯父'
    if (seniority === 'younger') return '叔父'
    return ''
  }

  if (parentGender === 'FEMALE') {
    if (targetGender === 'MALE') return '舅父'
    if (targetGender === 'FEMALE') return '姨母'
  }

  return ''
}

function parentSiblingSpouseTitle(graph: RelGraph, parentId: number, uncleId: number, targetId: number): string {
  const parentGender = genderOf(graph, parentId)
  const uncleGender = genderOf(graph, uncleId)
  const targetGender = genderOf(graph, targetId)

  if (parentGender === 'MALE' && uncleGender === 'MALE') {
    const seniority = compareReliableSeniority(graph, parentId, uncleId)
    if (seniority === 'older') return targetGender === 'FEMALE' ? '伯母' : '姑父'
    if (seniority === 'younger') return targetGender === 'FEMALE' ? '婶母' : '姑父'
    return ''
  }

  if (parentGender === 'MALE' && uncleGender === 'FEMALE') {
    return targetGender === 'MALE' ? '姑父' : '姑母'
  }

  if (parentGender === 'FEMALE' && uncleGender === 'MALE') {
    return targetGender === 'FEMALE' ? '舅母' : '舅父'
  }

  if (parentGender === 'FEMALE' && uncleGender === 'FEMALE') {
    return targetGender === 'MALE' ? '姨父' : '姨母'
  }

  return ''
}

function cousinKind(parentGender: string, uncleGender: string): 'tang' | 'biao' {
  if (parentGender === 'MALE' && uncleGender === 'MALE') return 'tang'
  return 'biao'
}

function cousinTitle(
  graph: RelGraph,
  meId: number,
  targetId: number,
  parentId: number,
  uncleId: number
): string {
  const kind = cousinKind(genderOf(graph, parentId), genderOf(graph, uncleId))
  const targetGender = genderOf(graph, targetId)
  /** 同辈长幼：以 viewer 与 target 出生先后为准，非伯父/叔父等父辈旁系长幼 */
  const seniority = compareSeniority(graph, meId, targetId)

  if (kind === 'tang') {
    if (targetGender === 'MALE') {
      if (seniority === 'older') return '堂兄'
      if (seniority === 'younger') return '堂弟'
      return ''
    }
    if (targetGender === 'FEMALE') {
      if (seniority === 'older') return '堂姐'
      if (seniority === 'younger') return '堂妹'
      return ''
    }
    return ''
  }

  if (targetGender === 'MALE') {
    if (seniority === 'older') return '表兄'
    if (seniority === 'younger') return '表弟'
    return ''
  }
  if (targetGender === 'FEMALE') {
    if (seniority === 'older') return '表姐'
    if (seniority === 'younger') return '表妹'
    return ''
  }
  return ''
}

/** 祖辈旁系的后代 = 父母的堂/表兄弟姐妹（父母辈，非我同辈） */
function parentGenerationCousinTitle(
  graph: RelGraph,
  parentId: number,
  gpId: number,
  targetId: number,
  line: GrandparentLine
): string {
  const targetGender = genderOf(graph, targetId)
  const seniority = compareSeniority(graph, parentId, targetId)

  if (line === 'paternalGrandfather') {
    if (targetGender === 'FEMALE') return '堂姑'
    if (targetGender === 'MALE') {
      if (seniority === 'older') return '堂伯'
      if (seniority === 'younger') return '堂叔'
      return ''
    }
    return ''
  }

  if (line === 'paternalGrandmother') {
    if (targetGender === 'FEMALE') return seniority === 'unknown' ? '' : '表姑'
    if (targetGender === 'MALE') {
      if (seniority === 'older') return '表伯'
      if (seniority === 'younger') return '表叔'
      return ''
    }
    return ''
  }

  if (line === 'maternalGrandfather' || line === 'maternalGrandmother') {
    if (targetGender === 'FEMALE') return seniority === 'unknown' ? '' : '表姨'
    if (targetGender === 'MALE') {
      if (seniority === 'older') return '表舅'
      if (seniority === 'younger') return '表叔'
      return ''
    }
    return ''
  }

  return ''
}

function cousinSpouseTitle(
  graph: RelGraph,
  meId: number,
  cousinId: number,
  parentId: number,
  uncleId: number
): string {
  const kind = cousinKind(genderOf(graph, parentId), genderOf(graph, uncleId))
  const prefix = kind === 'tang' ? '堂' : '表'
  const seniority = compareSeniority(graph, meId, cousinId)
  const cousinGender = genderOf(graph, cousinId)

  if (cousinGender === 'MALE') {
    if (seniority === 'older') return `${prefix}嫂`
    if (seniority === 'younger') return `${prefix}弟媳`
    return ''
  }
  if (cousinGender === 'FEMALE') {
    if (seniority === 'older') return `${prefix}姐夫`
    if (seniority === 'younger') return `${prefix}妹夫`
    return ''
  }
  return ''
}

function siblingSpouseTitle(graph: RelGraph, meId: number, siblingId: number): string {
  const siblingGender = genderOf(graph, siblingId)
  const seniority = compareSeniority(graph, meId, siblingId)
  if (siblingGender === 'MALE') {
    if (seniority === 'older') return '嫂子'
    if (seniority === 'younger') return '弟媳'
    return ''
  }
  if (siblingGender === 'FEMALE') {
    if (seniority === 'older') return '姐夫'
    if (seniority === 'younger') return '妹夫'
    return ''
  }
  return ''
}

function nephewTitle(graph: RelGraph, siblingId: number, targetId: number): string {
  const siblingGender = genderOf(graph, siblingId)
  const targetGender = genderOf(graph, targetId)
  if (siblingGender === 'MALE') {
    if (targetGender === 'MALE') return '侄子'
    if (targetGender === 'FEMALE') return '侄女'
  }
  if (siblingGender === 'FEMALE') {
    if (targetGender === 'MALE') return '外甥'
    if (targetGender === 'FEMALE') return '外甥女'
  }
  return ''
}

function nephewDescendantTitle(
  graph: RelGraph,
  siblingId: number,
  targetId: number,
  path: number[]
): string {
  const depth = path.length - 1
  const siblingGender = genderOf(graph, siblingId)
  const targetGender = genderOf(graph, targetId)
  const isMale = targetGender === 'MALE'
  const isFemale = targetGender === 'FEMALE'

  if (depth === 1) return nephewTitle(graph, siblingId, targetId)
  if (depth !== 2) return ''

  if (siblingGender === 'MALE') {
    const via = genderOf(graph, path[1])
    if (via === 'MALE') return isMale ? '侄孙' : isFemale ? '侄孙女' : ''
    if (via === 'FEMALE') return isMale ? '侄外孙' : isFemale ? '侄外孙女' : ''
    return ''
  }

  if (siblingGender === 'FEMALE') {
    const via = genderOf(graph, path[1])
    if (via === 'MALE') return isMale ? '外甥孙' : isFemale ? '外甥孙女' : ''
    if (via === 'FEMALE') return isMale ? '外甥外孙' : isFemale ? '外甥外孙女' : ''
    return ''
  }

  return ''
}

function cousinChildTitle(
  graph: RelGraph,
  parentId: number,
  uncleId: number,
  targetId: number
): string {
  const kind = cousinKind(genderOf(graph, parentId), genderOf(graph, uncleId))
  const prefix = kind === 'tang' ? '堂' : '表'
  const targetGender = genderOf(graph, targetId)
  if (targetGender === 'MALE') return `${prefix}侄`
  if (targetGender === 'FEMALE') return `${prefix}侄女`
  return ''
}

function cousinGrandchildTitle(
  graph: RelGraph,
  parentId: number,
  uncleId: number,
  targetId: number,
  path: number[]
): string {
  const kind = cousinKind(genderOf(graph, parentId), genderOf(graph, uncleId))
  const prefix = kind === 'tang' ? '堂' : '表'
  const targetGender = genderOf(graph, targetId)
  const depth = path.length - 1
  if (depth === 1) return cousinChildTitle(graph, parentId, uncleId, targetId)
  if (depth !== 2) return ''
  if (targetGender === 'MALE') return `${prefix}侄孙`
  if (targetGender === 'FEMALE') return `${prefix}侄孙女`
  return ''
}

function childSpouseTitle(graph: RelGraph, childId: number): string {
  const childGender = genderOf(graph, childId)
  if (childGender === 'MALE') return '儿媳'
  if (childGender === 'FEMALE') return '女婿'
  return ''
}

function grandchildSpouseTitle(graph: RelGraph, grandchildId: number): string {
  const gender = genderOf(graph, grandchildId)
  if (gender === 'MALE') return '孙媳'
  if (gender === 'FEMALE') return '孙女婿'
  return ''
}

function isSpouseSideRelative(graph: RelGraph, meId: number, targetId: number): boolean {
  for (const spouseId of spousesOf(graph, meId)) {
    if (
      findBloodAncestorPath(graph, spouseId, targetId) ||
      findBloodDescendantPath(graph, spouseId, targetId) ||
      isSiblingOf(graph, spouseId, targetId)
    ) {
      return true
    }
  }
  return false
}

/** 结构识别推导称谓（不使用 BFS 最短路径） */
export function resolveRelativeTitle(graph: RelGraph, meId: number, targetId: number): string {
  if (meId === targetId) return '我'

  if (isSpouseOf(graph, meId, targetId)) return spouseTitle(graph, meId, targetId)

  const ancestorPath = findBloodAncestorPath(graph, meId, targetId)
  if (ancestorPath) return titleFromAncestorPath(graph, ancestorPath)

  const descendantPath = findBloodDescendantPath(graph, meId, targetId)
  if (descendantPath) return titleFromDescendantPath(graph, descendantPath)

  if (isSiblingOf(graph, meId, targetId)) return rankedSiblingTitle(graph, meId, targetId)

  for (const siblingId of siblingsOf(graph, meId)) {
    if (isSpouseOf(graph, siblingId, targetId)) {
      return siblingSpouseTitle(graph, meId, siblingId)
    }
  }

  for (const parentId of parentsOf(graph, meId)) {
    for (const gpId of parentsOf(graph, parentId)) {
      if (gpId === targetId) continue
      if (isSiblingOf(graph, gpId, targetId)) {
        const line = resolveGrandparentLine(genderOf(graph, parentId), genderOf(graph, gpId))
        if (line) return grandparentSiblingTitle(graph, gpId, targetId, line)
      }
      const line = resolveGrandparentLine(genderOf(graph, parentId), genderOf(graph, gpId))
      if (!line) continue
      for (const grandSiblingId of siblingsOf(graph, gpId)) {
        if (isSpouseOf(graph, grandSiblingId, targetId)) {
          return grandparentSiblingSpouseTitle(graph, gpId, grandSiblingId, line)
        }
      }
    }
  }

  for (const path of collectAncestorPaths(graph, meId, 5)) {
    if (path.length - 1 < 3) continue
    const ancestorId = path[path.length - 1]
    if (isSiblingOf(graph, ancestorId, targetId)) {
      return distantAncestorSiblingTitle(graph, path, targetId)
    }
    for (const grandSiblingId of siblingsOf(graph, ancestorId)) {
      if (isSpouseOf(graph, grandSiblingId, targetId)) {
        return distantAncestorSiblingSpouseTitle(graph, path, grandSiblingId)
      }
    }
  }

  for (const parentId of parentsOf(graph, meId)) {
    if (isSiblingOf(graph, parentId, targetId)) {
      return parentSiblingTitle(graph, parentId, targetId)
    }
    for (const uncleId of siblingsOf(graph, parentId)) {
      if (isSpouseOf(graph, uncleId, targetId)) {
        return parentSiblingSpouseTitle(graph, parentId, uncleId, targetId)
      }
    }
  }

  for (const parentId of parentsOf(graph, meId)) {
    for (const uncleId of siblingsOf(graph, parentId)) {
      if (childrenOf(graph, uncleId).includes(targetId)) {
        return cousinTitle(graph, meId, targetId, parentId, uncleId)
      }
      for (const cousinId of childrenOf(graph, uncleId)) {
        if (isSpouseOf(graph, cousinId, targetId)) {
          return cousinSpouseTitle(graph, meId, cousinId, parentId, uncleId)
        }
      }
    }
  }

  /** 祖辈旁系子女 = 父母的堂/表兄弟姐妹（父母辈） */
  for (const parentId of parentsOf(graph, meId)) {
    for (const gpId of parentsOf(graph, parentId)) {
      const line = resolveGrandparentLine(genderOf(graph, parentId), genderOf(graph, gpId))
      if (!line) continue
      for (const grandUncleId of siblingsOf(graph, gpId)) {
        if (childrenOf(graph, grandUncleId).includes(targetId)) {
          return parentGenerationCousinTitle(graph, parentId, gpId, targetId, line)
        }
        for (const parentCousinId of childrenOf(graph, grandUncleId)) {
          if (isSpouseOf(graph, parentCousinId, targetId)) return ''
          const subPath = findBloodDescendantPath(graph, parentCousinId, targetId, 4)
          if (!subPath) continue
          const gap = bloodGenerationGap(graph, meId, targetId)
          if (gap === null) continue
          if (gap === 0) {
            return cousinTitle(graph, meId, targetId, parentId, parentCousinId)
          }
          if (gap === 1) {
            const kind = line === 'paternalGrandfather' ? '堂' : '表'
            const tg = genderOf(graph, targetId)
            if (tg === 'MALE') return `${kind}侄`
            if (tg === 'FEMALE') return `${kind}侄女`
            continue
          }
          if (gap === 2) {
            const kind = line === 'paternalGrandfather' ? '堂' : '表'
            const tg = genderOf(graph, targetId)
            if (tg === 'MALE') return `${kind}侄孙`
            if (tg === 'FEMALE') return `${kind}侄孙女`
          }
        }
      }
    }
  }

  for (const childId of childrenOf(graph, meId)) {
    if (isSpouseOf(graph, childId, targetId)) return childSpouseTitle(graph, childId)
    for (const grandchildId of childrenOf(graph, childId)) {
      if (isSpouseOf(graph, grandchildId, targetId)) return grandchildSpouseTitle(graph, grandchildId)
    }
  }

  for (const parentId of parentsOf(graph, meId)) {
    for (const uncleId of siblingsOf(graph, parentId)) {
      for (const cousinId of childrenOf(graph, uncleId)) {
        const subPath = findBloodDescendantPath(graph, cousinId, targetId, 4)
        if (!subPath) continue
        const gap = bloodGenerationGap(graph, meId, targetId)
        if (gap === null) continue
        if (gap === 1 && subPath.length === 2) {
          return cousinChildTitle(graph, parentId, uncleId, targetId)
        }
        if (gap === 2 && subPath.length === 3) {
          return cousinGrandchildTitle(graph, parentId, uncleId, targetId, subPath)
        }
      }
    }
  }

  for (const siblingId of siblingsOf(graph, meId)) {
    const subPath = findBloodDescendantPath(graph, siblingId, targetId, 6)
    if (!subPath) continue
    const gap = bloodGenerationGap(graph, meId, targetId)
    if (gap === null) continue
    if (gap === 1 && subPath.length === 2) return nephewTitle(graph, siblingId, targetId)
    if (gap === 2 && subPath.length === 3) {
      return nephewDescendantTitle(graph, siblingId, targetId, subPath)
    }
  }

  if (isSpouseSideRelative(graph, meId, targetId)) return ''

  return ''
}

export function resolveCanonicalRelativeTitle(
  graph: RelGraph,
  meId: number,
  targetId: number
): CanonicalKinshipResult {
  const pathDescription = describeGraphObjectivePath(graph, meId, targetId)
  const node = graph.nodeMap.get(meId)
  if (node?.memberType === 'SPOUSE') {
    return buildUnsupportedCanonicalResult(pathDescription, {
      incompleteInfo: false,
      explanation: '配偶节点不能作为族内称谓视角中心。'
    })
  }

  const siblingCanonical = rankedSiblingCanonical(graph, meId, targetId)
  if (siblingCanonical) {
    const validated = validateRegistryRelativeTitle(siblingCanonical.canonicalTitle)
    if (!validated.ok) {
      return buildUnsupportedCanonicalResult(pathDescription, {
        incompleteInfo: false,
        explanation: '兄弟姐妹称谓未在官方注册表中登记。'
      })
    }
    return {
      canonicalTitle: validated.title,
      rankLabel: siblingCanonical.rankLabel,
      unsupportedCanonicalTitle: false,
      incompleteInfo: false,
      pathDescription
    }
  }

  const title = resolveRelativeTitle(graph, meId, targetId)
  if (!title) {
    return buildUnsupportedCanonicalResult(pathDescription, {
      incompleteInfo: false,
      explanation: '暂无对应规范称谓。'
    })
  }

  const validated = validateRegistryRelativeTitle(title)
  if (!validated.ok) {
    return buildUnsupportedCanonicalResult(pathDescription, {
      incompleteInfo: false,
      explanation:
        validated.reason === 'forbidden'
          ? '该关系仅有泛称，无规范称谓。'
          : '该称谓尚未纳入官方注册表。'
    })
  }

  return {
    canonicalTitle: validated.title,
    unsupportedCanonicalTitle: false,
    incompleteInfo: false,
    pathDescription
  }
}
