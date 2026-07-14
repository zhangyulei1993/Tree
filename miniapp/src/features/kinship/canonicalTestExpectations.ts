export interface KinshipTestExpectationInput {
  id?: string
  name?: string
  expectedTitle?: string
  expectedOneOfTitles?: string[]
  expectedStatus?: 'resolved' | 'ambiguous' | 'unsupported'
  expectedForbiddenTitles?: string[]
  action?: { relation: string }
}

const COLLOQUIAL_TO_CANONICAL: Record<string, string> = {
  哥哥: '兄长',
  堂哥: '堂兄',
  表哥: '表兄',
  婶婶: '婶母',
  舅妈: '舅母'
}

const NON_CANONICAL_EXPECTED = new Set([
  '大舅子',
  '小舅子',
  '大姨子',
  '小姨子',
  '大伯子',
  '小叔子',
  '大姑子',
  '小姑子',
  '岳父',
  '岳母',
  '公公',
  '婆婆',
  '继女',
  '继子',
  '大伯哥',
  '内兄',
  '内弟',
  '内兄嫂',
  '内弟媳',
  '连襟',
  '妯娌',
  '子女的配偶',
  '兄弟的配偶',
  '姐妹的配偶',
  '姑姐夫',
  '姑妹夫',
  '侄媳妇',
  '外甥媳妇',
  '外孙媳',
  '外孙女婿',
  '妻侄媳妇',
  '妻侄女婿',
  '妻外甥媳妇',
  '夫侄媳妇',
  '夫侄女婿',
  '夫外甥女婿',
  '亲属关系路径已记录',
  '侄重孙媳'
])

const FORBIDDEN_BARE = new Set([
  '亲属',
  '姻亲',
  '祖辈亲属',
  '父母辈亲属',
  '晚辈亲属',
  '同辈亲属',
  '祖辈旁系亲属',
  '父母',
  '子女',
  '兄弟姐妹',
  '配偶'
])

export type CanonicalExpectation =
  | { kind: 'canonical'; title: string }
  | { kind: 'unsupported'; incomplete?: boolean }
  | { kind: 'skip' }

function normalizeExpectedTitle(title: string): string {
  return COLLOQUIAL_TO_CANONICAL[title] || title
}

export function deriveCanonicalExpectation(testCase: KinshipTestExpectationInput): CanonicalExpectation {
  if (testCase.action && !testCase.expectedTitle && !testCase.expectedStatus) {
    return { kind: 'skip' }
  }

  if (testCase.expectedStatus === 'ambiguous' || testCase.expectedStatus === 'unsupported') {
    return { kind: 'unsupported' }
  }

  if (testCase.expectedTitle) {
    if (NON_CANONICAL_EXPECTED.has(testCase.expectedTitle) || FORBIDDEN_BARE.has(testCase.expectedTitle)) {
      return { kind: 'unsupported' }
    }
    return { kind: 'canonical', title: normalizeExpectedTitle(testCase.expectedTitle) }
  }

  if (testCase.expectedOneOfTitles?.length) {
    const canonicalCandidates = testCase.expectedOneOfTitles
      .map(normalizeExpectedTitle)
      .filter((title) => !FORBIDDEN_BARE.has(title) && !NON_CANONICAL_EXPECTED.has(title))
    if (canonicalCandidates.length === 1) {
      return { kind: 'canonical', title: canonicalCandidates[0] }
    }
    return { kind: 'unsupported' }
  }

  if (testCase.expectedStatus === 'resolved' && !testCase.expectedTitle) {
    return { kind: 'skip' }
  }

  return { kind: 'skip' }
}

export function isForbiddenCanonicalTitle(title: string | undefined): boolean {
  if (!title) return false
  return FORBIDDEN_BARE.has(title)
}
