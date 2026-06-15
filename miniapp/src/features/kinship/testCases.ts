import { canAppendRelation, getRelationOptions } from './allowedRelations'
import { inferRelativeAgeFromFacts, normalizePersonFacts } from './helpers'
import { resolveKinship } from './resolveKinship'
import type { KinshipContext, KinshipRelation, KinshipStep } from './types'
import { MAX_KINSHIP_DEPTH } from './types'

export interface KinshipTestCase {
  id: string
  name: string
  context?: KinshipContext
  steps?: KinshipStep[]
  action?: {
    relation: KinshipRelation
  }
  expectedStatus?: 'resolved' | 'ambiguous' | 'unsupported'
  expectedTitle?: string
  expectedOneOfTitles?: string[]
  expectedForbiddenTitles?: string[]
  expectedForbiddenAliases?: string[]
  expectedForbiddenCandidates?: string[]
  expectedCanAppend?: boolean
  expectedDisabledRelation?: KinshipRelation
}

const defaultSelf = { gender: 'male' as const }

function createContext(steps: KinshipStep[] = [], self = defaultSelf): KinshipContext {
  return { self, steps, maxDepth: MAX_KINSHIP_DEPTH }
}

function parentStep(
  gender: 'male' | 'female' | 'unknown' = 'unknown',
  facts: { birthday?: string; age?: number } = {}
): KinshipStep {
  return { relation: 'parent', person: { gender, ...facts } }
}

function childStep(
  gender: 'male' | 'female' | 'unknown' = 'unknown',
  facts: { birthday?: string; age?: number } = {}
): KinshipStep {
  return { relation: 'child', person: { gender, ...facts } }
}

function siblingStep(
  gender: 'male' | 'female' | 'unknown' = 'unknown',
  relativeAge: 'older' | 'younger' | 'same' | 'unknown' = 'unknown',
  facts: { birthday?: string; age?: number } = {}
): KinshipStep {
  return { relation: 'sibling', person: { gender, ...facts }, relativeAge }
}

function spouseStep(gender: 'male' | 'female' | 'unknown' = 'unknown'): KinshipStep {
  return { relation: 'spouse', person: { gender } }
}

