import { bloodGenerationGap, buildRelGraph } from './relativeGraph'
import { isValidObjectivePathDescription } from './graphObjectivePath'
import { resolveCanonicalRelativeTitle } from './relativeRules'
import { buildIntraLineage43Fixture } from './canonicalMatrixFixture'
import {
  FORBIDDEN_BARE_CANONICAL_TITLES,
  isOfficialCanonicalTitle
} from '../kinship/officialCanonicalTitles'

export interface MatrixGoldenPair {
  name: string
  viewerId: number
  targetId: number
  expectedTitle: string
}

export interface ReverseKinshipPair {
  name: string
  forwardViewer: number
  forwardTarget: number
  forwardTitle: string
  reverseTitle: string
}

function buildGoldenPairs(): MatrixGoldenPair[] {
  const { ids } = buildIntraLineage43Fixture()
  const me = ids.me
  return [
    { name: 'me-father', viewerId: me, targetId: ids.fu, expectedTitle: '父亲' },
    { name: 'me-mother', viewerId: me, targetId: ids.mu, expectedTitle: '母亲' },
    { name: 'me-grandfather', viewerId: me, targetId: ids.gong, expectedTitle: '祖父' },
    { name: 'me-grandmother', viewerId: me, targetId: ids.gongSpouse, expectedTitle: '祖母' },
    { name: 'me-great-grandfather', viewerId: me, targetId: ids.zengfu, expectedTitle: '曾祖父' },
    { name: 'me-great-grandmother', viewerId: me, targetId: ids.zengmu, expectedTitle: '曾祖母' },
    { name: 'me-wife', viewerId: me, targetId: ids.spouse, expectedTitle: '妻子' },
    { name: 'me-son', viewerId: me, targetId: ids.son, expectedTitle: '儿子' },
    { name: 'me-daughter', viewerId: me, targetId: ids.daughter, expectedTitle: '女儿' },
    { name: 'me-grandson', viewerId: me, targetId: ids.grandson, expectedTitle: '孙子' },
    { name: 'me-granddaughter', viewerId: me, targetId: ids.granddaughter, expectedTitle: '孙女' },
    { name: 'me-great-grandson', viewerId: me, targetId: ids.greatGrandson, expectedTitle: '曾孙' },
    { name: 'me-uncle-younger', viewerId: me, targetId: ids.shufu, expectedTitle: '叔父' },
    { name: 'me-aunt-paternal', viewerId: me, targetId: ids.gufu, expectedTitle: '姑母' },
    { name: 'me-uncle-maternal', viewerId: me, targetId: ids.muBro, expectedTitle: '舅父' },
    { name: 'me-aunt-maternal', viewerId: me, targetId: ids.muSis, expectedTitle: '姨母' },
    { name: 'me-granduncle-elder', viewerId: me, targetId: ids.gongBro1, expectedTitle: '伯祖父' },
    { name: 'me-granduncle-younger', viewerId: me, targetId: ids.gongBro2, expectedTitle: '叔祖父' },
    { name: 'me-grandaunt', viewerId: me, targetId: ids.gongSis, expectedTitle: '姑祖母' },
    { name: 'me-cousin-tang-elder', viewerId: me, targetId: ids.tangGe, expectedTitle: '堂兄' },
    { name: 'me-cousin-tang-younger', viewerId: me, targetId: ids.tangDi, expectedTitle: '堂弟' },
    { name: 'me-cousin-biao', viewerId: me, targetId: ids.biaoGe, expectedTitle: '表兄' },
    { name: 'me-cousin-biao-female', viewerId: me, targetId: ids.biaoMei, expectedTitle: '表妹' },
    { name: 'me-nephew', viewerId: me, targetId: ids.nephew, expectedTitle: '侄子' },
    { name: 'me-niece', viewerId: me, targetId: ids.niece, expectedTitle: '侄女' },
    { name: 'me-maternal-nephew', viewerId: me, targetId: ids.waiSheng, expectedTitle: '外甥' },
    { name: 'me-maternal-niece', viewerId: me, targetId: ids.waiShengNv, expectedTitle: '外甥女' },
    { name: 'me-maternal-grandfather', viewerId: me, targetId: ids.waiGong, expectedTitle: '外祖父' },
    { name: 'me-maternal-grandmother', viewerId: me, targetId: ids.waiPo, expectedTitle: '外祖母' },
    { name: 'father-son', viewerId: ids.fu, targetId: me, expectedTitle: '儿子' },
    { name: 'son-father', viewerId: ids.son, targetId: me, expectedTitle: '父亲' },
    { name: 'grandfather-grandson', viewerId: ids.gong, targetId: ids.me, expectedTitle: '孙子' },
    { name: 'nephew-uncle', viewerId: ids.nephew, targetId: me, expectedTitle: '叔父' }
  ]
}

