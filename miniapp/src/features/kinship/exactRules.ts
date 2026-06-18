import type { KinshipContext, KinshipRuleMatch } from './types'

const EXACT_RULES: Record<string, KinshipRuleMatch> = {
  'parent:male': { status: 'resolved', primaryTitle: '父亲', aliases: ['爸爸', '爹'] },
  'parent:female': { status: 'resolved', primaryTitle: '母亲', aliases: ['妈妈', '娘'] },
  'parent:unknown': { status: 'resolved', primaryTitle: '父母' },
  'child:male': { status: 'resolved', primaryTitle: '儿子' },
  'child:female': { status: 'resolved', primaryTitle: '女儿' },
  'child:unknown': { status: 'resolved', primaryTitle: '子女' },
  'spouse:male': { status: 'resolved', primaryTitle: '丈夫', aliases: ['老公'] },
  'spouse:female': { status: 'resolved', primaryTitle: '妻子', aliases: ['老婆'] },
  'spouse:unknown': { status: 'resolved', primaryTitle: '配偶' },
  'sibling:male:older': { status: 'resolved', primaryTitle: '哥哥', aliases: ['兄长'] },
  'sibling:male:younger': { status: 'resolved', primaryTitle: '弟弟' },
  'sibling:female:older': { status: 'resolved', primaryTitle: '姐姐', aliases: ['姊姊'] },
  'sibling:female:younger': { status: 'resolved', primaryTitle: '妹妹' },
  'sibling:male:unknown': {
    status: 'ambiguous',
    candidates: ['哥哥', '弟弟', '兄弟姐妹'],
    explanation: '兄弟姐妹关系需要补充长幼信息。'
  },
  'sibling:male:same': {
    status: 'ambiguous',
    primaryTitle: '兄弟',
    candidates: ['兄弟', '哥哥', '弟弟', '双胞胎兄弟'],
    explanation: '同龄或生日相同的兄弟称谓存在家庭习惯差异。'
  },
  'sibling:female:unknown': {
    status: 'ambiguous',
    candidates: ['姐姐', '妹妹', '兄弟姐妹'],
    explanation: '兄弟姐妹关系需要补充长幼信息。'
  },
  'sibling:female:same': {
    status: 'ambiguous',
    primaryTitle: '姐妹',
    candidates: ['姐妹', '姐姐', '妹妹', '双胞胎姐妹'],
    explanation: '同龄或生日相同的姐妹称谓存在家庭习惯差异。'
  },
  'sibling:unknown:older': {
    status: 'ambiguous',
    primaryTitle: '年长的兄弟姐妹',
    candidates: ['年长的兄弟姐妹', '哥哥', '姐姐'],
    explanation: '需要补充兄弟姐妹性别以区分哥哥或姐姐。'
  },
  'sibling:unknown:younger': {
    status: 'ambiguous',
    primaryTitle: '年幼的兄弟姐妹',
    candidates: ['弟弟', '妹妹', '年幼的兄弟姐妹'],
    explanation: '需要补充兄弟姐妹性别以区分弟弟或妹妹。'
  },
  'sibling:unknown:same': {
    status: 'ambiguous',
    primaryTitle: '同龄兄弟姐妹',
    candidates: ['双胞胎兄弟', '双胞胎姐妹', '同龄兄弟姐妹'],
    explanation: '需要补充兄弟姐妹性别和家庭称谓习惯。'
  },
  'sibling:unknown:unknown': {
    status: 'ambiguous',
    candidates: ['哥哥', '弟弟', '姐姐', '妹妹', '兄弟姐妹'],
    explanation: '兄弟姐妹关系需要补充性别或长幼信息。'
  },

  'parent:male>parent:female': { status: 'resolved', primaryTitle: '祖母', aliases: ['奶奶'] },
  'parent:male>parent:male': { status: 'resolved', primaryTitle: '祖父', aliases: ['爷爷'] },
  'parent:female>parent:male': { status: 'resolved', primaryTitle: '外祖父', aliases: ['外公'] },
  'parent:female>parent:female': { status: 'resolved', primaryTitle: '外祖母', aliases: ['外婆'] },
  'parent:unknown>parent:male': {
    status: 'ambiguous',
    candidates: ['祖父', '外祖父'],
    aliases: ['爷爷', '外公'],
    explanation: '需要补充第一层父母性别以区分父系或母系祖辈。'
  },
  'parent:unknown>parent:female': {
    status: 'ambiguous',
    candidates: ['祖母', '外祖母'],
    aliases: ['奶奶', '外婆'],
    explanation: '需要补充第一层父母性别以区分父系或母系祖辈。'
  },

  'parent:male>sibling:male:older': { status: 'resolved', primaryTitle: '伯父', aliases: ['大伯', '伯伯'] },
  'parent:male>sibling:male:younger': { status: 'resolved', primaryTitle: '叔父', aliases: ['叔叔'] },
  'parent:male>sibling:female:older': { status: 'resolved', primaryTitle: '姑母', aliases: ['姑姑', '姑妈'] },
  'parent:male>sibling:female:younger': { status: 'resolved', primaryTitle: '姑母', aliases: ['姑姑', '姑妈'] },
  'parent:male>sibling:male:unknown': {
    status: 'ambiguous',
    candidates: ['伯父', '叔父'],
    explanation: '父系兄弟姐妹需要补充长幼信息。'
  },
  'parent:female>sibling:male:older': { status: 'resolved', primaryTitle: '舅父', aliases: ['舅舅'] },
  'parent:female>sibling:male:younger': { status: 'resolved', primaryTitle: '舅父', aliases: ['舅舅'] },
  'parent:female>sibling:female:older': { status: 'resolved', primaryTitle: '姨母', aliases: ['姨妈', '阿姨'] },
  'parent:female>sibling:female:younger': { status: 'resolved', primaryTitle: '姨母', aliases: ['姨妈', '阿姨'] },
  'parent:unknown>sibling:male:unknown': {
    status: 'ambiguous',
    candidates: ['伯父', '叔父', '舅父'],
    explanation: '需要补充第一层父母性别以区分父系或母系旁系。'
  },
  'parent:unknown>sibling:female:unknown': {
    status: 'ambiguous',
    candidates: ['姑母', '姨母'],
    explanation: '需要补充第一层父母性别以区分父系或母系旁系。'
  },

  'sibling:male:older>child:male': { status: 'resolved', primaryTitle: '侄子' },
  'sibling:male:older>child:female': { status: 'resolved', primaryTitle: '侄女' },
  'sibling:male:younger>child:male': { status: 'resolved', primaryTitle: '侄子' },
  'sibling:male:younger>child:female': { status: 'resolved', primaryTitle: '侄女' },
  'sibling:female:older>child:male': { status: 'resolved', primaryTitle: '外甥' },
  'sibling:female:older>child:female': { status: 'resolved', primaryTitle: '外甥女' },
  'sibling:female:younger>child:male': { status: 'resolved', primaryTitle: '外甥' },
  'sibling:female:younger>child:female': { status: 'resolved', primaryTitle: '外甥女' },
  'sibling:male:unknown>child:male': {
    status: 'resolved',
    primaryTitle: '侄子'
  },
  'sibling:male:unknown>child:female': {
    status: 'resolved',
    primaryTitle: '侄女'
  },
  'sibling:female:unknown>child:male': {
    status: 'resolved',
    primaryTitle: '外甥'
  },
  'sibling:female:unknown>child:female': {
    status: 'resolved',
    primaryTitle: '外甥女'
  },
  'sibling:male:unknown>child:unknown': {
    status: 'ambiguous',
    primaryTitle: '侄子女',
    candidates: ['侄子', '侄女', '侄子女'],
    explanation: '需要补充孩子性别以区分侄子或侄女。'
  },
  'sibling:female:unknown>child:unknown': {
    status: 'ambiguous',
    primaryTitle: '外甥子女',
    candidates: ['外甥', '外甥女', '外甥子女'],
    explanation: '需要补充孩子性别以区分外甥或外甥女。'
  },
  'sibling:unknown:unknown>child:unknown': {
    status: 'ambiguous',
    primaryTitle: '侄甥子女',
    candidates: ['侄子', '侄女', '外甥', '外甥女', '侄甥子女'],
    explanation: '需要补充兄弟姐妹性别和孩子性别以区分侄甥称谓。'
  },

  'spouse:female>sibling:male:older': { status: 'resolved', primaryTitle: '大舅子', aliases: ['内兄'] },
  'spouse:female>sibling:male:younger': { status: 'resolved', primaryTitle: '小舅子', aliases: ['内弟'] },
  'spouse:female>sibling:female:older': { status: 'resolved', primaryTitle: '大姨子', aliases: ['姨姐'] },
  'spouse:female>sibling:female:younger': { status: 'resolved', primaryTitle: '小姨子', aliases: ['姨妹'] },
  'spouse:male>sibling:male:older': { status: 'resolved', primaryTitle: '大伯子', aliases: ['大伯哥'] },
  'spouse:male>sibling:male:younger': { status: 'resolved', primaryTitle: '小叔子' },
  'spouse:male>sibling:female:older': { status: 'resolved', primaryTitle: '大姑子' },
  'spouse:male>sibling:female:younger': { status: 'resolved', primaryTitle: '小姑子' },

  'spouse:male>child:female': { status: 'resolved', primaryTitle: '继女', aliases: ['继子女'] },
  'spouse:female>child:male': { status: 'resolved', primaryTitle: '继子', aliases: ['继子女'] },

  'child:male>child:male': { status: 'resolved', primaryTitle: '孙子', aliases: ['孙儿'] },
  'child:male>child:female': { status: 'resolved', primaryTitle: '孙女' },
  'child:male>child:unknown': {
    status: 'ambiguous',
    primaryTitle: '孙辈',
    candidates: ['孙辈', '孙子', '孙女'],
    explanation: '需要补充孙辈性别以区分孙子或孙女。'
  },
  'child:female>child:male': { status: 'resolved', primaryTitle: '外孙' },
  'child:female>child:female': { status: 'resolved', primaryTitle: '外孙女' },
  'child:female>child:unknown': {
    status: 'ambiguous',
    primaryTitle: '外孙辈',
    candidates: ['外孙辈', '外孙', '外孙女'],
    explanation: '需要补充外孙辈性别以区分外孙或外孙女。'
  },
  'child:unknown>child:male': {
    status: 'ambiguous',
    candidates: ['孙子', '外孙'],
    explanation: '需要补充第一层子女性别以区分儿系或女系孙辈。'
  },
  'child:unknown>child:female': {
    status: 'ambiguous',
    candidates: ['孙女', '外孙女'],
    explanation: '需要补充第一层子女性别以区分儿系或女系孙辈。'
  },
  'child:unknown>child:unknown': {
    status: 'ambiguous',
    primaryTitle: '孙辈',
    candidates: ['孙辈', '孙子', '孙女', '外孙', '外孙女'],
    explanation: '需要补充第一层子女性别和孙辈性别以区分称谓。'
  }
}

