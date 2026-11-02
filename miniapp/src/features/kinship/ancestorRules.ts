import type { Gender, KinshipContext, KinshipRuleMatch, KinshipStep } from './types'
import { genderLabel } from './ruleUtils'

function getParentChainGenders(steps: KinshipStep[]): Gender[] {
  return steps.map((step) => step.person.gender || 'unknown')
}

function hasUnknownGenderInParentChain(genders: Gender[]): boolean {
  return genders.some((gender) => gender === 'unknown')
}

function countLeadingFemales(genders: Gender[]): number {
  let count = 0
  for (const gender of genders) {
    if (gender === 'female') count += 1
    else break
  }
  return count
}

function isPurePaternalChain(genders: Gender[]): boolean {
  if (genders[0] !== 'male' || hasUnknownGenderInParentChain(genders)) return false
  if (genders.every((gender) => gender === 'male')) return true
  return (
    genders[genders.length - 1] === 'female' &&
    genders.slice(0, -1).every((gender) => gender === 'male')
  )
}

const PATERNAL_ANCESTOR_TITLES: Record<number, { male: string; female: string; aliases?: string[] }> = {
  2: { male: '祖父', female: '祖母', aliases: ['爷爷', '奶奶'] },
  3: { male: '曾祖父', female: '曾祖母' },
  4: { male: '高祖父', female: '高祖母' },
  5: { male: '五世祖父', female: '五世祖母', aliases: ['天祖父'] }
}

const MATERNAL_ANCESTOR_TITLES: Record<
  number,
  { male: string; female: string; maleAliases?: string[]; femaleAliases?: string[] }
> = {
  2: { male: '外祖父', female: '外祖母', maleAliases: ['外公'], femaleAliases: ['外婆'] },
  3: {
    male: '外曾祖父',
    female: '外曾祖母',
    maleAliases: ['母系曾祖父'],
    femaleAliases: ['母系曾祖母']
  },
  4: {
    male: '外高祖父',
    female: '外高祖母',
    maleAliases: ['母系高祖父'],
    femaleAliases: ['母系高祖母']
  },
  5: {
    male: '外五世祖父',
    female: '外五世祖母',
    maleAliases: ['母系五世祖父'],
    femaleAliases: ['母系五世祖母']
  }
}

const MATERNAL_DOUBLE_FEMALE_TITLES: Record<
  number,
  { male: string; female: string; maleAliases?: string[]; femaleAliases?: string[] }
> = {
  3: {
    male: '外曾外祖父',
    female: '外曾外祖母',
    maleAliases: ['母系曾外祖父'],
    femaleAliases: ['母系曾外祖母']
  }
}

function getGenderedAliases(
  gender: Gender,
  config: { maleAliases?: string[]; femaleAliases?: string[]; aliases?: string[] }
): string[] {
  if (gender === 'male') return config.maleAliases || config.aliases || []
  if (gender === 'female') return config.femaleAliases || config.aliases || []
  return []
}

function resolveAmbiguousMaternalAncestor(
  lastGender: Gender,
  maleCandidates: string[],
  femaleCandidates: string[],
  explanation: string
): KinshipRuleMatch {
  if (lastGender === 'male') {
    return {
      status: 'ambiguous',
      primaryTitle: maleCandidates[0] || '母系男性祖辈',
      candidates: [...maleCandidates, '祖辈亲属'],
      explanation
    }
  }
  if (lastGender === 'female') {
    return {
      status: 'ambiguous',
      primaryTitle: femaleCandidates[0] || '母系女性祖辈',
      candidates: [...femaleCandidates, '祖辈亲属'],
      explanation
    }
  }
  return {
    status: 'ambiguous',
    primaryTitle: '母系祖辈',
    candidates: ['母系祖辈', ...maleCandidates, ...femaleCandidates],
    explanation: `${explanation}还需要补充终点性别。`
  }
}