function buildReversePairs(): ReverseKinshipPair[] {
  const { ids } = buildIntraLineage43Fixture()
  return [
    {
      name: 'father-son-reverse',
      forwardViewer: ids.me,
      forwardTarget: ids.fu,
      forwardTitle: '父亲',
      reverseTitle: '儿子'
    },
    {
      name: 'mother-child-reverse',
      forwardViewer: ids.me,
      forwardTarget: ids.mu,
      forwardTitle: '母亲',
      reverseTitle: '儿子'
    },
    {
      name: 'grandparent-grandchild-reverse',
      forwardViewer: ids.me,
      forwardTarget: ids.gong,
      forwardTitle: '祖父',
      reverseTitle: '孙子'
    },
    {
      name: 'uncle-nephew-reverse',
      forwardViewer: ids.me,
      forwardTarget: ids.shufu,
      forwardTitle: '叔父',
      reverseTitle: '侄子'
    },
    {
      name: 'cousin-tang-reverse',
      forwardViewer: ids.me,
      forwardTarget: ids.tangGe,
      forwardTitle: '堂兄',
      reverseTitle: '堂弟'
    }
  ]
}

export function runCanonicalMatrixTest(): {
  lineageCenterCount: number
  goldenTotal: number
  goldenPassed: number
  spouseRejected: number
  failures: string[]
} {
  const { nodes, edges, lineageMemberIds, spouseMemberIds } = buildIntraLineage43Fixture()
  const graph = buildRelGraph(nodes, edges)
  const failures: string[] = []

  if (lineageMemberIds.length !== 43) {
    failures.push(`族内中心数量应为 43，实际 ${lineageMemberIds.length}`)
  }

  for (const spouseId of spouseMemberIds) {
    const targetId = lineageMemberIds[0]
    const result = resolveCanonicalRelativeTitle(graph, spouseId, targetId)
    if (!result.unsupportedCanonicalTitle) {
      failures.push(`配偶节点 ${spouseId} 作为视角中心应被拒绝，实际返回 ${result.canonicalTitle || result.pathDescription}`)
    }
  }

  const goldenPairs = buildGoldenPairs()
  let goldenPassed = 0
  for (const pair of goldenPairs) {
    const result = resolveCanonicalRelativeTitle(graph, pair.viewerId, pair.targetId)
    if (result.unsupportedCanonicalTitle || !result.canonicalTitle) {
      failures.push(
        `${pair.name}: 期望规范称谓「${pair.expectedTitle}」，实际 unsupported=${result.unsupportedCanonicalTitle} title=${result.canonicalTitle || '无'} path=${result.pathDescription}`
      )
      continue
    }
    if (result.canonicalTitle !== pair.expectedTitle) {
      failures.push(
        `${pair.name}: 期望「${pair.expectedTitle}」，实际「${result.canonicalTitle}」`
      )
      continue
    }
    goldenPassed += 1
  }

  for (const centerId of lineageMemberIds) {
    const selfResult = resolveCanonicalRelativeTitle(graph, centerId, centerId)
    if (selfResult.canonicalTitle !== '我') {
      failures.push(`中心 ${centerId} 自身称谓应为「我」，实际 ${selfResult.canonicalTitle}`)
    }
  }

  return {
    lineageCenterCount: lineageMemberIds.length,
    goldenTotal: goldenPairs.length,
    goldenPassed,
    spouseRejected: spouseMemberIds.length,
    failures
  }
}

export function runReverseKinshipTest(): {
  total: number
  passed: number
  failures: string[]
} {
  const { nodes, edges } = buildIntraLineage43Fixture()
  const graph = buildRelGraph(nodes, edges)
  const pairs = buildReversePairs()
  const failures: string[] = []
  let passed = 0

  for (const pair of pairs) {
    const forward = resolveCanonicalRelativeTitle(graph, pair.forwardViewer, pair.forwardTarget)
    const reverse = resolveCanonicalRelativeTitle(graph, pair.forwardTarget, pair.forwardViewer)

    if (forward.unsupportedCanonicalTitle || forward.canonicalTitle !== pair.forwardTitle) {
      failures.push(
        `${pair.name} 正向: 期望「${pair.forwardTitle}」，实际「${forward.canonicalTitle || forward.pathDescription}」`
      )
      continue
    }
    if (reverse.unsupportedCanonicalTitle || reverse.canonicalTitle !== pair.reverseTitle) {
      failures.push(
        `${pair.name} 反向: 期望「${pair.reverseTitle}」，实际「${reverse.canonicalTitle || reverse.pathDescription}」`
      )
      continue
    }
    passed += 1
  }

  return { total: pairs.length, passed, failures }
}