import type { KinshipContext } from './types'

function matchSpouseParent(context: KinshipContext, pathKey: string): KinshipRuleMatch | null {
  if (!pathKey.match(/^spouse:(male|female)>parent:(male|female)$/)) return null
  const selfGender = context.self.gender || 'unknown'
  const parentGender = context.steps[1].person.gender || 'unknown'

  if (selfGender === 'male' && parentGender === 'male') {
    return { status: 'resolved', primaryTitle: '岳父', aliases: ['丈人'] }
  }
  if (selfGender === 'male' && parentGender === 'female') {
    return { status: 'resolved', primaryTitle: '岳母', aliases: ['丈母娘'] }
  }
  if (selfGender === 'female' && parentGender === 'male') {
    return { status: 'resolved', primaryTitle: '公公' }
  }
  if (selfGender === 'female' && parentGender === 'female') {
    return { status: 'resolved', primaryTitle: '婆婆' }
  }
  return {
    status: 'ambiguous',
    candidates: ['公公', '婆婆', '岳父', '岳母', '配偶的父母'],
    explanation: '需要补充本人性别和对方父母性别。'
  }
}

function matchSpouseSiblingUnknown(context: KinshipContext, pathKey: string): KinshipRuleMatch | null {
  if (!pathKey.startsWith('spouse:unknown>sibling:')) return null
  const parts = pathKey.split('>')
  if (parts.length !== 2) return null
  const siblingPart = parts[1]
  if (siblingPart === 'sibling:male:older') {
    return {
      status: 'ambiguous',
      candidates: ['配偶的哥哥', '大舅子', '大伯子'],
      explanation: '需要补充配偶性别以区分称谓。'
    }
  }
  if (siblingPart === 'sibling:male:younger') {
    return {
      status: 'ambiguous',
      candidates: ['配偶的弟弟', '小舅子', '小叔子'],
      explanation: '需要补充配偶性别以区分称谓。'
    }
  }
  if (siblingPart === 'sibling:female:older') {
    return {
      status: 'ambiguous',
      candidates: ['配偶的姐姐', '大姨子', '大姑子'],
      explanation: '需要补充配偶性别以区分称谓。'
    }
  }
  if (siblingPart === 'sibling:female:younger') {
    return {
      status: 'ambiguous',
      candidates: ['配偶的妹妹', '小姨子', '小姑子'],
      explanation: '需要补充配偶性别以区分称谓。'
    }
  }
  return {
    status: 'ambiguous',
    candidates: ['配偶的哥哥', '配偶的弟弟', '配偶的姐姐', '配偶的妹妹'],
    explanation: '配偶兄弟姐妹的称谓因配偶性别、长幼而异，请补充信息。'
  }
}

