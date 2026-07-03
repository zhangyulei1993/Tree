import { buildRelGraph } from '../family-tree/relativeGraph'
import type { RelGraph } from '../family-tree/relativeGraph'
import { resolveCanonicalRelativeTitle } from '../family-tree/relativeRules'
import { canAppendRelation } from './allowedRelations'
import { isOfficialCanonicalTitle } from './officialCanonicalTitles'
import { resolveCanonicalKinship } from './resolveCanonicalKinship'
import type { KinshipContext, KinshipStep } from './types'
import { MAX_KINSHIP_DEPTH } from './types'
import type { MemberType, TreeEdge, TreeNode } from '@/types/api'

const defaultSelf = { gender: 'male' as const }

function createContext(steps: KinshipStep[] = [], self = defaultSelf): KinshipContext {
  return { self, steps, maxDepth: MAX_KINSHIP_DEPTH }
}

function childStep(gender: 'male' | 'female'): KinshipStep {
  return { relation: 'child', person: { gender } }
}

function spouseStep(gender: 'male' | 'female'): KinshipStep {
  return { relation: 'spouse', person: { gender } }
}

function siblingStep(gender: 'male' | 'female', relativeAge: 'older' | 'younger'): KinshipStep {
  return { relation: 'sibling', person: { gender }, relativeAge }
}

function childChain(count: number, gender: 'male' | 'female' = 'male'): KinshipStep[] {
  return Array.from({ length: count }, () => childStep(gender))
}

interface DeepDescendantCase {
  name: string
  steps: KinshipStep[]
  expectedTitle: string
}

const PATH_CASES: DeepDescendantCase[] = [
  { name: 'child×3 male', steps: [...childChain(3, 'male')], expectedTitle: '曾孙' },
  { name: 'child×3 female', steps: [...childChain(2, 'male'), childStep('female')], expectedTitle: '曾孙女' },
  {
    name: 'child×3 male + spouse female',
    steps: [...childChain(3, 'male'), spouseStep('female')],
    expectedTitle: '曾孙媳妇'
  },
  {
    name: 'child×3 female + spouse male',
    steps: [...childChain(2, 'male'), childStep('female'), spouseStep('male')],
    expectedTitle: '曾孙女婿'
  },
  { name: 'child×4 male', steps: [...childChain(4, 'male')], expectedTitle: '玄孙' },
  {
    name: 'child×4 female',
    steps: [...childChain(3, 'male'), childStep('female')],
    expectedTitle: '玄孙女'
  },
  {
    name: 'child×4 male + spouse female',
    steps: [...childChain(4, 'male'), spouseStep('female')],
    expectedTitle: '玄孙媳妇'
  },
  {
    name: 'child×4 female + spouse male',
    steps: [...childChain(3, 'male'), childStep('female'), spouseStep('male')],
    expectedTitle: '玄孙女婿'
  },
  { name: 'child×5 male', steps: [...childChain(5, 'male')], expectedTitle: '来孙' },
  {
    name: 'child×5 female',
    steps: [...childChain(4, 'male'), childStep('female')],
    expectedTitle: '来孙女'
  },
  {
    name: 'child×5 male + spouse female',
    steps: [...childChain(5, 'male'), spouseStep('female')],
    expectedTitle: '来孙媳妇'
  },
  {
    name: 'child×5 female + spouse male',
    steps: [...childChain(4, 'male'), childStep('female'), spouseStep('male')],
    expectedTitle: '来孙女婿'
  }
]

interface ChainNodeSpec {
  gender: 'MALE' | 'FEMALE'
  memberType?: MemberType
}

