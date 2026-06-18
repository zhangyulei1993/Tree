import type { KinshipContext, KinshipRuleMatch } from './types'
import { genderLabel } from './ruleUtils'

const NEPHEW_RULES: Record<string, KinshipRuleMatch> = {
  'sibling:male:older>child:male>child:male': { status: 'resolved', primaryTitle: '侄孙' },
  'sibling:male:older>child:male>child:female': { status: 'resolved', primaryTitle: '侄孙女' },
  'sibling:male:younger>child:male>child:male': { status: 'resolved', primaryTitle: '侄孙' },
  'sibling:male:younger>child:female>child:male': {
    status: 'ambiguous',
    primaryTitle: '侄甥后代',
    candidates: ['侄外孙', '侄孙', '侄甥后代'],
    explanation: '侄甥后代称谓存在地区差异，当前版本给出保守结果。'
  },
  'sibling:female:older>child:male>child:male': { status: 'resolved', primaryTitle: '外甥孙' },
  'sibling:female:older>child:male>child:female': { status: 'resolved', primaryTitle: '外甥孙女' },
  'sibling:female:younger>child:female>child:female': {
    status: 'ambiguous',
    primaryTitle: '外甥后代',
    candidates: ['外甥外孙女', '外甥孙女', '外甥后代'],
    explanation: '外甥后代称谓存在地区差异，当前版本给出保守结果。'
  },
  'sibling:female:younger>child:male>child:male': { status: 'resolved', primaryTitle: '外甥孙' },
  'sibling:male:older>child:male>child:male>child:male': {
    status: 'resolved',
    primaryTitle: '侄重孙'
  },
  'sibling:male:older>child:male>child:male>child:female': {
    status: 'resolved',
    primaryTitle: '侄重孙女'
  },
  'sibling:male:younger>child:male>child:male>child:male': {
    status: 'resolved',
    primaryTitle: '侄重孙'
  },
  'sibling:male:younger>child:male>child:male>child:female': {
    status: 'resolved',
    primaryTitle: '侄重孙女'
  },
  'sibling:female:older>child:male>child:male>child:male': {
    status: 'resolved',
    primaryTitle: '外甥重孙'
  },
  'sibling:female:older>child:male>child:male>child:female': {
    status: 'resolved',
    primaryTitle: '外甥重孙女'
  },
  'sibling:female:younger>child:male>child:male>child:male': {
    status: 'resolved',
    primaryTitle: '外甥重孙'
  },
  'sibling:female:younger>child:male>child:male>child:female': {
    status: 'resolved',
    primaryTitle: '外甥重孙女'
  },
  'sibling:male:unknown>child:unknown>child:unknown': {
    status: 'ambiguous',
    candidates: ['侄孙', '侄孙女', '外甥孙', '外甥孙女', '侄甥后代'],
    explanation: '侄甥后代需要补充性别和长幼信息。'
  }
}

export function matchNephewRules(_context: KinshipContext, pathKey: string): KinshipRuleMatch | null {
  if (pathKey in NEPHEW_RULES) {
    return NEPHEW_RULES[pathKey]
  }

  const relKey = pathKey
    .split('>')
    .map((segment) => segment.split(':')[0])
    .join('>')

  if (relKey === 'sibling>child>child' && !pathKey.includes('sibling:female')) {
    const lastGender = pathKey.split('>').pop()?.split(':')[1]
    if (lastGender === 'male' || lastGender === 'female') {
      return {
        status: 'ambiguous',
        primaryTitle: '侄甥后代',
        candidates: ['侄孙', '侄孙女', '侄甥后代'],
        explanation: '侄甥后代称谓存在地区差异，请补充更多信息。'
      }
    }
  }

  if (relKey === 'sibling>child>child' && pathKey.includes('sibling:female')) {
    return {
      status: 'ambiguous',
      primaryTitle: '外甥后代',
      candidates: ['外甥孙', '外甥孙女', '外甥后代'],
      explanation: '外甥后代称谓存在地区差异，请补充更多信息。'
    }
  }

  return null
}

export function matchSiblingChildRule(pathKey: string): KinshipRuleMatch | null {
  if (!pathKey.match(/^sibling:(male|female):(older|younger|unknown)>child:(male|female)$/)) {
    return null
  }
  const [, siblingGender, , childGender] =
    pathKey.match(/^sibling:(male|female):(older|younger|unknown)>child:(male|female)$/) || []
  if (siblingGender === 'male') {
    return {
      status: 'resolved',
      primaryTitle: genderLabel(childGender as 'male' | 'female', '侄子', '侄女', '侄子女')
    }
  }
  if (siblingGender === 'female') {
    return {
      status: 'resolved',
      primaryTitle: genderLabel(childGender as 'male' | 'female', '外甥', '外甥女', '外甥子女')
    }
  }
  return null
}

export { NEPHEW_RULES }
