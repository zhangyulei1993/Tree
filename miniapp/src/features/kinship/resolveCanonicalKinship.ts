import type { CanonicalKinshipResult } from './canonicalTypes'
import {
  buildCanonicalResult,
  buildUnsupportedCanonicalResult,
  CANONICAL_KINSHIP_TERMS,
  lookupCanonicalKeyByPathKey
} from './canonicalKinshipTerms'
import {
  isForbiddenBareCanonicalTitle,
  isOfficialCanonicalTitle
} from './officialCanonicalTitles'
import { matchAffinityRules } from './affinityRules'
import { matchAncestorRules } from './ancestorRules'
import { matchCousinPeerSeniority, matchCousinRules } from './cousinRules'
import { matchExactRules } from './exactRules'
import { matchNephewRules } from './nephewRules'
import { formatObjectivePathDescription, pathHasIncompleteInfo } from './objectivePath'
import { buildPathKey } from './pathKey'
import { validateContextStructure } from './terminalRules'
import type { KinshipContext, KinshipRuleMatch } from './types'

/** 口语/地区称谓 → 规范称谓；未列出且不在规范表中的视为不支持。 */
const COLLOQUIAL_TO_CANONICAL: Record<string, string> = {
  哥哥: '兄长',
  姐姐: '姐姐',
  堂哥: '堂兄',
  表哥: '表兄',
  婶婶: '婶母',
  奶奶: '祖母',
  爷爷: '祖父',
  外公: '外祖父',
  外婆: '外祖母'
}

const FORBIDDEN_BARE_TITLES = new Set([
  '亲属',
  '姻亲',
  '祖辈亲属',
  '父母辈亲属',
  '晚辈亲属',
  '同辈亲属',
  '祖辈旁系亲属',
  '兄弟姐妹',
  '父母',
  '子女',
  '配偶'
])

const NON_CANONICAL_SPOUSE_INLAW = new Set([
  '大舅子',
  '小舅子',
  '大姨子',
  '小姨子',
  '大伯子',
  '小叔子',
  '大伯哥',
  '内兄',
  '内弟',
  '岳父',
  '岳母',
  '公公',
  '婆婆',
  '丈人',
  '丈母娘',
  '继女',
  '继子'
])

function normalizeTitle(title: string): string {
  return COLLOQUIAL_TO_CANONICAL[title] || title
}

function isCanonicalTitle(title: string): boolean {
  return isOfficialCanonicalTitle(title)
}

function fromResolvedTitle(pathDescription: string, rawTitle: string): CanonicalKinshipResult | null {
  const title = normalizeTitle(rawTitle)
  if (FORBIDDEN_BARE_TITLES.has(title) || NON_CANONICAL_SPOUSE_INLAW.has(title)) {
    return null
  }
  if (isForbiddenBareCanonicalTitle(title) || !isCanonicalTitle(title)) {
    return null
  }
  return {
    canonicalTitle: title,
    unsupportedCanonicalTitle: false,
    incompleteInfo: false,
    pathDescription
  }
}

function fromLegacyMatch(
  pathDescription: string,
  match: KinshipRuleMatch | null
): CanonicalKinshipResult | null {
  if (!match || match.status !== 'resolved' || !match.primaryTitle) return null
  return fromResolvedTitle(pathDescription, match.primaryTitle)
}

export function resolveCanonicalKinship(context: KinshipContext): CanonicalKinshipResult {
  const pathDescription = formatObjectivePathDescription(context)
  const structure = validateContextStructure(context)
  if (!structure.valid) {
    return buildUnsupportedCanonicalResult(pathDescription, {
      incompleteInfo: true,
      explanation: structure.reason
    })
  }

  if (context.steps.length === 0) {
    return buildCanonicalResult('SELF', pathDescription)
  }

  if (pathHasIncompleteInfo(context)) {
    return buildUnsupportedCanonicalResult(pathDescription, {
      incompleteInfo: true,
      explanation: '关系信息不完整，请补充性别或长幼信息。'
    })
  }

  const pathKey = buildPathKey(context)
  const cousinPeerMatch = matchCousinPeerSeniority(context, pathKey)
  if (cousinPeerMatch) {
    const peerResult = fromLegacyMatch(pathDescription, cousinPeerMatch)
    if (peerResult) return peerResult
  }

  const canonicalKey = lookupCanonicalKeyByPathKey(pathKey)
  if (canonicalKey) {
    return buildCanonicalResult(canonicalKey, pathDescription)
  }

  const legacyMatchers = [
    () => fromLegacyMatch(pathDescription, matchExactRules(context, pathKey)),
    () => fromLegacyMatch(pathDescription, matchAncestorRules(context, pathKey)),
    () => fromLegacyMatch(pathDescription, matchCousinRules(context, pathKey)),
    () => fromLegacyMatch(pathDescription, matchNephewRules(context, pathKey)),
    () => fromLegacyMatch(pathDescription, matchAffinityRules(context, pathKey))
  ]

  for (const matcher of legacyMatchers) {
    const result = matcher()
    if (result) return result
  }

  return buildUnsupportedCanonicalResult(pathDescription, {
    incompleteInfo: false,
    explanation: '当前关系路径暂无对应规范称谓。'
  })
}