function buildChainGraph(
  descendants: ChainNodeSpec[],
  options?: { spouse?: { gender: 'MALE' | 'FEMALE'; memberType?: MemberType } }
): { graph: RelGraph; meId: number; targetId: number } {
  const nodes: TreeNode[] = [{
    memberId: 1,
    displayName: '我',
    gender: 'MALE',
    memberType: 'LINEAGE_MEMBER',
    userBindingState: 'NOT_REQUIRED',
    canExpand: false
  }]
  const edges: TreeEdge[] = []
  let edgeId = 1
  let parentId = 1

  descendants.forEach((spec, index) => {
    const memberId = index + 2
    nodes.push({
      memberId,
      displayName: String(memberId),
      gender: spec.gender,
      memberType: spec.memberType || 'LINEAGE_MEMBER',
      userBindingState: 'NOT_REQUIRED',
      canExpand: false
    })
    edges.push({
      relationshipId: edgeId++,
      fromMemberId: parentId,
      toMemberId: memberId,
      relationshipType: 'PARENT_CHILD'
    })
    parentId = memberId
  })

  let targetId = parentId
  if (options?.spouse) {
    const spouseId = descendants.length + 2
    nodes.push({
      memberId: spouseId,
      displayName: String(spouseId),
      gender: options.spouse.gender,
      memberType: options.spouse.memberType || 'SPOUSE',
      userBindingState: 'NOT_REQUIRED',
      canExpand: false
    })
    edges.push({
      relationshipId: edgeId++,
      fromMemberId: parentId,
      toMemberId: spouseId,
      relationshipType: 'SPOUSE'
    })
    targetId = spouseId
  }

  return { graph: buildRelGraph(nodes, edges), meId: 1, targetId }
}

function buildDirectLineGraph(childDepth: number, lastGender: 'MALE' | 'FEMALE' = 'MALE') {
  const descendants = Array.from({ length: childDepth }, (_, index) => ({
    gender: (index === childDepth - 1 ? lastGender : 'MALE') as 'MALE' | 'FEMALE'
  }))
  return buildChainGraph(descendants)
}

function assertUnsupportedPath(
  name: string,
  steps: KinshipStep[],
  forbiddenTitles: string[],
  failures: string[]
): boolean {
  const result = resolveCanonicalKinship(createContext(steps))
  if (!result.unsupportedCanonicalTitle) {
    failures.push(`${name}: 应 unsupportedCanonicalTitle=true，实际为「${result.canonicalTitle || '(空)'}」`)
    return false
  }
  if (!result.pathDescription?.trim()) {
    failures.push(`${name}: 应存在客观路径描述`)
    return false
  }
  for (const forbidden of forbiddenTitles) {
    if (result.canonicalTitle === forbidden) {
      failures.push(`${name}: 不得返回「${forbidden}」`)
      return false
    }
  }
  return true
}

function assertUnsupportedGraph(
  name: string,
  graph: RelGraph,
  meId: number,
  targetId: number,
  forbiddenTitles: string[],
  failures: string[]
): boolean {
  const result = resolveCanonicalRelativeTitle(graph, meId, targetId)
  if (!result.unsupportedCanonicalTitle) {
    failures.push(`${name}: 应 unsupportedCanonicalTitle=true，实际为「${result.canonicalTitle || '(空)'}」`)
    return false
  }
  if (!result.pathDescription?.trim()) {
    failures.push(`${name}: 应存在客观路径描述`)
    return false
  }
  for (const forbidden of forbiddenTitles) {
    if (result.canonicalTitle === forbidden) {
      failures.push(`${name}: 不得返回「${forbidden}」`)
      return false
    }
  }
  return true
}