function isNephewNieceTitle(title: string): boolean {
  return title === '侄子'
    || title === '侄女'
    || title === '堂侄'
    || title === '堂侄女'
    || title === '表侄'
    || title === '表侄女'
    || title === '外甥'
    || title === '外甥女'
}

function isNephewGrandchildTitle(title: string): boolean {
  return title === '侄孙'
    || title === '侄孙女'
    || title === '侄外孙'
    || title === '侄外孙女'
    || title === '外甥孙'
    || title === '外甥孙女'
    || title === '外甥外孙'
    || title === '外甥外孙女'
    || title === '堂侄孙'
    || title === '堂侄孙女'
    || title === '表侄孙'
    || title === '表侄孙女'
}

export function runFullMatrix43x42Test(): {
  totalPairs: number
  supported: number
  unsupported: number
  registryViolations: number
  forbiddenBareViolations: number
  invalidPathViolations: number
  generationMismatchViolations: number
  failures: string[]
} {
  const { nodes, edges, lineageMemberIds } = buildIntraLineage43Fixture()
  const graph = buildRelGraph(nodes, edges)
  const failures: string[] = []
  let supported = 0
  let unsupported = 0
  let registryViolations = 0
  let forbiddenBareViolations = 0
  let invalidPathViolations = 0
  let generationMismatchViolations = 0

  const totalPairs = lineageMemberIds.length * (lineageMemberIds.length - 1)

  for (const viewerId of lineageMemberIds) {
    for (const targetId of lineageMemberIds) {
      if (viewerId === targetId) continue

      const result = resolveCanonicalRelativeTitle(graph, viewerId, targetId)

      if (!isValidObjectivePathDescription(result.pathDescription)) {
        invalidPathViolations += 1
        if (failures.length < 20) {
          failures.push(
            `invalid-path ${viewerId}->${targetId}: path=${result.pathDescription || '空'}`
          )
        }
      }

      if (result.canonicalTitle) {
        const actualGap = bloodGenerationGap(graph, viewerId, targetId)
        if (isNephewNieceTitle(result.canonicalTitle)) {
          if (actualGap !== 1) {
            generationMismatchViolations += 1
            if (failures.length < 20) {
              failures.push(
                `generation-mismatch ${viewerId}->${targetId}: title=${result.canonicalTitle} requiresGap=1 actualGap=${actualGap ?? 'null'}`
              )
            }
          }
        } else if (isNephewGrandchildTitle(result.canonicalTitle)) {
          if (actualGap !== 2) {
            generationMismatchViolations += 1
            if (failures.length < 20) {
              failures.push(
                `generation-mismatch ${viewerId}->${targetId}: title=${result.canonicalTitle} requiresGap=2 actualGap=${actualGap ?? 'null'}`
              )
            }
          }
        } else if (actualGap === 0 && /^(堂|表)?侄(女)?$/.test(result.canonicalTitle)) {
          generationMismatchViolations += 1
          if (failures.length < 20) {
            failures.push(
              `generation-mismatch ${viewerId}->${targetId}: same-generation must not use ${result.canonicalTitle}`
            )
          }
        }

        if (FORBIDDEN_BARE_CANONICAL_TITLES.has(result.canonicalTitle)) {
          forbiddenBareViolations += 1
          if (failures.length < 20) {
            failures.push(`forbidden-bare ${viewerId}->${targetId}: ${result.canonicalTitle}`)
          }
        } else if (!isOfficialCanonicalTitle(result.canonicalTitle)) {
          registryViolations += 1
          if (failures.length < 20) {
            failures.push(`unregistered ${viewerId}->${targetId}: ${result.canonicalTitle}`)
          }
        } else if (!result.unsupportedCanonicalTitle) {
          supported += 1
        }
      }

      if (result.unsupportedCanonicalTitle || !result.canonicalTitle) {
        unsupported += 1
      }
    }
  }

  if (registryViolations > 0) {
    failures.unshift(`注册表违规 ${registryViolations} 对`)
  }
  if (generationMismatchViolations > 0) {
    failures.unshift(`代际与称谓不符 ${generationMismatchViolations} 对`)
  }
  if (forbiddenBareViolations > 0) {
    failures.unshift(`裸称谓违规 ${forbiddenBareViolations} 对`)
  }
  if (invalidPathViolations > 0) {
    failures.unshift(`客观路径无效 ${invalidPathViolations} 对`)
  }

  return {
    totalPairs,
    supported,
    unsupported,
    registryViolations,
    forbiddenBareViolations,
    invalidPathViolations,
    generationMismatchViolations,
    failures
  }
}