function resolveUnknownAncestorChain(genders: Gender[]): KinshipRuleMatch {
  const lastGender = genders[genders.length - 1]
  const specificCandidates =
    lastGender === 'male'
      ? ['曾祖父', '外曾祖父', '其他祖辈称谓']
      : lastGender === 'female'
        ? ['曾祖母', '外曾祖母', '其他祖辈称谓']
        : ['曾祖父母', '外曾祖父母', '其他祖辈称谓']
  return {
    status: 'ambiguous',
    primaryTitle: '祖辈亲属',
    candidates: ['祖辈亲属', ...specificCandidates],
    explanation: '需要补充父系/母系方向才能确定祖辈称谓。'
  }
}

function resolveMaternalAncestorChain(genders: Gender[], depth: number): KinshipRuleMatch {
  const leadingFemales = countLeadingFemales(genders)
  const lastGender = genders[genders.length - 1]

  if (leadingFemales === 1) {
    const rest = genders.slice(1)
    const restIsPaternalMale =
      rest.every((gender) => gender === 'male') ||
      (rest[rest.length - 1] === 'female' &&
        rest.slice(0, -1).every((gender) => gender === 'male'))

    if (restIsPaternalMale) {
      const config = MATERNAL_ANCESTOR_TITLES[depth]
      if (config) {
        return {
          status: 'resolved',
          primaryTitle: genderLabel(lastGender, config.male, config.female, '外系祖辈'),
          aliases: getGenderedAliases(lastGender, config)
        }
      }
    }
  }

  if (leadingFemales >= 2 && depth === 3) {
    if (genders.every((gender) => gender === 'female')) {
      return {
        status: 'resolved',
        primaryTitle: '外曾外祖母',
        aliases: getGenderedAliases('female', MATERNAL_DOUBLE_FEMALE_TITLES[depth])
      }
    }
    const config = MATERNAL_DOUBLE_FEMALE_TITLES[depth]
    return {
      status: 'resolved',
      primaryTitle: genderLabel(lastGender, config.male, config.female, '外系祖辈'),
      aliases: getGenderedAliases(lastGender, config)
    }
  }

  if (leadingFemales >= 2 && depth >= 4) {
    return resolveAmbiguousMaternalAncestor(
      lastGender,
      depth === 5
        ? ['外五世祖父', '外五世外祖父', '母系男性五世祖辈']
        : ['外高祖父', '外高外祖父', '母系男性高祖辈'],
      depth === 5
        ? ['外五世祖母', '外五世外祖母', '母系女性五世祖辈']
        : ['外高祖母', '外高外祖母', '母系女性高祖辈'],
      '该祖辈路径涉及母系多代交叉，称谓存在地区差异，当前版本给出保守结果。'
    )
  }

  return resolveAmbiguousMaternalAncestor(
    lastGender,
    ['外曾祖父', '外曾外祖父', '母系男性祖辈'],
    ['外曾祖母', '外曾外祖母', '母系女性祖辈'],
    '该祖辈路径涉及母系方向，称谓存在地区差异，当前版本给出保守结果。'
  )
}

function resolvePaternalMixedAncestorChain(genders: Gender[], depth: number): KinshipRuleMatch {
  const lastGender = genders[genders.length - 1]

  if (depth === 3) {
    return {
      status: 'resolved',
      primaryTitle: genderLabel(lastGender, '曾外祖父', '曾外祖母', '父系外系曾祖辈'),
      aliases: lastGender === 'male'
        ? ['祖母的父亲']
        : lastGender === 'female'
          ? ['祖母的母亲']
          : []
    }
  }

  if (depth === 4) {
    if (lastGender === 'male') {
      return {
        status: 'ambiguous',
        primaryTitle: '高外祖父',
        candidates: ['高外祖父', '高外曾祖父', '祖母的祖父'],
        explanation: '该路径属于父系中的外系高辈祖先，不同地区称谓可能不同。'
      }
    }
    if (lastGender === 'female') {
      return {
        status: 'ambiguous',
        primaryTitle: '高外祖母',
        candidates: ['高外祖母', '高外曾祖母', '祖母的祖母'],
        explanation: '该路径属于父系中的外系高辈祖先，不同地区称谓可能不同。'
      }
    }
    return {
      status: 'ambiguous',
      primaryTitle: '父系外系高祖辈',
      candidates: ['高外祖父', '高外祖母', '父系外系高祖辈'],
      explanation: '需要补充终点性别以区分更具体的称谓。'
    }
  }

  if (depth === 5) {
    return {
      status: 'ambiguous',
      primaryTitle: genderLabel(lastGender, '父系外系五世祖父', '父系外系五世祖母', '父系外系五世祖辈'),
      candidates: lastGender === 'male'
        ? ['父系外系五世祖父', '五世外祖父', '高外祖父之父']
        : lastGender === 'female'
          ? ['父系外系五世祖母', '五世外祖母', '高外祖父之母']
          : ['父系外系五世祖父', '父系外系五世祖母', '父系外系五世祖辈'],
      explanation: '该路径属于父系中的外系五世祖辈，不同地区称谓可能不同，当前版本给出保守结果。'
    }
  }

  return {
    status: 'ambiguous',
    primaryTitle: '父系外系祖辈',
    candidates: ['父系外系祖辈', '外系祖辈'],
    explanation: '该祖辈路径涉及父系中的外系分支，称谓存在地区差异。'
  }
}