export function runDeepDirectDescendantKinshipTests(): {
  total: number
  passed: number
  failures: string[]
} {
  const failures: string[] = []
  let passed = 0
  let total = 0

  const countPass = (ok: boolean) => {
    total += 1
    if (ok) passed += 1
  }

  for (const testCase of PATH_CASES) {
    const result = resolveCanonicalKinship(createContext(testCase.steps))
    const title = result.canonicalTitle || ''
    if (
      result.unsupportedCanonicalTitle
      || title !== testCase.expectedTitle
      || !isOfficialCanonicalTitle(testCase.expectedTitle)
      || title === '曾孙媳'
    ) {
      failures.push(
        `${testCase.name}: expected「${testCase.expectedTitle}」, got「${title || result.pathDescription}」`
      )
    } else {
      passed += 1
    }
    total += 1
  }

  countPass(canAppendRelation(createContext(childChain(5, 'male')), 'spouse').valid)

  const spouseViewerGraph = buildRelGraph(
    [
      {
        memberId: 1,
        displayName: '我',
        gender: 'MALE',
        memberType: 'LINEAGE_MEMBER',
        userBindingState: 'NOT_REQUIRED',
        canExpand: false
      },
      {
        memberId: 2,
        displayName: '配偶',
        gender: 'FEMALE',
        memberType: 'SPOUSE',
        userBindingState: 'NOT_REQUIRED',
        canExpand: false
      }
    ],
    [{ relationshipId: 1, fromMemberId: 1, toMemberId: 2, relationshipType: 'SPOUSE' }]
  )
  const spouseViewer = resolveCanonicalRelativeTitle(spouseViewerGraph, 2, 1)
  countPass(spouseViewer.unsupportedCanonicalTitle === true)

  const patrilinealForbidden = ['来孙', '来孙媳妇', '来孙女婿', '玄孙媳妇', '曾孙媳妇', '曾孙媳']

  countPass(assertUnsupportedPath(
    '旁系 child×5',
    [siblingStep('male', 'older'), ...childChain(5, 'male')],
    patrilinealForbidden,
    failures
  ))

  countPass(assertUnsupportedPath(
    '外系 child×5 血亲',
    [
      childStep('female'),
      childStep('male'),
      childStep('male'),
      childStep('male'),
      childStep('male')
    ],
    ['来孙', '来孙女'],
    failures
  ))

  countPass(assertUnsupportedPath(
    '外系 child×5 末端配偶',
    [
      childStep('female'),
      childStep('male'),
      childStep('male'),
      childStep('male'),
      childStep('male'),
      spouseStep('female')
    ],
    patrilinealForbidden,
    failures
  ))

  countPass(assertUnsupportedPath(
    '混合路径 child:male>child:female>child:male + spouse',
    [
      childStep('male'),
      childStep('female'),
      childStep('male'),
      spouseStep('female')
    ],
    ['曾孙媳妇', '曾孙女婿', '曾孙媳'],
    failures
  ))

  for (const [childDepth, expectedTitle] of [
    [3, '曾孙'],
    [4, '玄孙'],
    [5, '来孙']
  ] as const) {
    const { graph, targetId } = buildDirectLineGraph(childDepth, 'MALE')
    const graphResult = resolveCanonicalRelativeTitle(graph, 1, targetId)
    if (graphResult.canonicalTitle !== expectedTitle) {
      failures.push(`graph child×${childDepth} male: expected「${expectedTitle}」, got「${graphResult.canonicalTitle || graphResult.pathDescription}」`)
      countPass(false)
    } else {
      countPass(true)
    }
  }

  const graphWithSpouse = buildChainGraph(
    [
      { gender: 'MALE' },
      { gender: 'MALE' },
      { gender: 'MALE' }
    ],
    { spouse: { gender: 'FEMALE', memberType: 'SPOUSE' } }
  )
  const spouseTitle = resolveCanonicalRelativeTitle(graphWithSpouse.graph, 1, graphWithSpouse.targetId)
  if (spouseTitle.canonicalTitle !== '曾孙媳妇' || spouseTitle.canonicalTitle === '曾孙媳') {
    failures.push(`graph 曾孙配偶: expected「曾孙媳妇」, got「${spouseTitle.canonicalTitle || spouseTitle.pathDescription}」`)
    countPass(false)
  } else {
    countPass(true)
  }

  const externalSpouseGraph = buildChainGraph(
    [
      { gender: 'FEMALE' },
      { gender: 'MALE' },
      { gender: 'MALE' },
      { gender: 'MALE' },
      { gender: 'MALE' }
    ],
    { spouse: { gender: 'FEMALE', memberType: 'SPOUSE' } }
  )
  countPass(assertUnsupportedGraph(
    'graph 外系 child×5 末端配偶',
    externalSpouseGraph.graph,
    1,
    externalSpouseGraph.targetId,
    patrilinealForbidden,
    failures
  ))

  const mixedSpouseGraph = buildChainGraph(
    [
      { gender: 'MALE' },
      { gender: 'FEMALE' },
      { gender: 'MALE' }
    ],
    { spouse: { gender: 'FEMALE', memberType: 'SPOUSE' } }
  )
  countPass(assertUnsupportedGraph(
    'graph 混合路径 child:male>child:female>child:male + spouse',
    mixedSpouseGraph.graph,
    1,
    mixedSpouseGraph.targetId,
    ['曾孙媳妇', '曾孙女婿', '曾孙媳'],
    failures
  ))

  const externalMemberGraph = buildChainGraph(
    [
      { gender: 'MALE' },
      { gender: 'MALE', memberType: 'EXTERNAL_MEMBER' },
      { gender: 'MALE' }
    ],
    { spouse: { gender: 'FEMALE', memberType: 'SPOUSE' } }
  )
  countPass(assertUnsupportedGraph(
    'graph 中间 EXTERNAL_MEMBER + spouse',
    externalMemberGraph.graph,
    1,
    externalMemberGraph.targetId,
    ['曾孙媳妇', '曾孙女婿', '曾孙媳'],
    failures
  ))

  const spouseMemberGraph = buildChainGraph(
    [
      { gender: 'MALE' },
      { gender: 'MALE', memberType: 'SPOUSE' },
      { gender: 'MALE' }
    ],
    { spouse: { gender: 'FEMALE', memberType: 'SPOUSE' } }
  )
  countPass(assertUnsupportedGraph(
    'graph 中间 SPOUSE memberType + spouse',
    spouseMemberGraph.graph,
    1,
    spouseMemberGraph.targetId,
    ['曾孙媳妇', '曾孙女婿', '曾孙媳'],
    failures
  ))

  const deepBloodForbidden = [
    '曾孙', '曾孙女', '玄孙', '玄孙女', '来孙', '来孙女',
    '曾孙媳妇', '曾孙女婿', '玄孙媳妇', '玄孙女婿', '来孙媳妇', '来孙女婿', '曾孙媳'
  ]

  const terminalExternalBloodGraph = buildChainGraph([
    { gender: 'MALE' },
    { gender: 'MALE' },
    { gender: 'MALE', memberType: 'EXTERNAL_MEMBER' }
  ])
  countPass(assertUnsupportedGraph(
    'graph 末端 bloodId=EXTERNAL_MEMBER',
    terminalExternalBloodGraph.graph,
    1,
    terminalExternalBloodGraph.targetId,
    deepBloodForbidden,
    failures
  ))

  const terminalSpouseBloodGraph = buildChainGraph([
    { gender: 'MALE' },
    { gender: 'MALE' },
    { gender: 'MALE', memberType: 'SPOUSE' }
  ])
  countPass(assertUnsupportedGraph(
    'graph 末端 bloodId=SPOUSE',
    terminalSpouseBloodGraph.graph,
    1,
    terminalSpouseBloodGraph.targetId,
    deepBloodForbidden,
    failures
  ))

  const lineageTargetSpouseGraph = buildChainGraph(
    [
      { gender: 'MALE' },
      { gender: 'MALE' },
      { gender: 'MALE' }
    ],
    { spouse: { gender: 'FEMALE', memberType: 'LINEAGE_MEMBER' } }
  )
  countPass(assertUnsupportedGraph(
    'graph targetId=LINEAGE_MEMBER 非 SPOUSE 配偶',
    lineageTargetSpouseGraph.graph,
    1,
    lineageTargetSpouseGraph.targetId,
    deepBloodForbidden,
    failures
  ))

  return { total, passed, failures }
}
