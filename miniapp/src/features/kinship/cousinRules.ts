import type { KinshipContext, KinshipRuleMatch } from './types'
import { buildRelationKey } from './pathKey'

const COUSIN_RULES: Record<string, KinshipRuleMatch> = {
  'parent:male>sibling:male:older>child:male': { status: 'resolved', primaryTitle: '堂兄', aliases: ['堂哥'] },
  'parent:male>sibling:male:younger>child:male': { status: 'resolved', primaryTitle: '堂弟' },
  'parent:male>sibling:male:older>child:female': { status: 'resolved', primaryTitle: '堂姐' },
  'parent:male>sibling:male:younger>child:female': { status: 'resolved', primaryTitle: '堂妹' },
  'parent:male>sibling:male:unknown>child:male': {
    status: 'ambiguous',
    candidates: ['堂兄', '堂弟', '堂兄弟'],
    explanation: '堂亲兄弟需要补充长幼信息。'
  },
  'parent:male>sibling:male:unknown>child:female': {
    status: 'ambiguous',
    candidates: ['堂姐', '堂妹', '堂姐妹'],
    explanation: '堂亲姐妹需要补充长幼信息。'
  },
  'parent:male>sibling:female:older>child:male': {
    status: 'ambiguous',
    candidates: ['表兄', '表弟', '表兄弟'],
    explanation: '姑表兄弟称谓因长幼而异，请补充信息。'
  },
  'parent:male>sibling:female:younger>child:male': {
    status: 'ambiguous',
    candidates: ['表兄', '表弟', '表兄弟'],
    explanation: '姑表兄弟称谓因长幼而异，请补充信息。'
  },
  'parent:male>sibling:female:older>child:female': {
    status: 'ambiguous',
    candidates: ['表姐', '表妹', '表姐妹'],
    explanation: '姑表姐妹称谓因长幼而异，请补充信息。'
  },
  'parent:male>sibling:female:younger>child:female': {
    status: 'ambiguous',
    candidates: ['表姐', '表妹', '表姐妹'],
    explanation: '姑表姐妹称谓因长幼而异，请补充信息。'
  },
  'parent:male>sibling:female:unknown>child:unknown': {
    status: 'ambiguous',
    candidates: ['表兄', '表弟', '表姐', '表妹', '表亲'],
    explanation: '姑表亲需要补充子女性别和长幼信息。'
  },

  'parent:female>sibling:male:older>child:male': { status: 'resolved', primaryTitle: '表兄', aliases: ['表哥'] },
  'parent:female>sibling:male:younger>child:male': { status: 'resolved', primaryTitle: '表弟' },
  'parent:female>sibling:male:older>child:female': { status: 'resolved', primaryTitle: '表姐' },
  'parent:female>sibling:male:younger>child:female': { status: 'resolved', primaryTitle: '表妹' },
  'parent:female>sibling:male:unknown>child:male': {
    status: 'ambiguous',
    candidates: ['表兄', '表弟', '表兄弟'],
    explanation: '舅表兄弟需要补充长幼信息。'
  },
  'parent:female>sibling:female:older>child:male': { status: 'resolved', primaryTitle: '表兄', aliases: ['表哥'] },
  'parent:female>sibling:female:younger>child:male': { status: 'resolved', primaryTitle: '表弟' },
  'parent:female>sibling:female:older>child:female': { status: 'resolved', primaryTitle: '表姐' },
  'parent:female>sibling:female:younger>child:female': { status: 'resolved', primaryTitle: '表妹' },
  'parent:female>sibling:female:unknown>child:female': {
    status: 'ambiguous',
    candidates: ['表姐', '表妹', '表姐妹'],
    explanation: '姨表姐妹需要补充长幼信息。'
  },
  'parent:unknown>sibling:male:older>child:male': {
    status: 'ambiguous',
    candidates: ['堂兄', '堂弟', '表兄', '表弟'],
    explanation: '需要补充父母性别以区分堂亲或表亲。'
  }
}

function getCousinSide(pathKey: string): 'paternal' | 'maternalOrCross' | 'unknown' {
  const parts = pathKey.split('>')
  const parentGender = parts[0]?.split(':')[1]
  const siblingGender = parts[1]?.split(':')[1]
  if (parentGender === 'male' && siblingGender === 'male') return 'paternal'
  if (parentGender === 'unknown' || siblingGender === 'unknown') return 'unknown'
  return 'maternalOrCross'
}

function genderedTitle(gender: string | undefined, male: string, female: string, neutral: string): KinshipRuleMatch {
  if (gender === 'male') return { status: 'resolved', primaryTitle: male }
  if (gender === 'female') return { status: 'resolved', primaryTitle: female }
  return {
    status: 'ambiguous',
    primaryTitle: neutral,
    candidates: [neutral, male, female],
    explanation: '需要补充性别以区分更具体的称谓。'
  }
}

function matchCousinDescendantRules(context: KinshipContext, pathKey: string): KinshipRuleMatch | null {
  const relationKey = buildRelationKey(context)
  if (relationKey !== 'parent>sibling>child>child' && relationKey !== 'parent>sibling>child>child>child') {
    return null
  }

  const side = getCousinSide(pathKey)
  const lastGender = pathKey.split('>').pop()?.split(':')[1]
  if (relationKey === 'parent>sibling>child>child') {
    if (side === 'paternal') {
      return genderedTitle(lastGender, '堂侄', '堂侄女', '堂侄子女')
    }
    if (side === 'maternalOrCross') {
      return genderedTitle(lastGender, '表侄', '表侄女', '表侄子女')
    }
    return {
      status: 'ambiguous',
      primaryTitle: '堂表亲侄辈',
      candidates: ['堂侄', '堂侄女', '表侄', '表侄女', '堂表亲侄辈'],
      explanation: '需要补充父母性别、旁系性别和孩子性别以区分堂亲或表亲侄辈。'
    }
  }

  if (side === 'paternal') {
    return genderedTitle(lastGender, '堂侄孙', '堂侄孙女', '堂侄孙辈')
  }
  if (side === 'maternalOrCross') {
    return genderedTitle(lastGender, '表侄孙', '表侄孙女', '表侄孙辈')
  }
  return {
    status: 'ambiguous',
    primaryTitle: '堂表亲侄孙辈',
    candidates: ['堂侄孙', '堂侄孙女', '表侄孙', '表侄孙女', '堂表亲侄孙辈'],
    explanation: '需要补充父母性别、旁系性别和后代性别以区分堂亲或表亲侄孙辈。'
  }
}

export function matchCousinRules(context: KinshipContext, pathKey: string): KinshipRuleMatch | null {
  return COUSIN_RULES[pathKey] || matchCousinDescendantRules(context, pathKey)
}

export { COUSIN_RULES }