function resolveParentOnlyChain(steps: KinshipStep[]): KinshipRuleMatch | null {
  const genders = getParentChainGenders(steps)
  const depth = genders.length

  if (depth < 2) return null

  if (hasUnknownGenderInParentChain(genders)) {
    return resolveUnknownAncestorChain(genders)
  }

  if (isPurePaternalChain(genders)) {
    const config = PATERNAL_ANCESTOR_TITLES[depth]
    if (!config) return null
    const lastGender = genders[genders.length - 1]
    return {
      status: 'resolved',
      primaryTitle: genderLabel(lastGender, config.male, config.female, `${depth}世祖辈`),
      aliases: getGenderedAliases(lastGender, config)
    }
  }

  if (genders[0] === 'female') {
    return resolveMaternalAncestorChain(genders, depth)
  }

  return resolvePaternalMixedAncestorChain(genders, depth)
}

const ANCESTOR_COLLATERAL_RULES: Record<string, KinshipRuleMatch> = {
  'parent:male>parent:male>sibling:male:older': {
    status: 'resolved',
    primaryTitle: '伯祖父',
    aliases: ['伯公', '堂伯祖父']
  },
  'parent:male>parent:male>sibling:male:younger': {
    status: 'resolved',
    primaryTitle: '叔祖父',
    aliases: ['叔公', '堂叔祖父']
  },
  'parent:male>parent:male>sibling:female:older': {
    status: 'resolved',
    primaryTitle: '姑祖母',
    aliases: ['姑婆', '堂姑祖母']
  },
  'parent:male>parent:male>sibling:female:younger': {
    status: 'resolved',
    primaryTitle: '姑祖母',
    aliases: ['姑婆', '堂姑祖母']
  },
  'parent:female>parent:male>sibling:male:older': {
    status: 'ambiguous',
    primaryTitle: '外伯祖父',
    candidates: ['外伯祖父', '外族伯祖父', '母系伯祖父'],
    explanation: '母系祖辈旁系称谓存在地区差异，当前版本给出保守结果。'
  },
  'parent:female>parent:male>sibling:male:younger': {
    status: 'ambiguous',
    primaryTitle: '外叔祖父',
    candidates: ['外叔祖父', '外族叔祖父', '母系叔祖父'],
    explanation: '母系祖辈旁系称谓存在地区差异，当前版本给出保守结果。'
  },
  'parent:female>parent:female>sibling:female:older': {
    status: 'ambiguous',
    primaryTitle: '外姨祖母',
    candidates: ['外姨祖母', '外族姨祖母', '母系姨祖母'],
    explanation: '母系祖辈旁系称谓存在地区差异，当前版本给出保守结果。'
  },
  'parent:female>parent:female>sibling:female:younger': {
    status: 'ambiguous',
    primaryTitle: '外姨祖母',
    candidates: ['外姨祖母', '外族姨祖母', '母系姨祖母'],
    explanation: '母系祖辈旁系称谓存在地区差异，当前版本给出保守结果。'
  },
  'parent:male>parent:male>parent:male>sibling:male:older': {
    status: 'resolved',
    primaryTitle: '曾伯祖父'
  },
  'parent:male>parent:male>parent:male>sibling:male:younger': {
    status: 'resolved',
    primaryTitle: '曾叔祖父'
  },
  'parent:male>parent:male>parent:male>sibling:female:older': {
    status: 'resolved',
    primaryTitle: '曾姑祖母'
  },
  'parent:male>parent:male>parent:male>sibling:female:younger': {
    status: 'resolved',
    primaryTitle: '曾姑祖母'
  },
  'parent:male>parent:male>parent:female>sibling:male:older': {
    status: 'resolved',
    primaryTitle: '曾舅祖父'
  },
  'parent:male>parent:male>parent:female>sibling:male:younger': {
    status: 'resolved',
    primaryTitle: '曾舅祖父'
  },
  'parent:male>parent:male>parent:female>sibling:female:older': {
    status: 'resolved',
    primaryTitle: '曾姨祖母'
  },
  'parent:male>parent:male>parent:female>sibling:female:younger': {
    status: 'resolved',
    primaryTitle: '曾姨祖母'
  },
  'parent:male>parent:male>parent:male>parent:male>sibling:male:older': {
    status: 'resolved',
    primaryTitle: '高伯祖父'
  },
  'parent:male>parent:male>parent:male>parent:male>sibling:male:younger': {
    status: 'resolved',
    primaryTitle: '高叔祖父'
  },
  'parent:male>parent:male>parent:male>parent:male>sibling:female:older': {
    status: 'resolved',
    primaryTitle: '高姑祖母'
  },
  'parent:male>parent:male>parent:male>parent:male>sibling:female:younger': {
    status: 'resolved',
    primaryTitle: '高姑祖母'
  },
  'parent:male>parent:male>parent:male>parent:female>sibling:male:older': {
    status: 'resolved',
    primaryTitle: '高舅祖父'
  },
  'parent:male>parent:male>parent:male>parent:female>sibling:male:younger': {
    status: 'resolved',
    primaryTitle: '高舅祖父'
  },
  'parent:male>parent:male>parent:male>parent:female>sibling:female:older': {
    status: 'resolved',
    primaryTitle: '高姨祖母'
  },
  'parent:male>parent:male>parent:male>parent:female>sibling:female:younger': {
    status: 'resolved',
    primaryTitle: '高姨祖母'
  }
}

