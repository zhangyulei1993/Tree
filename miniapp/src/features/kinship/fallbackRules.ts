import { buildRelationKey } from './pathKey'
import type { KinshipContext, KinshipRuleMatch } from './types'

const RELATION_FALLBACKS: Record<string, KinshipRuleMatch> = {
  'parent>sibling>child': {
    status: 'ambiguous',
    primaryTitle: '父母旁系晚辈',
    candidates: ['堂亲', '表亲', '父母旁系晚辈'],
    explanation: '该路径属于父母兄弟姐妹的子女，常见为堂亲或表亲，请补充性别和长幼信息。'
  },
  'spouse>sibling>child': {
    status: 'ambiguous',
    primaryTitle: '姻亲晚辈',
    candidates: ['姻亲晚辈', '配偶侄辈', '配偶外甥辈'],
    explanation: '该路径属于配偶兄弟姐妹的子女，称谓存在地区差异。'
  },
  'parent>parent>sibling': {
    status: 'ambiguous',
    primaryTitle: '祖辈旁系亲属',
    candidates: ['伯祖父', '叔祖父', '姑祖母', '外祖辈旁系'],
    explanation: '该路径属于祖辈的兄弟姐妹，称谓存在地区差异。'
  },
  'parent>parent>sibling>child': {
    status: 'ambiguous',
    primaryTitle: '祖辈旁系晚辈',
    candidates: ['祖辈旁系晚辈', '从祖亲属', '远房旁系亲属'],
    explanation: '该路径属于祖辈兄弟姐妹的子女，称谓存在地区差异。'
  },
  'parent>parent>sibling>child>spouse': {
    status: 'ambiguous',
    primaryTitle: '祖辈旁系晚辈的配偶',
    candidates: ['祖辈旁系晚辈的配偶', '远房姻亲', '从祖亲属配偶'],
    explanation: '该路径属于祖辈旁系晚辈的配偶，常见称谓地区差异较大，当前版本给出保守泛称。'
  },
  'parent>parent>sibling>child>child': {
    status: 'ambiguous',
    primaryTitle: '祖辈旁系后代',
    candidates: ['祖辈旁系后代', '远房旁系亲属'],
    explanation: '该路径属于祖辈旁系的后代，称谓存在地区差异。'
  },
  'parent>parent>parent>sibling': {
    status: 'ambiguous',
    primaryTitle: '曾祖辈旁系亲属',
    candidates: ['曾伯祖父', '曾叔祖父', '曾姑祖母', '曾舅祖父', '曾姨祖母', '曾祖辈旁系亲属'],
    explanation: '该路径属于曾祖辈的兄弟姐妹，称谓存在地区差异。'
  },
  'parent>parent>parent>sibling>child': {
    status: 'ambiguous',
    primaryTitle: '曾祖辈旁系晚辈',
    candidates: ['曾祖辈旁系晚辈', '远房旁系亲属'],
    explanation: '该路径属于曾祖辈旁系的子女，称谓存在地区差异。'
  },
  'parent>parent>parent>parent>sibling': {
    status: 'ambiguous',
    primaryTitle: '高祖辈旁系亲属',
    candidates: ['高伯祖父', '高叔祖父', '高姑祖母', '高舅祖父', '高姨祖母', '高祖辈旁系亲属'],
    explanation: '该路径属于高祖辈的兄弟姐妹，称谓存在地区差异。'
  },
  'sibling>child>child': {
    status: 'ambiguous',
    primaryTitle: '侄甥后代',
    candidates: ['侄孙', '侄孙女', '外甥孙', '外甥孙女', '侄甥后代'],
    explanation: '该路径属于兄弟姐妹子女的后代，称谓存在地区差异。'
  },
  'spouse>sibling>child>child': {
    status: 'ambiguous',
    primaryTitle: '姻亲晚辈',
    candidates: ['姻亲晚辈', '配偶侄辈', '配偶外甥辈'],
    explanation: '该路径属于配偶旁系亲属的下一代，称谓存在地区差异。'
  },
  'sibling>child>child>child': {
    status: 'ambiguous',
    primaryTitle: '侄重孙辈',
    candidates: ['侄重孙', '侄重孙女', '外甥重孙', '外甥重孙女', '侄重孙辈'],
    explanation: '该路径属于侄甥后代的更晚辈，称谓存在地区差异。'
  },
  'sibling>child>child>child>child': {
    status: 'ambiguous',
    primaryTitle: '侄甥玄孙辈',
    candidates: ['侄甥玄孙辈', '侄甥后代'],
    explanation: '该路径属于侄甥后代的更晚辈，当前版本给出保守泛称。'
  },
  'spouse>parent>sibling': {
    status: 'ambiguous',
    primaryTitle: '姻亲长辈旁系',
    candidates: ['姻亲长辈旁系', '配偶伯父', '配偶叔父', '配偶姑母', '配偶舅父', '配偶姨母'],
    explanation: '该路径属于配偶父母的兄弟姐妹，称谓存在地区差异。'
  },
  'child>spouse>parent': {
    status: 'ambiguous',
    primaryTitle: '亲家',
    candidates: ['亲家', '亲家公', '亲家母', '姻亲长辈'],
    explanation: '该路径属于子女配偶的父母，常见称为亲家。'
  },
  'parent>sibling>child>child': {
    status: 'ambiguous',
    primaryTitle: '堂表亲后代',
    candidates: ['堂表亲后代', '远房堂表亲', '父母旁系晚辈的后代'],
    explanation: '该路径属于堂表亲的下一代，称谓存在地区差异。'
  },
  'spouse>parent>parent': {
    status: 'ambiguous',
    primaryTitle: '配偶祖辈',
    candidates: ['配偶祖父', '配偶外祖父', '配偶祖辈', '姻亲长辈'],
    explanation: '该路径属于配偶的祖辈，称谓存在地区差异。'
  },
  'child>child>child': {
    status: 'ambiguous',
    primaryTitle: '曾孙辈',
    candidates: ['曾孙', '曾孙女', '外曾孙', '外曾孙女', '曾孙辈'],
    explanation: '该路径属于孙辈的子女，请补充子女性别以区分儿系或女系。'
  },
  'sibling>spouse>parent': {
    status: 'ambiguous',
    primaryTitle: '姻亲长辈',
    candidates: ['姻亲长辈', '兄弟姐妹配偶的父母'],
    explanation: '该路径属于兄弟姐妹配偶的父母，称谓存在地区差异。'
  }
}

const DEFAULT_FALLBACK: KinshipRuleMatch = {
  status: 'unsupported',
  primaryTitle: '亲属关系路径已记录',
  explanation: '亲属关系路径已记录，暂未收录常见称谓。你仍然可以保留完整关系路径。'
}

export function matchFallbackRules(context: KinshipContext): KinshipRuleMatch {
  const relKey = buildRelationKey(context)
  if (relKey in RELATION_FALLBACKS) {
    return RELATION_FALLBACKS[relKey]
  }

  if (relKey.startsWith('parent>sibling')) {
    return {
      status: 'ambiguous',
      primaryTitle: '父母旁系亲属',
      candidates: ['父母兄弟姐妹', '父母旁系亲属'],
      explanation: '该路径属于父母的旁系亲属，请补充更多信息。'
    }
  }

  if (relKey.startsWith('spouse>')) {
    return {
      status: 'ambiguous',
      primaryTitle: '姻亲亲属',
      candidates: ['姻亲亲属', '姻亲长辈', '姻亲晚辈', '姻亲旁系'],
      explanation: '该路径属于配偶一方的亲属，称谓存在地区差异。'
    }
  }

  return DEFAULT_FALLBACK
}

export { RELATION_FALLBACKS, DEFAULT_FALLBACK }
