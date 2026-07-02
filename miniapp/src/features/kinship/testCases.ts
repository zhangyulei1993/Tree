import {
  canAppendRelation,
  getRelationOptions,
  PARENT_SPOUSE_DISABLED_REASON,
  CHILD_SIBLING_DISABLED_REASON,
  DIRECT_DESCENDANT_DEPTH_DISABLED_REASON,
  COLLATERAL_DESCENDANT_DEPTH_DISABLED_REASON,
  COUSIN_DESCENDANT_DEPTH_DISABLED_REASON,
  PARENT_CHILD_DISABLED_REASON
} from './allowedRelations'
import { inferRelativeAgeFromFacts, normalizePersonFacts } from './helpers'
import { deriveCanonicalExpectation, isForbiddenCanonicalTitle } from './canonicalTestExpectations'
import { resolveCanonicalKinship } from './resolveCanonicalKinship'
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
  expectedCandidatesIncludes?: string[]
  expectedCanAppend?: boolean
  expectedDisabledRelation?: KinshipRelation
  expectedDisabledReason?: string
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
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['姐妹的配偶', '姐夫', '妹夫'],
    expectedForbiddenTitles: ['兄弟姐妹的配偶']
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
    expectedStatus: 'resolved',
    expectedTitle: '外五世祖父',
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
    name: '母系四层女性终点使用外高祖母候选',
    context: createContext([
      parentStep('female'),
      parentStep('female'),
      parentStep('female'),
      parentStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '外高祖母',
    expectedForbiddenTitles: ['外高祖父', '外高外祖父', '高祖父'],
    expectedForbiddenAliases: ['母系高祖父'],
    expectedForbiddenCandidates: ['外高祖父', '外高外祖父', '母系男性高祖辈']
  },
  {
    id: 'T062',
    name: '母系五层女性终点使用外五世祖母候选',
    context: createContext([
      parentStep('female'),
      parentStep('female'),
      parentStep('female'),
      parentStep('female'),
      parentStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '外五世祖母',
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
  },

  // --- M19-3: 状态机与终点 ---
  {
    id: 'T064',
    name: '儿子 -> 妻子为终点',
    context: createContext([childStep('male'), spouseStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '儿媳',
    action: { relation: 'sibling' },
    expectedCanAppend: false
  },
  {
    id: 'T065',
    name: '哥哥 -> 妻子为终点',
    context: createContext([siblingStep('male', 'older'), spouseStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '嫂子',
    action: { relation: 'child' },
    expectedCanAppend: false
  },
  {
    id: 'T066',
    name: '父亲 -> 哥哥后仍可追加子女',
    context: createContext([parentStep('male'), siblingStep('male', 'older')]),
    action: { relation: 'child' },
    expectedCanAppend: true
  },
  {
    id: 'T067',
    name: '父亲 -> 哥哥 -> 儿子非终点',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'older'),
      childStep('male')
    ]),
    action: { relation: 'child' },
    expectedCanAppend: true
  },
  {
    id: 'T068',
    name: '女儿 -> 丈夫为终点',
    context: createContext([childStep('female'), spouseStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '女婿',
    action: { relation: 'spouse' },
    expectedCanAppend: false
  },
  {
    id: 'T069',
    name: '姐姐 -> 丈夫为终点',
    context: createContext([siblingStep('female', 'older'), spouseStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '姐夫',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T070',
    name: '弟弟 -> 妻子为终点',
    context: createContext([siblingStep('male', 'younger'), spouseStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '弟媳',
    action: { relation: 'sibling' },
    expectedCanAppend: false
  },
  {
    id: 'T071',
    name: '妹妹 -> 丈夫为终点',
    context: createContext([siblingStep('female', 'younger'), spouseStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '妹夫',
    action: { relation: 'child' },
    expectedCanAppend: false
  },
  {
    id: 'T072',
    name: '终点后配偶也禁用',
    context: createContext([childStep('male'), spouseStep('female')]),
    action: { relation: 'spouse' },
    expectedCanAppend: false
  },
  {
    id: 'T073',
    name: '五层与终点并存时五层优先',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('male')
    ]),
    action: { relation: 'parent' },
    expectedCanAppend: false
  },

  // --- 基础称谓补充 ---
  {
    id: 'T074',
    name: '丈夫 -> 公公',
    context: createContext([spouseStep('male'), parentStep('male')], { gender: 'female' }),
    expectedStatus: 'resolved',
    expectedTitle: '公公'
  },
  {
    id: 'T075',
    name: '妻子 -> 婆婆',
    context: createContext([spouseStep('female'), parentStep('female')], { gender: 'male' }),
    expectedStatus: 'resolved',
    expectedTitle: '岳母'
  },
  {
    id: 'T076',
    name: '妻子 -> 小姨子',
    context: createContext([spouseStep('female'), siblingStep('female', 'younger')]),
    expectedStatus: 'resolved',
    expectedTitle: '小姨子'
  },
  {
    id: 'T077',
    name: '丈夫 -> 大姑子',
    context: createContext([spouseStep('male'), siblingStep('female', 'older')]),
    expectedStatus: 'resolved',
    expectedTitle: '大姑子'
  },
  {
    id: 'T078',
    name: '父亲 -> 哥哥生日推导堂兄路径基础',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'older'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '堂兄'
  },

  // --- 祖辈旁系 ---
  {
    id: 'T079',
    name: '父亲 -> 父亲 -> 哥哥',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      siblingStep('male', 'older')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '伯祖父'
  },
  {
    id: 'T080',
    name: '父亲 -> 父亲 -> 弟弟',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      siblingStep('male', 'younger')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '叔祖父'
  },
  {
    id: 'T081',
    name: '父亲 -> 父亲 -> 姐姐',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      siblingStep('female', 'older')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '姑祖母'
  },
  {
    id: 'T082',
    name: '母亲 -> 父亲 -> 哥哥',
    context: createContext([
      parentStep('female'),
      parentStep('male'),
      siblingStep('male', 'older')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '外伯祖父'
  },
  {
    id: 'T083',
    name: '母亲 -> 母亲 -> 姐姐',
    context: createContext([
      parentStep('female'),
      parentStep('female'),
      siblingStep('female', 'older')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '外姨祖母'
  },
  {
    id: 'T084',
    name: '父亲 -> 父亲 -> 父亲 -> 母亲',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '高祖母'
  },
  {
    id: 'T085',
    name: '母亲 -> 父亲 -> 弟弟',
    context: createContext([
      parentStep('female'),
      parentStep('male'),
      siblingStep('male', 'younger')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '外叔祖父'
  },

  // --- 堂表规则 ---
  {
    id: 'T086',
    name: '父亲 -> 哥哥 -> 儿子',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'older'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '堂兄'
  },
  {
    id: 'T087',
    name: '父亲 -> 弟弟 -> 儿子',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'younger'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '堂弟'
  },
  {
    id: 'T088',
    name: '父亲 -> 哥哥 -> 女儿',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'older'),
      childStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '堂姐'
  },
  {
    id: 'T089',
    name: '父亲 -> 弟弟 -> 女儿',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'younger'),
      childStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '堂妹'
  },
  {
    id: 'T220',
    name: '堂亲反例：叔父支但同辈年长仍称堂兄',
    context: createContext(
      [
        parentStep('male'),
        siblingStep('male', 'younger'),
        childStep('male', { birthday: '1998-01-01' })
      ],
      { gender: 'male', birthday: '2000-01-01' }
    ),
    expectedStatus: 'resolved',
    expectedTitle: '堂兄',
    expectedForbiddenTitles: ['堂弟']
  },
  {
    id: 'T221',
    name: '表亲反例：舅父支但同辈年幼仍称表弟',
    context: createContext(
      [
        parentStep('female'),
        siblingStep('male', 'older'),
        childStep('male', { birthday: '2005-01-01' })
      ],
      { gender: 'male', birthday: '2000-01-01' }
    ),
    expectedStatus: 'resolved',
    expectedTitle: '表弟',
    expectedForbiddenTitles: ['表兄']
  },
  {
    id: 'T090',
    name: '父亲 -> 姐姐 -> 儿子',
    context: createContext([
      parentStep('male'),
      siblingStep('female', 'older'),
      childStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['表兄', '表弟']
  },
  {
    id: 'T091',
    name: '父亲 -> 姐姐 -> 女儿',
    context: createContext([
      parentStep('male'),
      siblingStep('female', 'older'),
      childStep('female')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['表姐', '表妹']
  },
  {
    id: 'T092',
    name: '母亲 -> 哥哥 -> 儿子',
    context: createContext([
      parentStep('female'),
      siblingStep('male', 'older'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '表兄'
  },
  {
    id: 'T093',
    name: '母亲 -> 弟弟 -> 儿子',
    context: createContext([
      parentStep('female'),
      siblingStep('male', 'younger'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '表弟'
  },
  {
    id: 'T094',
    name: '母亲 -> 姐姐 -> 女儿',
    context: createContext([
      parentStep('female'),
      siblingStep('female', 'older'),
      childStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '表姐'
  },
  {
    id: 'T095',
    name: '母亲 -> 妹妹 -> 女儿',
    context: createContext([
      parentStep('female'),
      siblingStep('female', 'younger'),
      childStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '表妹'
  },
  {
    id: 'T096',
    name: '父亲 -> 哥哥 -> 儿子长幼未知',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'unknown'),
      childStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['堂兄', '堂弟']
  },
  {
    id: 'T097',
    name: '母亲 -> 哥哥 -> 儿子长幼未知',
    context: createContext([
      parentStep('female'),
      siblingStep('male', 'unknown'),
      childStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['表兄', '表弟']
  },
  {
    id: 'T098',
    name: '父母未知 -> 哥哥 -> 儿子',
    context: createContext([
      parentStep('unknown'),
      siblingStep('male', 'older'),
      childStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['堂兄', '表弟', '表兄']
  },

  // --- 侄甥规则 ---
  {
    id: 'T099',
    name: '哥哥 -> 儿子 -> 儿子',
    context: createContext([
      siblingStep('male', 'older'),
      childStep('male'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '侄孙'
  },
  {
    id: 'T100',
    name: '哥哥 -> 儿子 -> 女儿',
    context: createContext([
      siblingStep('male', 'older'),
      childStep('male'),
      childStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '侄孙女'
  },
  {
    id: 'T101',
    name: '弟弟 -> 女儿 -> 儿子',
    context: createContext([
      siblingStep('male', 'younger'),
      childStep('female'),
      childStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['侄外孙', '侄孙', '侄甥后代']
  },
  {
    id: 'T102',
    name: '姐姐 -> 儿子 -> 儿子',
    context: createContext([
      siblingStep('female', 'older'),
      childStep('male'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '外甥孙'
  },
  {
    id: 'T103',
    name: '妹妹 -> 女儿 -> 女儿',
    context: createContext([
      siblingStep('female', 'younger'),
      childStep('female'),
      childStep('female')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['外甥外孙女', '外甥孙女', '外甥后代']
  },
  {
    id: 'T104',
    name: '兄弟姐妹未知 -> 子女 -> 子女',
    context: createContext([
      siblingStep('male', 'unknown'),
      childStep('unknown'),
      childStep('unknown')
    ]),
    expectedStatus: 'ambiguous'
  },

  // --- 姻亲规则 ---
  {
    id: 'T105',
    name: '妻子 -> 哥哥 -> 儿子',
    context: createContext([
      spouseStep('female'),
      siblingStep('male', 'older'),
      childStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['妻侄', '内侄', '姻亲晚辈']
  },
  {
    id: 'T106',
    name: '丈夫 -> 姐姐 -> 女儿',
    context: createContext([
      spouseStep('male'),
      siblingStep('female', 'older'),
      childStep('female')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['夫家外甥女', '姻亲晚辈']
  },
  {
    id: 'T107',
    name: '妻子 -> 父亲 -> 哥哥',
    context: createContext([
      spouseStep('female'),
      parentStep('male'),
      siblingStep('male', 'older')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['配偶父辈旁系', '配偶伯父']
  },
  {
    id: 'T108',
    name: '妻子 -> 姐姐 -> 儿子',
    context: createContext([
      spouseStep('female'),
      siblingStep('female', 'older'),
      childStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['姻亲晚辈', '姨侄', '内侄']
  },
  {
    id: 'T109',
    name: '丈夫 -> 弟弟 -> 女儿',
    context: createContext([
      spouseStep('male'),
      siblingStep('male', 'younger'),
      childStep('female')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['姻亲晚辈', '叔侄女']
  },
  {
    id: 'T110',
    name: '儿子 -> 妻子后称谓',
    context: createContext([childStep('male'), spouseStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '儿媳'
  },
  {
    id: 'T111',
    name: '女儿 -> 丈夫后称谓',
    context: createContext([childStep('female'), spouseStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '女婿'
  },

  // --- fallback 泛称 ---
  {
    id: 'T112',
    name: '父亲 -> 哥哥 -> 儿子 -> 儿子为堂侄',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'older'),
      childStep('male'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '堂侄'
  },
  {
    id: 'T113',
    name: '妻子 -> 哥哥 -> 儿子 -> 女儿 fallback',
    context: createContext([
      spouseStep('female'),
      siblingStep('male', 'older'),
      childStep('male'),
      childStep('female')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['姻亲晚辈', '配偶侄辈']
  },
  {
    id: 'T114',
    name: '儿子 -> 女儿 -> 儿子 fallback 曾孙辈',
    context: createContext([
      childStep('male'),
      childStep('female'),
      childStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['曾孙', '曾孙女', '曾孙辈']
  },
  {
    id: 'T115',
    name: '妻子 -> 父亲 -> 父亲 fallback 配偶祖辈',
    context: createContext([
      spouseStep('female'),
      parentStep('male'),
      parentStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['配偶祖父', '配偶祖辈', '姻亲长辈']
  },
  {
    id: 'T116',
    name: '哥哥 -> 妻子 -> 父亲 fallback 姻亲长辈',
    context: createContext([
      siblingStep('male', 'older'),
      spouseStep('female'),
      parentStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['姻亲长辈', '兄弟姐妹配偶的父母']
  },
  {
    id: 'T117',
    name: '父亲 -> 哥哥未选子女 fallback',
    context: createContext([parentStep('male'), siblingStep('male', 'older')]),
    expectedStatus: 'resolved',
    expectedTitle: '伯父'
  },
  {
    id: 'T118',
    name: '未知路径 fallback 泛称',
    context: createContext([
      childStep('male'),
      spouseStep('female'),
      siblingStep('male', 'older')
    ]),
    expectedStatus: 'unsupported',
    expectedOneOfTitles: ['亲属关系路径已记录']
  },
  {
    id: 'T119',
    name: '配偶旁系 fallback',
    context: createContext([
      spouseStep('female'),
      parentStep('female'),
      siblingStep('female', 'older')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['姻亲长辈旁系', '配偶姑母', '配偶姨母']
  },
  {
    id: 'T120',
    name: '兄弟姐妹配偶的父母 fallback',
    context: createContext([
      siblingStep('female', 'older'),
      spouseStep('male'),
      parentStep('female')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['姻亲长辈', '兄弟姐妹配偶的父母']
  },
  {
    id: 'T121',
    name: '父亲 -> 父亲 -> 哥哥长幼未知祖辈旁系',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      siblingStep('male', 'unknown')
    ]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['祖辈旁系亲属', '伯祖父', '叔祖父']
  },
  {
    id: 'T122',
    name: '母亲 -> 妹妹 -> 儿子',
    context: createContext([
      parentStep('female'),
      siblingStep('female', 'younger'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '表弟'
  },
  {
    id: 'T123',
    name: '弟弟 -> 儿子 -> 儿子',
    context: createContext([
      siblingStep('male', 'younger'),
      childStep('male'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '侄孙'
  },
  {
    id: 'T124',
    name: '妹妹 -> 儿子 -> 儿子',
    context: createContext([
      siblingStep('female', 'younger'),
      childStep('male'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '外甥孙'
  },
  {
    id: 'T125',
    name: '终点儿媳不可追加父母',
    context: createContext([childStep('male'), spouseStep('female')]),
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T126',
    name: '儿子 -> 配偶 unknown',
    context: createContext([childStep('male'), spouseStep('unknown')]),
    expectedStatus: 'ambiguous',
    expectedTitle: '子女的配偶',
    expectedCandidatesIncludes: ['子女的配偶', '儿媳', '女婿'],
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T127',
    name: '女儿 -> 配偶 unknown',
    context: createContext([childStep('female'), spouseStep('unknown')]),
    expectedStatus: 'ambiguous',
    expectedTitle: '子女的配偶',
    expectedCandidatesIncludes: ['子女的配偶', '儿媳', '女婿'],
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T128',
    name: '哥哥 -> 配偶 unknown',
    context: createContext([siblingStep('male', 'older'), spouseStep('unknown')]),
    expectedStatus: 'ambiguous',
    expectedTitle: '兄弟的配偶',
    expectedCandidatesIncludes: ['兄弟的配偶', '嫂子', '弟媳'],
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T129',
    name: '妹妹 -> 配偶 unknown',
    context: createContext([siblingStep('female', 'younger'), spouseStep('unknown')]),
    expectedStatus: 'ambiguous',
    expectedTitle: '姐妹的配偶',
    expectedCandidatesIncludes: ['姐妹的配偶', '姐夫', '妹夫'],
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T130',
    name: '兄弟长幼 unknown -> 妻子',
    context: createContext([siblingStep('male', 'unknown'), spouseStep('female')]),
    expectedStatus: 'ambiguous',
    expectedCandidatesIncludes: ['嫂子', '弟媳'],
    expectedForbiddenTitles: ['兄弟姐妹的配偶']
  },
  {
    id: 'T131',
    name: '姐妹长幼 unknown -> 丈夫',
    context: createContext([siblingStep('female', 'unknown'), spouseStep('male')]),
    expectedStatus: 'ambiguous',
    expectedCandidatesIncludes: ['姐夫', '妹夫'],
    expectedForbiddenTitles: ['兄弟姐妹的配偶']
  },
  {
    id: 'T132',
    name: 'sibling->spouse terminal 不影响堂表规则',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'older'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '堂兄',
    action: { relation: 'child' },
    expectedCanAppend: true
  },
  {
    id: 'T133',
    name: 'child->spouse terminal 不影响孙辈解析',
    context: createContext([childStep('male'), childStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '孙女',
    action: { relation: 'child' },
    expectedCanAppend: true
  },
  {
    id: 'T134',
    name: 'parent -> spouse 被禁用',
    context: createContext([parentStep('male')]),
    action: { relation: 'spouse' },
    expectedCanAppend: false,
    expectedDisabledRelation: 'spouse',
    expectedDisabledReason: PARENT_SPOUSE_DISABLED_REASON
  },
  {
    id: 'T135',
    name: '父亲 -> 配偶禁用原因正确',
    context: createContext([parentStep('male')]),
    action: { relation: 'spouse' },
    expectedCanAppend: false,
    expectedDisabledReason: PARENT_SPOUSE_DISABLED_REASON
  },
  {
    id: 'T136',
    name: '母亲 -> 配偶禁用原因正确',
    context: createContext([parentStep('female')]),
    action: { relation: 'spouse' },
    expectedCanAppend: false,
    expectedDisabledReason: PARENT_SPOUSE_DISABLED_REASON
  },
  {
    id: 'T137',
    name: 'parent -> sibling 仍允许',
    context: createContext([parentStep('male')]),
    action: { relation: 'sibling' },
    expectedCanAppend: true
  },
  {
    id: 'T138',
    name: 'parent -> sibling -> child 仍允许',
    context: createContext([parentStep('male'), siblingStep('male', 'older')]),
    action: { relation: 'child' },
    expectedCanAppend: true
  },
  {
    id: 'T139',
    name: 'parent -> sibling -> spouse 仍允许',
    context: createContext([parentStep('male'), siblingStep('male', 'older')]),
    action: { relation: 'spouse' },
    expectedCanAppend: true
  },
  {
    id: 'T140',
    name: 'spouse -> parent 仍允许',
    context: createContext([spouseStep('female')]),
    action: { relation: 'parent' },
    expectedCanAppend: true
  },
  {
    id: 'T141',
    name: 'spouse -> sibling 仍允许',
    context: createContext([spouseStep('female')]),
    action: { relation: 'sibling' },
    expectedCanAppend: true
  },
  {
    id: 'T142',
    name: 'sibling -> child 仍允许',
    context: createContext([siblingStep('male', 'older')]),
    action: { relation: 'child' },
    expectedCanAppend: true
  },
  {
    id: 'T143',
    name: 'sibling -> spouse 仍允许',
    context: createContext([siblingStep('male', 'older')]),
    action: { relation: 'spouse' },
    expectedCanAppend: true
  },
  {
    id: 'T144',
    name: 'child -> sibling 被禁用',
    context: createContext([childStep('male')]),
    action: { relation: 'sibling' },
    expectedCanAppend: false,
    expectedDisabledRelation: 'sibling',
    expectedDisabledReason: CHILD_SIBLING_DISABLED_REASON
  },
  {
    id: 'T145',
    name: '儿子 -> 兄弟姐妹禁用原因正确',
    context: createContext([childStep('male')]),
    action: { relation: 'sibling' },
    expectedCanAppend: false,
    expectedDisabledReason: CHILD_SIBLING_DISABLED_REASON
  },
  {
    id: 'T146',
    name: '从本人往下允许到重孙辈',
    context: createContext([childStep('male'), childStep('male')]),
    action: { relation: 'child' },
    expectedCanAppend: true
  },
  {
    id: 'T147',
    name: '旁系后代不受本人直系下行限制',
    context: createContext([siblingStep('male', 'younger'), childStep('male')]),
    action: { relation: 'child' },
    expectedCanAppend: true
  },
  {
    id: 'T148',
    name: '三代直系下行解析为曾孙辈',
    context: createContext([childStep('male'), childStep('male'), childStep('male')]),
    expectedStatus: 'ambiguous',
    expectedOneOfTitles: ['曾孙', '曾孙女', '曾孙辈']
  },
  {
    id: 'T149',
    name: '父亲 -> 哥哥 -> 配偶为伯母并终止',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'older'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '伯母',
    action: { relation: 'child' },
    expectedCanAppend: false
  },
  {
    id: 'T150',
    name: '父亲 -> 弟弟 -> 配偶为婶婶并终止',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'younger'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '婶婶',
    action: { relation: 'child' },
    expectedCanAppend: false
  },
  {
    id: 'T151',
    name: '父亲 -> 姐妹 -> 配偶为姑父并终止',
    context: createContext([
      parentStep('male'),
      siblingStep('female', 'older'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '姑父',
    action: { relation: 'child' },
    expectedCanAppend: false
  },
  {
    id: 'T152',
    name: '母亲 -> 兄弟 -> 配偶为舅妈并终止',
    context: createContext([
      parentStep('female'),
      siblingStep('male', 'older'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '舅妈',
    action: { relation: 'child' },
    expectedCanAppend: false
  },
  {
    id: 'T153',
    name: '母亲 -> 姐妹 -> 配偶为姨父并终止',
    context: createContext([
      parentStep('female'),
      siblingStep('female', 'younger'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '姨父',
    action: { relation: 'child' },
    expectedCanAppend: false
  },
  {
    id: 'T154',
    name: '父亲兄弟长幼未知 -> 配偶需要候选并终止',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'unknown'),
      spouseStep('female')
    ]),
    expectedStatus: 'ambiguous',
    expectedCandidatesIncludes: ['伯母', '婶婶'],
    action: { relation: 'child' },
    expectedCanAppend: false
  },
  {
    id: 'T155',
    name: '妻子 -> 哥哥 -> 配偶为内兄嫂并终止',
    context: createContext([
      spouseStep('female'),
      siblingStep('male', 'older'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '内兄嫂',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T156',
    name: '妻子 -> 弟弟 -> 配偶为内弟媳并终止',
    context: createContext([
      spouseStep('female'),
      siblingStep('male', 'younger'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '内弟媳',
    action: { relation: 'child' },
    expectedCanAppend: false
  },
  {
    id: 'T157',
    name: '妻子 -> 姐妹 -> 配偶为连襟并终止',
    context: createContext([
      spouseStep('female'),
      siblingStep('female', 'older'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '连襟',
    action: { relation: 'sibling' },
    expectedCanAppend: false
  },
  {
    id: 'T158',
    name: '丈夫 -> 哥哥 -> 配偶为妯娌并终止',
    context: createContext([
      spouseStep('male'),
      siblingStep('male', 'older'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '妯娌',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T159',
    name: '丈夫 -> 姐姐 -> 配偶为姑姐夫并终止',
    context: createContext([
      spouseStep('male'),
      siblingStep('female', 'older'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '姑姐夫',
    action: { relation: 'child' },
    expectedCanAppend: false
  },
  {
    id: 'T160',
    name: '丈夫 -> 妹妹 -> 配偶为姑妹夫并终止',
    context: createContext([
      spouseStep('male'),
      siblingStep('female', 'younger'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '姑妹夫',
    action: { relation: 'child' },
    expectedCanAppend: false
  },
  {
    id: 'T161',
    name: '哥哥 -> 儿子 -> 配偶为侄媳妇并终止',
    context: createContext([
      siblingStep('male', 'older'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '侄媳妇',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T162',
    name: '哥哥 -> 女儿 -> 配偶为侄女婿并终止',
    context: createContext([
      siblingStep('male', 'older'),
      childStep('female'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '侄女婿',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T163',
    name: '姐姐 -> 儿子 -> 配偶为外甥媳妇并终止',
    context: createContext([
      siblingStep('female', 'older'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '外甥媳妇',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T164',
    name: '姐姐 -> 女儿 -> 配偶为外甥女婿并终止',
    context: createContext([
      siblingStep('female', 'older'),
      childStep('female'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '外甥女婿',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T165',
    name: '父亲 -> 哥哥 -> 儿子 -> 配偶为堂嫂并终止',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'older'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '堂嫂',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T166',
    name: '父亲 -> 弟弟 -> 儿子 -> 配偶为堂弟媳并终止',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'younger'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '堂弟媳',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T167',
    name: '父亲 -> 姐姐 -> 女儿 -> 配偶为表姐夫并终止',
    context: createContext([
      parentStep('male'),
      siblingStep('female', 'older'),
      childStep('female'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '表姐夫',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T168',
    name: '母亲 -> 弟弟 -> 儿子 -> 配偶为表弟媳并终止',
    context: createContext([
      parentStep('female'),
      siblingStep('male', 'younger'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '表弟媳',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T169',
    name: '孙子 -> 配偶为孙媳并终止',
    context: createContext([
      childStep('male'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '孙媳',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T170',
    name: '孙女 -> 配偶为孙女婿并终止',
    context: createContext([
      childStep('male'),
      childStep('female'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '孙女婿',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T171',
    name: '外孙 -> 配偶为外孙媳并终止',
    context: createContext([
      childStep('female'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '外孙媳',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T172',
    name: '外孙女 -> 配偶为外孙女婿并终止',
    context: createContext([
      childStep('female'),
      childStep('female'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '外孙女婿',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T173',
    name: '妻子 -> 哥哥 -> 儿子 -> 配偶为妻侄媳妇并终止',
    context: createContext([
      spouseStep('female'),
      siblingStep('male', 'older'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '妻侄媳妇',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T174',
    name: '妻子 -> 弟弟 -> 女儿 -> 配偶为妻侄女婿并终止',
    context: createContext([
      spouseStep('female'),
      siblingStep('male', 'younger'),
      childStep('female'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '妻侄女婿',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T175',
    name: '妻子 -> 姐姐 -> 儿子 -> 配偶为妻外甥媳妇并终止',
    context: createContext([
      spouseStep('female'),
      siblingStep('female', 'older'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '妻外甥媳妇',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T176',
    name: '丈夫 -> 哥哥 -> 儿子 -> 配偶为夫侄媳妇并终止',
    context: createContext([
      spouseStep('male'),
      siblingStep('male', 'older'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '夫侄媳妇',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T177',
    name: '丈夫 -> 弟弟 -> 女儿 -> 配偶为夫侄女婿并终止',
    context: createContext([
      spouseStep('male'),
      siblingStep('male', 'younger'),
      childStep('female'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '夫侄女婿',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T178',
    name: '丈夫 -> 姐姐 -> 女儿 -> 配偶为夫外甥女婿并终止',
    context: createContext([
      spouseStep('male'),
      siblingStep('female', 'older'),
      childStep('female'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '夫外甥女婿',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T179',
    name: '从本人往下超过重孙辈被禁用',
    context: createContext([childStep('male'), childStep('male'), childStep('male')]),
    action: { relation: 'child' },
    expectedCanAppend: false,
    expectedDisabledReason: DIRECT_DESCENDANT_DEPTH_DISABLED_REASON
  },
  {
    id: 'T180',
    name: '曾孙 -> 配偶为曾孙媳并终止',
    context: createContext([
      childStep('male'),
      childStep('male'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '曾孙媳',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T181',
    name: '曾孙女 -> 配偶为曾孙女婿并终止',
    context: createContext([
      childStep('male'),
      childStep('male'),
      childStep('female'),
      spouseStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '曾孙女婿',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T182',
    name: '哥哥 -> 儿子 -> 儿子 -> 儿子为侄重孙',
    context: createContext([
      siblingStep('male', 'older'),
      childStep('male'),
      childStep('male'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '侄重孙'
  },
  {
    id: 'T183',
    name: '哥哥 -> 儿子 -> 儿子 -> 儿子 -> 配偶为侄重孙媳并终止',
    context: createContext([
      siblingStep('male', 'older'),
      childStep('male'),
      childStep('male'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '侄重孙媳',
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T184',
    name: '姐姐 -> 儿子 -> 儿子 -> 女儿为外甥重孙女',
    context: createContext([
      siblingStep('female', 'older'),
      childStep('male'),
      childStep('male'),
      childStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '外甥重孙女'
  },
  {
    id: 'T185',
    name: '父亲 -> 父亲 -> 父亲 -> 哥哥为曾伯祖父',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      siblingStep('male', 'older')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '曾伯祖父'
  },
  {
    id: 'T186',
    name: '父亲 -> 父亲 -> 父亲 -> 弟弟为曾叔祖父',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      siblingStep('male', 'younger')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '曾叔祖父'
  },
  {
    id: 'T187',
    name: '父亲 -> 父亲 -> 父亲 -> 姐妹为曾姑祖母',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      siblingStep('female', 'older')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '曾姑祖母'
  },
  {
    id: 'T188',
    name: '兄弟姐妹后不能再选父母',
    context: createContext([siblingStep('male', 'older')]),
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T189',
    name: '父母的父母不能往回选子女',
    context: createContext([parentStep('male'), parentStep('male')]),
    action: { relation: 'child' },
    expectedCanAppend: false,
    expectedDisabledReason: PARENT_CHILD_DISABLED_REASON
  },
  {
    id: 'T190',
    name: '父母的父母的父母不能往回选子女',
    context: createContext([parentStep('male'), parentStep('male'), parentStep('male')]),
    action: { relation: 'child' },
    expectedCanAppend: false,
    expectedDisabledReason: PARENT_CHILD_DISABLED_REASON
  },
  {
    id: 'T191',
    name: '同龄兄弟 -> 配偶命中通用终点',
    context: createContext([
      siblingStep('male', 'same'),
      spouseStep('female')
    ]),
    expectedStatus: 'ambiguous',
    expectedCandidatesIncludes: ['嫂子', '弟媳'],
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T192',
    name: '未知堂表亲 -> 配偶命中通用终点',
    context: createContext([
      parentStep('unknown'),
      siblingStep('unknown', 'same'),
      childStep('unknown'),
      spouseStep('unknown')
    ]),
    expectedStatus: 'ambiguous',
    expectedCandidatesIncludes: ['堂嫂', '表嫂'],
    action: { relation: 'parent' },
    expectedCanAppend: false
  },
  {
    id: 'T193',
    name: '高祖父 -> 哥哥为高伯祖父',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      siblingStep('male', 'older')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '高伯祖父'
  },
  {
    id: 'T194',
    name: '高祖父 -> 弟弟为高叔祖父',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      siblingStep('male', 'younger')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '高叔祖父'
  },
  {
    id: 'T195',
    name: '高祖父 -> 姐妹为高姑祖母',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      siblingStep('female', 'older')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '高姑祖母'
  },
  {
    id: 'T196',
    name: '同龄兄弟返回保守泛称',
    context: createContext([siblingStep('male', 'same')]),
    expectedStatus: 'ambiguous',
    expectedTitle: '兄弟',
    expectedCandidatesIncludes: ['双胞胎兄弟']
  },
  {
    id: 'T197',
    name: '同龄姐妹返回保守泛称',
    context: createContext([siblingStep('female', 'same')]),
    expectedStatus: 'ambiguous',
    expectedTitle: '姐妹',
    expectedCandidatesIncludes: ['双胞胎姐妹']
  },
  {
    id: 'T198',
    name: '未知性别年长兄弟姐妹返回泛称',
    context: createContext([siblingStep('unknown', 'older')]),
    expectedStatus: 'ambiguous',
    expectedTitle: '年长的兄弟姐妹',
    expectedCandidatesIncludes: ['哥哥', '姐姐']
  },
  {
    id: 'T199',
    name: '兄弟的未知性别子女返回侄子女',
    context: createContext([siblingStep('male', 'older'), childStep('unknown')]),
    expectedStatus: 'ambiguous',
    expectedTitle: '侄子女',
    expectedCandidatesIncludes: ['侄子', '侄女']
  },
  {
    id: 'T200',
    name: '姐妹的未知性别子女返回外甥子女',
    context: createContext([siblingStep('female', 'older'), childStep('unknown')]),
    expectedStatus: 'ambiguous',
    expectedTitle: '外甥子女',
    expectedCandidatesIncludes: ['外甥', '外甥女']
  },
  {
    id: 'T201',
    name: '儿子的未知性别子女返回孙辈',
    context: createContext([childStep('male'), childStep('unknown')]),
    expectedStatus: 'ambiguous',
    expectedTitle: '孙辈',
    expectedCandidatesIncludes: ['孙子', '孙女']
  },
  {
    id: 'T202',
    name: '女儿的未知性别子女返回外孙辈',
    context: createContext([childStep('female'), childStep('unknown')]),
    expectedStatus: 'ambiguous',
    expectedTitle: '外孙辈',
    expectedCandidatesIncludes: ['外孙', '外孙女']
  },
  {
    id: 'T203',
    name: '祖辈旁系晚辈配偶返回保守泛称',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      siblingStep('female', 'older'),
      childStep('male'),
      spouseStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedTitle: '祖辈旁系晚辈的配偶',
    expectedCandidatesIncludes: ['远房姻亲']
  },
  {
    id: 'T204',
    name: '侄重孙后禁止继续下探子女',
    context: createContext([
      siblingStep('male', 'older'),
      childStep('male'),
      childStep('male'),
      childStep('male')
    ]),
    action: { relation: 'child' },
    expectedCanAppend: false,
    expectedDisabledRelation: 'child',
    expectedDisabledReason: COLLATERAL_DESCENDANT_DEPTH_DISABLED_REASON
  },
  {
    id: 'T205',
    name: '侄重孙仍允许选择配偶',
    context: createContext([
      siblingStep('male', 'older'),
      childStep('male'),
      childStep('male'),
      childStep('male')
    ]),
    action: { relation: 'spouse' },
    expectedCanAppend: true
  },
  {
    id: 'T206',
    name: '同龄兄弟的儿子仍为侄子',
    context: createContext([siblingStep('male', 'same'), childStep('male')]),
    expectedStatus: 'resolved',
    expectedTitle: '侄子'
  },
  {
    id: 'T207',
    name: '同龄姐妹的女儿仍为外甥女',
    context: createContext([siblingStep('female', 'same'), childStep('female')]),
    expectedStatus: 'resolved',
    expectedTitle: '外甥女'
  },
  {
    id: 'T208',
    name: '堂兄的儿子为堂侄',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'older'),
      childStep('male'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '堂侄'
  },
  {
    id: 'T209',
    name: '表姐的女儿为表侄女',
    context: createContext([
      parentStep('female'),
      siblingStep('female', 'older'),
      childStep('female'),
      childStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '表侄女'
  },
  {
    id: 'T210',
    name: '堂侄的儿子为堂侄孙',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'older'),
      childStep('male'),
      childStep('male'),
      childStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '堂侄孙'
  },
  {
    id: 'T211',
    name: '表侄女的女儿为表侄孙女',
    context: createContext([
      parentStep('female'),
      siblingStep('female', 'older'),
      childStep('female'),
      childStep('female'),
      childStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '表侄孙女'
  },
  {
    id: 'T212',
    name: '堂表亲侄孙后禁止继续下探子女',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'older'),
      childStep('male'),
      childStep('male'),
      childStep('male')
    ]),
    action: { relation: 'child' },
    expectedCanAppend: false,
    expectedDisabledRelation: 'child',
    expectedDisabledReason: COUSIN_DESCENDANT_DEPTH_DISABLED_REASON
  },
  {
    id: 'T213',
    name: '堂表亲侄孙允许选择配偶',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'older'),
      childStep('male'),
      childStep('male'),
      childStep('male')
    ]),
    action: { relation: 'spouse' },
    expectedCanAppend: true
  },
  {
    id: 'T214',
    name: '堂侄孙配偶为堂侄孙媳',
    context: createContext([
      parentStep('male'),
      siblingStep('male', 'older'),
      childStep('male'),
      childStep('male'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '堂侄孙媳'
  },
  {
    id: 'T215',
    name: '表侄孙配偶为表侄孙媳',
    context: createContext([
      parentStep('female'),
      siblingStep('female', 'older'),
      childStep('female'),
      childStep('male'),
      childStep('male'),
      spouseStep('female')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '表侄孙媳'
  },
  {
    id: 'T216',
    name: '父亲 -> 母亲 -> 父亲为曾外祖父',
    context: createContext([
      parentStep('male'),
      parentStep('female'),
      parentStep('male')
    ]),
    expectedStatus: 'resolved',
    expectedTitle: '曾外祖父'
  },
  {
    id: 'T217',
    name: '父亲 -> 母亲 -> 父亲 -> 父亲为高外祖父',
    context: createContext([
      parentStep('male'),
      parentStep('female'),
      parentStep('male'),
      parentStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedTitle: '高外祖父',
    expectedCandidatesIncludes: ['祖母的祖父']
  },
  {
    id: 'T218',
    name: '父亲 -> 母亲 -> 父亲 -> 母亲 -> 父亲为父系外系五世祖父',
    context: createContext([
      parentStep('male'),
      parentStep('female'),
      parentStep('male'),
      parentStep('female'),
      parentStep('male')
    ]),
    expectedStatus: 'ambiguous',
    expectedTitle: '父系外系五世祖父',
    expectedCandidatesIncludes: ['五世外祖父']
  },
  {
    id: 'T219',
    name: '父亲 -> 父亲 -> 父亲 -> 母亲 -> 母亲为父系外系五世祖母',
    context: createContext([
      parentStep('male'),
      parentStep('male'),
      parentStep('male'),
      parentStep('female'),
      parentStep('female')
    ]),
    expectedStatus: 'ambiguous',
    expectedTitle: '父系外系五世祖母',
    expectedCandidatesIncludes: ['五世外祖母']
  }
]

export interface KinshipSelfCheckResult {
  passed: number
  total: number
  failures: string[]
}

function titleMatches(resolution: ReturnType<typeof resolveCanonicalKinship>, testCase: KinshipTestCase) {
  const expected = deriveCanonicalExpectation(testCase)
  if (expected.kind === 'skip') return true
  if (expected.kind === 'unsupported') {
    return resolution.unsupportedCanonicalTitle
  }
  return resolution.canonicalTitle === expected.title
}

export function runKinshipSelfChecks(): KinshipSelfCheckResult {
  const failures: string[] = []

  for (const testCase of kinshipTestCases) {
    const context = testCase.context || createContext(testCase.steps || [])
    const expected = deriveCanonicalExpectation(testCase)

    if (expected.kind !== 'skip' && (testCase.expectedStatus || testCase.expectedTitle || testCase.expectedOneOfTitles)) {
      const resolution = resolveCanonicalKinship(context)

      if (resolution.canonicalTitle && isForbiddenCanonicalTitle(resolution.canonicalTitle)) {
        failures.push(`${testCase.id} ${testCase.name}: 不得返回裸称谓 ${resolution.canonicalTitle}`)
      }

      if (expected.kind === 'canonical') {
        if (resolution.unsupportedCanonicalTitle || resolution.canonicalTitle !== expected.title) {
          failures.push(
            `${testCase.id} ${testCase.name}: 期望规范称谓 ${expected.title}，实际 ${resolution.canonicalTitle || resolution.pathDescription}`
          )
        }
      } else if (expected.kind === 'unsupported') {
        if (!resolution.unsupportedCanonicalTitle) {
          failures.push(
            `${testCase.id} ${testCase.name}: 期望无规范称谓（展示路径），实际 ${resolution.canonicalTitle || '已解析'}`
          )
        }
      }

      if (testCase.expectedForbiddenTitles?.length) {
        const title = resolution.canonicalTitle || ''
        const forbidden = testCase.expectedForbiddenTitles.find((item) => title === item)
        if (forbidden) {
          failures.push(`${testCase.id} ${testCase.name}: 不应返回称谓 ${forbidden}`)
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
      if (testCase.expectedDisabledReason && validation.reason !== testCase.expectedDisabledReason) {
        failures.push(
          `${testCase.id} ${testCase.name}: 期望禁用原因「${testCase.expectedDisabledReason}」，实际「${validation.reason || '无'}」`
        )
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
      const wifeBrother = resolveCanonicalKinship(
        createContext([spouseStep('female'), siblingStep('male', 'older')])
      )
      const husbandBrother = resolveCanonicalKinship(
        createContext([spouseStep('male'), siblingStep('male', 'older')])
      )
      if (!wifeBrother.unsupportedCanonicalTitle || !husbandBrother.unsupportedCanonicalTitle) {
        failures.push(`${testCase.id} ${testCase.name}: 配偶兄弟姐妹路径应无规范称谓`)
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