export function matchAncestorRules(context: KinshipContext, pathKey: string): KinshipRuleMatch | null {
  if (context.steps.every((step) => step.relation === 'parent')) {
    return resolveParentOnlyChain(context.steps)
  }

  if (pathKey in ANCESTOR_COLLATERAL_RULES) {
    return ANCESTOR_COLLATERAL_RULES[pathKey]
  }

  if (context.steps.length >= 3) {
    const prefix = context.steps.slice(0, 2).every((s) => s.relation === 'parent')
    const suffix = context.steps[2]?.relation === 'sibling'
    if (prefix && suffix) {
      const lineGender = context.steps[0].person.gender
      const siblingGender = context.steps[2].person.gender
      if (lineGender === 'male' && siblingGender === 'male') {
        return {
          status: 'ambiguous',
          primaryTitle: '祖辈旁系亲属',
          candidates: ['伯祖父', '叔祖父', '祖辈旁系男性'],
          explanation: '祖辈旁系需要补充长幼信息以区分伯祖父或叔祖父。'
        }
      }
      if (lineGender === 'female' && siblingGender === 'female') {
        return {
          status: 'ambiguous',
          primaryTitle: '祖辈旁系亲属',
          candidates: ['外姨祖母', '外姑祖母', '母系祖辈旁系女性'],
          explanation: '母系祖辈旁系称谓存在地区差异，当前版本给出保守结果。'
        }
      }
    }
  }

  return null
}

export function isParentOnlyPath(context: KinshipContext): boolean {
  return context.steps.length > 0 && context.steps.every((step) => step.relation === 'parent')
}