export function matchExactRules(context: KinshipContext, pathKey: string): KinshipRuleMatch | null {
  if (pathKey in EXACT_RULES) {
    return EXACT_RULES[pathKey]
  }
  const siblingChildKnown = pathKey.match(/^sibling:(male|female|unknown):(older|younger|same|unknown)>child:(male|female)$/)
  if (siblingChildKnown) {
    const siblingGender = siblingChildKnown[1]
    const childGender = siblingChildKnown[3]
    if (siblingGender === 'male') {
      return {
        status: 'resolved',
        primaryTitle: childGender === 'male' ? '侄子' : '侄女'
      }
    }
    if (siblingGender === 'female') {
      return {
        status: 'resolved',
        primaryTitle: childGender === 'male' ? '外甥' : '外甥女'
      }
    }
    return {
      status: 'ambiguous',
      primaryTitle: '侄甥子女',
      candidates: childGender === 'male'
        ? ['侄甥子女', '侄子', '外甥']
        : ['侄甥子女', '侄女', '外甥女'],
      explanation: '需要补充兄弟姐妹性别以区分侄甥称谓。'
    }
  }
  const siblingChildUnknown = pathKey.match(/^sibling:(male|female|unknown):(older|younger|same|unknown)>child:unknown$/)
  if (siblingChildUnknown) {
    const siblingGender = siblingChildUnknown[1]
    if (siblingGender === 'male') {
      return {
        status: 'ambiguous',
        primaryTitle: '侄子女',
        candidates: ['侄子女', '侄子', '侄女'],
        explanation: '需要补充孩子性别以区分侄子或侄女。'
      }
    }
    if (siblingGender === 'female') {
      return {
        status: 'ambiguous',
        primaryTitle: '外甥子女',
        candidates: ['外甥子女', '外甥', '外甥女'],
        explanation: '需要补充孩子性别以区分外甥或外甥女。'
      }
    }
    return {
      status: 'ambiguous',
      primaryTitle: '侄甥子女',
      candidates: ['侄甥子女', '侄子', '侄女', '外甥', '外甥女'],
      explanation: '需要补充兄弟姐妹性别和孩子性别以区分侄甥称谓。'
    }
  }
  return matchSpouseParent(context, pathKey) || matchSpouseSiblingUnknown(context, pathKey)
}

export function getExactRule(pathKey: string): KinshipRuleMatch | undefined {
  return EXACT_RULES[pathKey]
}

export { EXACT_RULES }