export const kinshipTestCases: KinshipTestCase[] = [
  {
    id: 'T001',
    name: '空路径',
    context: createContext(),
    expectedStatus: 'resolved',
    expectedTitle: '我'
  },
  {
    id: 'T002',
    name: 'parent 一层',
    context: createContext([parentStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '父亲'
  },
  {
    id: 'T003',
    name: 'child 一层',
    context: createContext([childStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '女儿'
  },
  {
    id: 'T004',
    name: 'sibling 一层',
    context: createContext([siblingStep('male', 'older')]),
    expectedStatus: 'resolved',
    expectedTitle: '哥哥'
  },
  {
    id: 'T005',
    name: 'spouse 一层',
    context: createContext([spouseStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '妻子'
  },
  {
    id: 'T006',
    name: '父亲 -> 母亲',
    context: createContext([parentStep('male'), parentStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '祖母'
  },
  {
    id: 'T007',
    name: '父亲 -> 哥哥',
    context: createContext([parentStep('male'), siblingStep('male', 'older')]),
    expectedStatus: 'resolved',
    expectedTitle: '伯父'
  },
  {
    id: 'T008',
    name: 'sibling -> child',
    context: createContext([siblingStep('male'), childStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '侄子'
  },
  {
    id: 'T009',
    name: 'sibling -> spouse',
    context: createContext([siblingStep('female'), spouseStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '兄弟姐妹的配偶'
  },
  {
    id: 'T010',
    name: 'spouse -> parent',
    context: createContext([spouseStep('female'), parentStep('male')], { gender: 'male' }),
    expectedStatus: 'resolved',
    expectedTitle: '岳父'
  },
  {
    id: 'T011',
    name: '妻子 -> 哥哥',
    context: createContext([spouseStep('female'), siblingStep('male', 'older')]),
    expectedStatus: 'resolved',
    expectedTitle: '大舅子'
  },
  {
    id: 'T012',
    name: 'spouse -> child',
    context: createContext([spouseStep('male'), childStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '继女'
  },
  {
    id: 'T013',
    name: '儿子 -> 女儿',
    context: createContext([childStep('male'), childStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '孙女'
  },
  {
    id: 'T014',
    name: 'parent -> child 应禁用',
    context: createContext([parentStep('male')]),
    action: { relation: 'child' },
    expectedCanAppend: false,
    expectedDisabledRelation: 'child'
  },
  {
    id: 'T015',
    name: 'parent -> parent -> child 应禁用',
    context: createContext([parentStep('male'), parentStep('female')]),
    action: { relation: 'child' },
    expectedCanAppend: false,
    expectedDisabledRelation: 'child'
  },
  {
    id: 'T016',
    name: 'child -> parent 应禁用',
    context: createContext([childStep('male')]),
    action: { relation: 'parent' },
    expectedCanAppend: false,
    expectedDisabledRelation: 'parent'
  },
  {
    id: 'T017',
    name: 'child -> child -> parent 应禁用',
    context: createContext([childStep('male'), childStep('female')]),
    action: { relation: 'parent' },
    expectedCanAppend: false,
    expectedDisabledRelation: 'parent'
  },
  {
    id: 'T018',
    name: 'sibling -> parent 应禁用',
    context: createContext([siblingStep('male')]),
    action: { relation: 'parent' },
    expectedCanAppend: false,
    expectedDisabledRelation: 'parent'
  },
  {
    id: 'T019',
    name: 'sibling -> sibling 应禁用',
    context: createContext([siblingStep('male')]),
    action: { relation: 'sibling' },
    expectedCanAppend: false,
    expectedDisabledRelation: 'sibling'
  },
  {
    id: 'T020',
    name: 'spouse -> spouse 应禁用',
    context: createContext([spouseStep('female')]),
    action: { relation: 'spouse' },
    expectedCanAppend: false,
    expectedDisabledRelation: 'spouse'
  },
  {
    id: 'T021',
    name: '五层 parent 允许',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('male')
    ]),
    action: { relation: 'parent' },
    expectedCanAppend: true
  },
  {
    id: 'T022',
    name: '五层后所有关系禁用',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('male')
    ]),
    expectedCanAppend: false
  },
  {
    id: 'T023',
    name: '五层撤销后可继续',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('male')
    ]),
    action: { relation: 'parent' },
    expectedCanAppend: true
  },
  {
    id: 'T024',
    name: '禁止追加不改变原路径',
    context: createContext([parentStep('male')]),
    action: { relation: 'child' },
    expectedCanAppend: false
  },
  {
    id: 'T025',
    name: 'unknown gender 返回中性或 ambiguous',
    context: createContext([parentStep('unknown')]),
    expectedStatus: 'resolved',
    expectedTitle: '父母'
  },
  {
    id: 'T026',
    name: 'sibling 未知长幼返回 ambiguous',
    context: createContext([siblingStep('male', 'unknown')]),
    expectedStatus: 'ambiguous'
  },
  {
    id: 'T027',
    name: 'birthday 与 age 同时存在时生日优先',
    context: createContext()
  },
  {
    id: 'T028',
    name: '母亲 -> 父亲',
    context: createContext([parentStep('female'), parentStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '外祖父'
  },
  {
    id: 'T029',
    name: '母亲 -> 母亲',
    context: createContext([parentStep('female'), parentStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '外祖母'
  },
  {
    id: 'T030',
    name: '父亲 -> 父亲',
    context: createContext([parentStep('male'), parentStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '祖父'
  },
  {
    id: 'T031',
    name: '母亲 -> 姐姐',
    context: createContext([parentStep('female'), siblingStep('female', 'older')]),
    expectedStatus: 'resolved',
    expectedTitle: '姨母'
  },
  {
    id: 'T032',
    name: '母亲 -> 弟弟',
    context: createContext([parentStep('female'), siblingStep('male', 'younger')]),
    expectedStatus: 'resolved',
    expectedTitle: '舅父'
  },
  {
    id: 'T033',
    name: '父亲 -> 姐姐',
    context: createContext([parentStep('male'), siblingStep('female', 'older')]),
    expectedStatus: 'resolved',
    expectedTitle: '姑母'
  },
  {
    id: 'T034',
    name: '父亲 -> 弟弟',
    context: createContext([parentStep('male'), siblingStep('male', 'younger')]),
    expectedStatus: 'resolved',
    expectedTitle: '叔父'
  },
  {
    id: 'T035',
    name: '女儿 -> 儿子',
    context: createContext([childStep('female'), childStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '外孙'
  },
  {
    id: 'T036',
    name: '女儿 -> 女儿',
    context: createContext([childStep('female'), childStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '外孙女'
  },
  {
    id: 'T037',
    name: '儿子 -> 儿子',
    context: createContext([childStep('male'), childStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '孙子'
  },
  {
    id: 'T038',
    name: '儿子 -> 女儿',
    context: createContext([childStep('male'), childStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '孙女'
  },
  {
    id: 'T039',
    name: '妻子 -> 弟弟',
    context: createContext([spouseStep('female'), siblingStep('male', 'younger')]),
    expectedStatus: 'resolved',
    expectedTitle: '小舅子'
  },
  {
    id: 'T040',
    name: '丈夫 -> 哥哥',
    context: createContext([spouseStep('male'), siblingStep('male', 'older')]),
    expectedStatus: 'resolved',
    expectedTitle: '大伯子'
  },
  {
    id: 'T041',
    name: '丈夫 -> 弟弟',
    context: createContext([spouseStep('male'), siblingStep('male', 'younger')]),
    expectedStatus: 'resolved',
    expectedTitle: '小叔子'
  },
  {
    id: 'T042',
    name: '妻子与丈夫兄弟称谓不同',
    context: createContext([spouseStep('female'), siblingStep('male', 'older')]),
    expectedStatus: 'resolved',
    expectedTitle: '大舅子'
  },
  {
    id: 'T043',
    name: '父亲生日推导伯父',
    context: createContext([
      parentStep('male', { birthday: '1970-01-01' }),
      siblingStep('male', 'unknown', { birthday: '1968-01-01' })
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '伯父'
  },
  {
    id: 'T044',
    name: '父亲生日推导叔父',
    context: createContext([
      parentStep('male', { birthday: '1970-01-01' }),
      siblingStep('male', 'unknown', { birthday: '1975-01-01' })
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '叔父'
  },
  {
    id: 'T045',
    name: '本人生日推导哥哥',
    context: createContext(
      [siblingStep('male', 'unknown', { birthday: '1998-01-01' })],
      { gender: 'unknown', birthday: '2000-01-01' }
    ),
    expectedStatus: 'resolved',
    expectedTitle: '哥哥'
  },
  {
    id: 'T046',
    name: '本人年龄推导妹妹',
    context: createContext(
      [siblingStep('female', 'unknown', { age: 18 })],
      { gender: 'unknown', age: 20 }
    ),
    expectedStatus: 'resolved',
    expectedTitle: '妹妹'
  },
  {
    id: 'T047',
    name: '生日优先于年龄推导哥哥',
    context: createContext(
      [siblingStep('male', 'unknown', { birthday: '1998-01-01', age: 10 })],
      { gender: 'unknown', birthday: '2000-01-01', age: 50 }
    ),
    expectedStatus: 'resolved',
    expectedTitle: '哥哥'
  },
  {
    id: 'T048',
    name: '父亲 -> 父亲 -> 父亲',
    context: createContext([parentStep('male'), parentStep('male'), parentStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '曾祖父'
  },
  {
    id: 'T049',
    name: '父亲 -> 父亲 -> 母亲',
    context: createContext([parentStep('male'), parentStep('male'), parentStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '曾祖母'
  },
  {
    id: 'T050',
    name: '母亲 -> 父亲 -> 父亲',
    context: createContext([parentStep('female'), parentStep('male'), parentStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '外曾祖父',
    expectedForbiddenTitles: ['曾祖父']
  },
  {
    id: 'T051',
    name: '母亲 -> 母亲 -> 父亲',
    context: createContext([parentStep('female'), parentStep('female'), parentStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '外曾外祖父',
    expectedForbiddenTitles: ['曾祖父', '高祖父']
  },
  {
    id: 'T052',
    name: 'unknown parent -> 父亲 -> 父亲',
    context: createContext([parentStep('unknown'), parentStep('male'), parentStep('male')]),
    expectedStatus: 'ambiguous',
    expectedForbiddenTitles: ['曾祖父']
  },
  {
    id: 'T053',
    name: '父亲 -> unknown parent -> 父亲',
    context: createContext([parentStep('male'), parentStep('unknown'), parentStep('male')]),
    expectedStatus: 'ambiguous',
    expectedForbiddenTitles: ['曾祖父', '高祖父']
  },
  {
    id: 'T054',
    name: '父亲 -> 父亲 -> 父亲 -> 父亲',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '高祖父'
  },
  {
    id: 'T055',
    name: '母亲 -> 母亲 -> 母亲 -> 父亲',
    context: createContext([
      parentStep('female'),
      parentStep('female'),
      parentStep('female'),
      parentStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedForbiddenTitles: ['高祖父', '曾祖父']
  },
  {
    id: 'T056',
    name: '父亲五层祖辈链',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '五世祖父'
  },
  {
    id: 'T057',
    name: '母亲四层后接父亲不为父系五世祖父',
    context: createContext([
      parentStep('female'),
      parentStep('female'),
      parentStep('female'),
      parentStep('female'),
      parentStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedForbiddenTitles: ['五世祖父', '高祖父', '曾祖父']
  },
  {
    id: 'T058',
    name: '母亲 -> 母亲 -> 母亲保持女性称谓',
    context: createContext([parentStep('female'), parentStep('female'), parentStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '外曾外祖母',
    expectedForbiddenTitles: ['外曾祖父', '外曾外祖父'],
    expectedForbiddenAliases: ['母系曾祖父', '母系曾外祖父']
  },
  {
    id: 'T059',
    name: '母亲 -> 母亲 -> 父亲别称保持男性',
    context: createContext([parentStep('female'), parentStep('female'), parentStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '外曾外祖父',
    expectedForbiddenAliases: ['母系曾外祖母']
  },
  {
    id: 'T060',
    name: '母亲 -> 父亲 -> 母亲别称保持女性',
    context: createContext([parentStep('female'), parentStep('male'), parentStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '外曾祖母',
    expectedForbiddenAliases: ['母系曾祖父']
  },
  {
    id: 'T061',
    name: '母系四层女性终点使用中性主标题和女性候选',
    context: createContext([
      parentStep('female'),
      parentStep('female'),
      parentStep('female'),
      parentStep('female')
    ]),
    expectedStatus: 'ambiguous',
    expectedTitle: '祖辈亲属',
    expectedForbiddenTitles: ['外高祖父', '外高外祖父', '高祖父'],
    expectedForbiddenAliases: ['母系高祖父'],
    expectedForbiddenCandidates: ['外高祖父', '外高外祖父', '母系男性高祖辈']
  },
  {
    id: 'T062',
    name: '母系五层女性终点使用中性主标题和女性候选',
    context: createContext([
      parentStep('female'),
      parentStep('female'),
      parentStep('female'),
      parentStep('female'),
      parentStep('female')
    ]),
    expectedStatus: 'ambiguous',
    expectedTitle: '祖辈亲属',
    expectedForbiddenTitles: ['外五世祖父', '外五世外祖父', '五世祖父'],
    expectedForbiddenAliases: ['母系五世祖父'],
    expectedForbiddenCandidates: ['外五世祖父', '外五世外祖父', '母系男性五世祖辈']
  },
  {
    id: 'T063',
    name: '母系多层未知终点保持 ambiguous',
    context: createContext([parentStep('female'), parentStep('female'), parentStep('unknown')]),
    expectedStatus: 'ambiguous',
    expectedTitle: '祖辈亲属'
  }
]

export interface KinshipSelfCheckResult {
  passed: number
  total: number
  failures: string[]
}

function titleMatches(resolution: ReturnType<typeof resolveKinship>, testCase: KinshipTestCase) {
  if (testCase.expectedOneOfTitles?.length) {
    return testCase.expectedOneOfTitles.includes(resolution.primaryTitle || '')
  }
  return resolution.primaryTitle === testCase.expectedTitle
}

export function runKinshipSelfChecks(): KinshipSelfCheckResult {
  const failures: string[] = []

  for (const testCase of kinshipTestCases) {
    const context = testCase.context || createContext(testCase.steps || [])

    if (testCase.expectedStatus || testCase.expectedTitle || testCase.expectedOneOfTitles) {
      const resolution = resolveKinship(context)
      if (testCase.expectedStatus && resolution.status !== testCase.expectedStatus) {
        failures.push(`${testCase.id} ${testCase.name}: 期望状态 ${testCase.expectedStatus}，实际 ${resolution.status}`)
      }
      if ((testCase.expectedTitle || testCase.expectedOneOfTitles) && !titleMatches(resolution, testCase)) {
        failures.push(
          `${testCase.id} ${testCase.name}: 期望称谓 ${testCase.expectedTitle || testCase.expectedOneOfTitles?.join('/')}，实际 ${resolution.primaryTitle || '无'}`
        )
      }
      if (testCase.expectedForbiddenTitles?.length) {
        const title = resolution.primaryTitle || ''
        const forbidden = testCase.expectedForbiddenTitles.find((item) => title === item)
        if (forbidden) {
          failures.push(`${testCase.id} ${testCase.name}: 不应返回称谓 ${forbidden}`)
        }
      }
      if (testCase.expectedForbiddenAliases?.length) {
        const forbidden = testCase.expectedForbiddenAliases.find((item) => resolution.aliases.includes(item))
        if (forbidden) {
          failures.push(`${testCase.id} ${testCase.name}: aliases 不应包含 ${forbidden}`)
        }
      }
      if (testCase.expectedForbiddenCandidates?.length) {
        const forbidden = testCase.expectedForbiddenCandidates.find((item) =>
          resolution.candidates?.includes(item)
        )
        if (forbidden) {
          failures.push(`${testCase.id} ${testCase.name}: candidates 不应包含 ${forbidden}`)
        }
      }
    }

    if (testCase.action) {
      const validation = canAppendRelation(context, testCase.action.relation)
      if (testCase.expectedCanAppend !== undefined && validation.valid !== testCase.expectedCanAppend) {
        failures.push(
          `${testCase.id} ${testCase.name}: 期望可追加 ${testCase.expectedCanAppend}，实际 ${validation.valid}`
        )
      }
      if (testCase.expectedDisabledRelation) {
        const option = getRelationOptions(context).find((item) => item.relation === testCase.expectedDisabledRelation)
        if (!option || option.enabled) {
          failures.push(`${testCase.id} ${testCase.name}: 期望禁用 ${testCase.expectedDisabledRelation}`)
        }
      }
      if (testCase.id === 'T024') {
        const originalLength = context.steps.length
        const originalSnapshot = JSON.stringify(context.steps)
        if (validation.valid) {
          failures.push(`${testCase.id} ${testCase.name}: 非法追加不应通过校验`)
        }
        if (context.steps.length !== originalLength) {
          failures.push(`${testCase.id} ${testCase.name}: 校验后路径长度发生变化`)
        }
        if (JSON.stringify(context.steps) !== originalSnapshot) {
          failures.push(`${testCase.id} ${testCase.name}: 校验后路径内容发生变化`)
        }
      }
    }

    if (testCase.id === 'T022') {
      const options = getRelationOptions(context)
      if (options.some((item) => item.enabled)) {
        failures.push(`${testCase.id} ${testCase.name}: 五层后仍有可用关系`)
      }
    }

    if (testCase.id === 'T027') {
      const normalized = normalizePersonFacts({
        gender: 'male',
        birthday: '1990-01-01',
        age: 30
      })
      if (normalized.birthday !== '1990-01-01' || normalized.age !== undefined) {
        failures.push(`${testCase.id} ${testCase.name}: 生日未优先保留`)
      }
    }

    if (testCase.id === 'T042') {
      const wifeBrother = resolveKinship(
        createContext([spouseStep('female'), siblingStep('male', 'older')])
      )
      const husbandBrother = resolveKinship(
        createContext([spouseStep('male'), siblingStep('male', 'older')])
      )
      if (wifeBrother.primaryTitle === husbandBrother.primaryTitle) {
        failures.push(`${testCase.id} ${testCase.name}: 妻子与丈夫的兄弟称谓不应相同`)
      }
    }
  }

  const failedIds = new Set(
    failures
      .map((message) => message.split(' ')[0])
      .filter((id) => kinshipTestCases.some((testCase) => testCase.id === id))
  )

  return {
    passed: kinshipTestCases.length - failedIds.size,
    total: kinshipTestCases.length,
    failures
  }
}

export function checkInferRelativeAgeFromFacts() {
  return {
    olderByBirthday: inferRelativeAgeFromFacts(
      { gender: 'male', birthday: '2000-01-01' },
      { gender: 'male', birthday: '1998-01-01' }
    ),
    youngerByAge: inferRelativeAgeFromFacts(
      { gender: 'unknown', age: 20 },
      { gender: 'female', age: 18 }
    ),
    birthdayPriority: inferRelativeAgeFromFacts(
      { gender: 'unknown', birthday: '2000-01-01', age: 50 },
      { gender: 'male', birthday: '1998-01-01', age: 10 }
    )
  }
}
