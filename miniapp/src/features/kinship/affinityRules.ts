import { buildRelationKey } from './pathKey'
import type { KinshipContext, KinshipRuleMatch } from './types'

const AFFINITY_RULES: Record<string, KinshipRuleMatch> = {
  'child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '儿媳',
    aliases: ['儿媳妇'],
    terminal: true
  },
  'child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '女婿',
    terminal: true
  },
  'child:male>spouse:male': {
    status: 'ambiguous',
    primaryTitle: '子女的配偶',
    candidates: ['子女的配偶', '儿婿', '儿媳'],
    terminal: true,
    explanation: '子女的配偶称谓因具体关系而异。'
  },
  'child:female>spouse:female': {
    status: 'ambiguous',
    primaryTitle: '子女的配偶',
    candidates: ['子女的配偶', '女婿', '儿媳'],
    terminal: true,
    explanation: '子女的配偶称谓因具体关系而异。'
  },
  'child:male>spouse:unknown': {
    status: 'ambiguous',
    primaryTitle: '子女的配偶',
    candidates: ['子女的配偶', '儿媳', '女婿'],
    terminal: true,
    explanation: '需要补充配偶性别以区分儿媳或女婿。'
  },
  'child:female>spouse:unknown': {
    status: 'ambiguous',
    primaryTitle: '子女的配偶',
    candidates: ['子女的配偶', '儿媳', '女婿'],
    terminal: true,
    explanation: '需要补充配偶性别以区分儿媳或女婿。'
  },

  'sibling:male:older>spouse:female': {
    status: 'resolved',
    primaryTitle: '嫂子',
    aliases: ['嫂嫂'],
    terminal: true
  },
  'sibling:male:younger>spouse:female': {
    status: 'resolved',
    primaryTitle: '弟媳',
    aliases: ['弟妹'],
    terminal: true
  },
  'sibling:female:older>spouse:male': {
    status: 'resolved',
    primaryTitle: '姐夫',
    terminal: true
  },
  'sibling:female:younger>spouse:male': {
    status: 'resolved',
    primaryTitle: '妹夫',
    terminal: true
  },
  'sibling:male:older>spouse:unknown': {
    status: 'ambiguous',
    primaryTitle: '兄弟的配偶',
    candidates: ['兄弟的配偶', '嫂子', '弟媳'],
    terminal: true,
    explanation: '需要补充配偶性别以区分称谓。'
  },
  'sibling:male:younger>spouse:unknown': {
    status: 'ambiguous',
    primaryTitle: '兄弟的配偶',
    candidates: ['兄弟的配偶', '嫂子', '弟媳'],
    terminal: true,
    explanation: '需要补充配偶性别以区分称谓。'
  },
  'sibling:female:older>spouse:unknown': {
    status: 'ambiguous',
    primaryTitle: '姐妹的配偶',
    candidates: ['姐妹的配偶', '姐夫', '妹夫'],
    terminal: true,
    explanation: '需要补充配偶性别以区分称谓。'
  },
  'sibling:female:younger>spouse:unknown': {
    status: 'ambiguous',
    primaryTitle: '姐妹的配偶',
    candidates: ['姐妹的配偶', '姐夫', '妹夫'],
    terminal: true,
    explanation: '需要补充配偶性别以区分称谓。'
  },
  'sibling:male:unknown>spouse:female': {
    status: 'ambiguous',
    primaryTitle: '兄弟的配偶',
    candidates: ['嫂子', '弟媳', '兄弟的配偶'],
    terminal: true,
    explanation: '兄弟的配偶需要补充长幼信息。'
  },
  'sibling:female:unknown>spouse:male': {
    status: 'ambiguous',
    primaryTitle: '姐妹的配偶',
    candidates: ['姐夫', '妹夫', '姐妹的配偶'],
    terminal: true,
    explanation: '姐妹的配偶需要补充长幼信息。'
  },
  'sibling:male:unknown>spouse:unknown': {
    status: 'ambiguous',
    primaryTitle: '兄弟的配偶',
    candidates: ['兄弟的配偶', '嫂子', '弟媳'],
    terminal: true,
    explanation: '兄弟的配偶需要补充性别和长幼信息。'
  },
  'sibling:female:unknown>spouse:unknown': {
    status: 'ambiguous',
    primaryTitle: '姐妹的配偶',
    candidates: ['姐妹的配偶', '姐夫', '妹夫'],
    terminal: true,
    explanation: '姐妹的配偶需要补充性别和长幼信息。'
  },

  'parent:male>sibling:male:older>spouse:female': {
    status: 'resolved',
    primaryTitle: '伯母',
    aliases: ['伯娘'],
    terminal: true
  },
  'parent:male>sibling:male:younger>spouse:female': {
    status: 'resolved',
    primaryTitle: '婶婶',
    aliases: ['婶母'],
    terminal: true
  },
  'parent:male>sibling:female:older>spouse:male': {
    status: 'resolved',
    primaryTitle: '姑父',
    terminal: true
  },
  'parent:male>sibling:female:younger>spouse:male': {
    status: 'resolved',
    primaryTitle: '姑父',
    terminal: true
  },
  'parent:female>sibling:male:older>spouse:female': {
    status: 'resolved',
    primaryTitle: '舅妈',
    aliases: ['舅母'],
    terminal: true
  },
  'parent:female>sibling:male:younger>spouse:female': {
    status: 'resolved',
    primaryTitle: '舅妈',
    aliases: ['舅母'],
    terminal: true
  },
  'parent:female>sibling:female:older>spouse:male': {
    status: 'resolved',
    primaryTitle: '姨父',
    terminal: true
  },
  'parent:female>sibling:female:younger>spouse:male': {
    status: 'resolved',
    primaryTitle: '姨父',
    terminal: true
  },
  'parent:male>sibling:male:unknown>spouse:female': {
    status: 'ambiguous',
    primaryTitle: '父亲兄弟的配偶',
    candidates: ['伯母', '婶婶', '父亲兄弟的配偶'],
    terminal: true,
    explanation: '需要补充父亲兄弟的长幼信息以区分伯母或婶婶。'
  },
  'parent:male>sibling:female:unknown>spouse:male': {
    status: 'resolved',
    primaryTitle: '姑父',
    terminal: true
  },
  'parent:female>sibling:male:unknown>spouse:female': {
    status: 'resolved',
    primaryTitle: '舅妈',
    aliases: ['舅母'],
    terminal: true
  },
  'parent:female>sibling:female:unknown>spouse:male': {
    status: 'resolved',
    primaryTitle: '姨父',
    terminal: true
  },
  'parent:unknown>sibling:male:unknown>spouse:female': {
    status: 'ambiguous',
    primaryTitle: '父母兄弟的配偶',
    candidates: ['伯母', '婶婶', '舅妈', '父母兄弟的配偶'],
    terminal: true,
    explanation: '需要补充父母性别和兄弟长幼信息以区分称谓。'
  },
  'parent:unknown>sibling:female:unknown>spouse:male': {
    status: 'ambiguous',
    primaryTitle: '父母姐妹的配偶',
    candidates: ['姑父', '姨父', '父母姐妹的配偶'],
    terminal: true,
    explanation: '需要补充父母性别以区分姑父或姨父。'
  },

  'spouse:female>sibling:male:older>spouse:female': {
    status: 'resolved',
    primaryTitle: '内兄嫂',
    terminal: true
  },
  'spouse:female>sibling:male:younger>spouse:female': {
    status: 'resolved',
    primaryTitle: '内弟媳',
    terminal: true
  },
  'spouse:female>sibling:female:older>spouse:male': {
    status: 'resolved',
    primaryTitle: '连襟',
    terminal: true
  },
  'spouse:female>sibling:female:younger>spouse:male': {
    status: 'resolved',
    primaryTitle: '连襟',
    terminal: true
  },
  'spouse:male>sibling:male:older>spouse:female': {
    status: 'resolved',
    primaryTitle: '妯娌',
    terminal: true
  },
  'spouse:male>sibling:male:younger>spouse:female': {
    status: 'resolved',
    primaryTitle: '妯娌',
    terminal: true
  },
  'spouse:male>sibling:female:older>spouse:male': {
    status: 'resolved',
    primaryTitle: '姑姐夫',
    terminal: true
  },
  'spouse:male>sibling:female:younger>spouse:male': {
    status: 'resolved',
    primaryTitle: '姑妹夫',
    terminal: true
  },

  'sibling:male:older>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '侄媳妇',
    terminal: true
  },
  'sibling:male:younger>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '侄媳妇',
    terminal: true
  },
  'sibling:male:older>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '侄女婿',
    terminal: true
  },
  'sibling:male:younger>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '侄女婿',
    terminal: true
  },
  'sibling:female:older>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '外甥媳妇',
    terminal: true
  },
  'sibling:female:younger>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '外甥媳妇',
    terminal: true
  },
  'sibling:female:older>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '外甥女婿',
    terminal: true
  },
  'sibling:female:younger>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '外甥女婿',
    terminal: true
  },
  'sibling:male:older>child:male>child:male>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '侄重孙媳',
    terminal: true
  },
  'sibling:male:younger>child:male>child:male>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '侄重孙女婿',
    terminal: true
  },
  'sibling:female:older>child:male>child:male>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '外甥重孙媳',
    terminal: true
  },
  'sibling:female:younger>child:male>child:male>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '外甥重孙女婿',
    terminal: true
  },

  'parent:male>sibling:male:older>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '堂嫂',
    terminal: true
  },
  'parent:male>sibling:male:younger>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '堂弟媳',
    terminal: true
  },
  'parent:male>sibling:male:older>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '堂姐夫',
    terminal: true
  },
  'parent:male>sibling:male:younger>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '堂妹夫',
    terminal: true
  },
  'parent:male>sibling:female:older>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '表嫂',
    terminal: true
  },
  'parent:male>sibling:female:younger>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '表弟媳',
    terminal: true
  },
  'parent:male>sibling:female:older>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '表姐夫',
    terminal: true
  },
  'parent:male>sibling:female:younger>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '表妹夫',
    terminal: true
  },
  'parent:female>sibling:male:older>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '表嫂',
    terminal: true
  },
  'parent:female>sibling:male:younger>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '表弟媳',
    terminal: true
  },
  'parent:female>sibling:male:older>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '表姐夫',
    terminal: true
  },
  'parent:female>sibling:male:younger>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '表妹夫',
    terminal: true
  },
  'parent:female>sibling:female:older>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '表嫂',
    terminal: true
  },
  'parent:female>sibling:female:younger>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '表弟媳',
    terminal: true
  },
  'parent:female>sibling:female:older>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '表姐夫',
    terminal: true
  },
  'parent:female>sibling:female:younger>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '表妹夫',
    terminal: true
  },

  'child:male>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '孙媳',
    terminal: true
  },
  'child:male>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '孙女婿',
    terminal: true
  },
  'child:female>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '外孙媳',
    terminal: true
  },
  'child:female>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '外孙女婿',
    terminal: true
  },
  'child:male>child:male>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '曾孙媳',
    terminal: true
  },
  'child:male>child:male>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '曾孙女婿',
    terminal: true
  },
  'child:female>child:male>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '外曾孙媳',
    terminal: true
  },
  'child:female>child:female>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '外曾孙女婿',
    terminal: true
  },

  'spouse:female>sibling:male:older>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '妻侄媳妇',
    terminal: true
  },
  'spouse:female>sibling:male:younger>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '妻侄媳妇',
    terminal: true
  },
  'spouse:female>sibling:male:older>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '妻侄女婿',
    terminal: true
  },
  'spouse:female>sibling:male:younger>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '妻侄女婿',
    terminal: true
  },
  'spouse:female>sibling:female:older>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '妻外甥媳妇',
    terminal: true
  },
  'spouse:female>sibling:female:younger>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '妻外甥媳妇',
    terminal: true
  },
  'spouse:female>sibling:female:older>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '妻外甥女婿',
    terminal: true
  },
  'spouse:female>sibling:female:younger>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '妻外甥女婿',
    terminal: true
  },
  'spouse:male>sibling:male:older>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '夫侄媳妇',
    terminal: true
  },
  'spouse:male>sibling:male:younger>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '夫侄媳妇',
    terminal: true
  },
  'spouse:male>sibling:male:older>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '夫侄女婿',
    terminal: true
  },
  'spouse:male>sibling:male:younger>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '夫侄女婿',
    terminal: true
  },
  'spouse:male>sibling:female:older>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '夫外甥媳妇',
    terminal: true
  },
  'spouse:male>sibling:female:younger>child:male>spouse:female': {
    status: 'resolved',
    primaryTitle: '夫外甥媳妇',
    terminal: true
  },
  'spouse:male>sibling:female:older>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '夫外甥女婿',
    terminal: true
  },
  'spouse:male>sibling:female:younger>child:female>spouse:male': {
    status: 'resolved',
    primaryTitle: '夫外甥女婿',
    terminal: true
  },

  'spouse:female>sibling:male:older>child:male': {
    status: 'ambiguous',
    primaryTitle: '妻侄',
    candidates: ['妻侄', '内侄', '姻亲晚辈'],
    explanation: '配偶旁系晚辈称谓存在地区差异，当前版本给出保守结果。'
  },
  'spouse:male>sibling:female:older>child:female': {
    status: 'ambiguous',
    primaryTitle: '夫家外甥女',
    candidates: ['夫家外甥女', '姻亲晚辈', '夫家表亲'],
    explanation: '配偶旁系晚辈称谓存在地区差异，当前版本给出保守结果。'
  },
  'spouse:female>parent:male>sibling:male:older': {
    status: 'ambiguous',
    primaryTitle: '配偶父辈旁系',
    candidates: ['配偶伯父', '配偶叔父', '配偶父辈旁系', '姻亲长辈'],
    explanation: '配偶父辈旁系称谓存在地区差异，当前版本给出保守结果。'
  },
  'spouse:male>parent:female>sibling:female:older': {
    status: 'ambiguous',
    primaryTitle: '配偶母辈旁系',
    candidates: ['配偶姑母', '配偶姨母', '配偶母辈旁系', '姻亲长辈'],
    explanation: '配偶母辈旁系称谓存在地区差异，当前版本给出保守结果。'
  },

  'spouse:female>sibling:male:older>child:female': {
    status: 'ambiguous',
    primaryTitle: '姻亲晚辈',
    candidates: ['妻侄女', '内侄女', '姻亲晚辈'],
    explanation: '配偶旁系晚辈称谓存在地区差异。'
  },
  'spouse:male>sibling:female:younger>child:male': {
    status: 'ambiguous',
    primaryTitle: '姻亲晚辈',
    candidates: ['夫家外甥', '姻亲晚辈', '夫家表亲'],
    explanation: '配偶旁系晚辈称谓存在地区差异。'
  },

  'spouse:female>parent:male': {
    status: 'resolved',
    primaryTitle: '岳父',
    aliases: ['丈人']
  },
  'spouse:female>parent:female': {
    status: 'resolved',
    primaryTitle: '岳母',
    aliases: ['丈母娘']
  },
  'spouse:male>parent:male': {
    status: 'resolved',
    primaryTitle: '公公'
  },
  'spouse:male>parent:female': {
    status: 'resolved',
    primaryTitle: '婆婆'
  },

  'spouse:female>sibling:male:older': { status: 'resolved', primaryTitle: '大舅子', aliases: ['内兄'] },
  'spouse:female>sibling:male:younger': { status: 'resolved', primaryTitle: '小舅子', aliases: ['内弟'] },
  'spouse:male>sibling:male:older': { status: 'resolved', primaryTitle: '大伯子', aliases: ['大伯哥'] },
  'spouse:male>sibling:male:younger': { status: 'resolved', primaryTitle: '小叔子' },

  'spouse:female>sibling:female:older>child:male': {
    status: 'ambiguous',
    primaryTitle: '姻亲晚辈',
    candidates: ['姨侄', '姻亲晚辈', '内侄'],
    explanation: '配偶姐妹的子女称谓存在地区差异。'
  },
  'spouse:male>sibling:male:younger>child:female': {
    status: 'ambiguous',
    primaryTitle: '姻亲晚辈',
    candidates: ['叔侄女', '姻亲晚辈', '夫家侄辈'],
    explanation: '配偶兄弟的子女称谓存在地区差异。'
  }
}

const GENERIC_TERMINAL_AFFINITY_RULES: Record<string, KinshipRuleMatch> = {
  'child>spouse': {
    status: 'ambiguous',
    primaryTitle: '子女的配偶',
    candidates: ['儿媳', '女婿', '子女的配偶'],
    terminal: true,
    explanation: '子女的配偶需要补充子女性别和配偶性别。'
  },
  'sibling>spouse': {
    status: 'ambiguous',
    primaryTitle: '兄弟姐妹的配偶',
    candidates: ['嫂子', '弟媳', '姐夫', '妹夫', '兄弟姐妹的配偶'],
    terminal: true,
    explanation: '兄弟姐妹的配偶需要补充性别和长幼信息。'
  },
  'parent>sibling>spouse': {
    status: 'ambiguous',
    primaryTitle: '父母旁系的配偶',
    candidates: ['伯母', '婶婶', '姑父', '舅妈', '姨父', '父母旁系的配偶'],
    terminal: true,
    explanation: '父母旁系的配偶需要补充父母性别、旁系性别和长幼信息。'
  },
  'spouse>sibling>spouse': {
    status: 'ambiguous',
    primaryTitle: '配偶兄弟姐妹的配偶',
    candidates: ['内兄嫂', '内弟媳', '连襟', '妯娌', '姑姐夫', '姑妹夫'],
    terminal: true,
    explanation: '配偶兄弟姐妹的配偶需要补充配偶性别、旁系性别和长幼信息。'
  },
  'sibling>child>spouse': {
    status: 'ambiguous',
    primaryTitle: '侄甥配偶',
    candidates: ['侄媳妇', '侄女婿', '外甥媳妇', '外甥女婿', '侄甥配偶'],
    terminal: true,
    explanation: '侄甥配偶需要补充兄弟姐妹性别、子女性别和配偶性别。'
  },
  'sibling>child>child>spouse': {
    status: 'ambiguous',
    primaryTitle: '侄甥孙辈配偶',
    candidates: ['侄孙媳', '侄孙女婿', '外甥孙媳', '外甥孙女婿', '侄甥孙辈配偶'],
    terminal: true,
    explanation: '侄甥孙辈配偶需要补充兄弟姐妹性别、后代性别和配偶性别。'
  },
  'parent>parent>sibling>spouse': {
    status: 'ambiguous',
    primaryTitle: '祖辈旁系配偶',
    candidates: ['伯祖母', '叔祖母', '姑祖父', '舅祖母', '姨祖父', '祖辈旁系配偶'],
    terminal: true,
    explanation: '祖辈旁系配偶称谓存在地区差异，当前版本给出保守结果。'
  },
  'parent>parent>parent>sibling>spouse': {
    status: 'ambiguous',
    primaryTitle: '曾祖辈旁系配偶',
    candidates: ['曾伯祖母', '曾叔祖母', '曾姑祖父', '曾舅祖母', '曾姨祖父', '曾祖辈旁系配偶'],
    terminal: true,
    explanation: '曾祖辈旁系配偶称谓存在地区差异，当前版本给出保守结果。'
  },
  'parent>sibling>child>spouse': {
    status: 'ambiguous',
    primaryTitle: '堂表亲配偶',
    candidates: ['堂嫂', '堂弟媳', '堂姐夫', '堂妹夫', '表嫂', '表弟媳', '表姐夫', '表妹夫'],
    terminal: true,
    explanation: '堂表亲配偶需要补充父母性别、旁系性别、子女性别和长幼信息。'
  },
  'parent>sibling>child>child>spouse': {
    status: 'ambiguous',
    primaryTitle: '堂表亲侄辈配偶',
    candidates: ['堂侄媳', '堂侄女婿', '表侄媳', '表侄女婿', '堂表亲侄辈配偶'],
    terminal: true,
    explanation: '堂表亲侄辈配偶需要补充堂表方向、后代性别和配偶性别。'
  },
  'parent>sibling>child>child>child>spouse': {
    status: 'ambiguous',
    primaryTitle: '堂表亲侄孙辈配偶',
    candidates: ['堂侄孙媳', '堂侄孙女婿', '表侄孙媳', '表侄孙女婿', '堂表亲侄孙辈配偶'],
    terminal: true,
    explanation: '堂表亲侄孙辈配偶需要补充堂表方向、后代性别和配偶性别。'
  },
  'child>child>spouse': {
    status: 'ambiguous',
    primaryTitle: '孙辈配偶',
    candidates: ['孙媳', '孙女婿', '外孙媳', '外孙女婿', '孙辈配偶'],
    terminal: true,
    explanation: '孙辈配偶需要补充子女性别、孙辈性别和配偶性别。'
  },
  'child>child>child>spouse': {
    status: 'ambiguous',
    primaryTitle: '重孙辈配偶',
    candidates: ['曾孙媳', '曾孙女婿', '外曾孙媳', '外曾孙女婿', '重孙辈配偶'],
    terminal: true,
    explanation: '重孙辈配偶需要补充直系下行性别和配偶性别。'
  },
  'spouse>sibling>child>spouse': {
    status: 'ambiguous',
    primaryTitle: '配偶侄甥配偶',
    candidates: ['妻侄媳妇', '妻侄女婿', '妻外甥媳妇', '妻外甥女婿', '夫侄媳妇', '夫侄女婿', '夫外甥媳妇', '夫外甥女婿'],
    terminal: true,
    explanation: '配偶侄甥配偶需要补充配偶侧亲属性别和长幼信息。'
  },
  'sibling>child>child>child>spouse': {
    status: 'ambiguous',
    primaryTitle: '侄重孙辈配偶',
    candidates: ['侄重孙媳', '侄重孙女婿', '外甥重孙媳', '外甥重孙女婿', '侄重孙辈配偶'],
    terminal: true,
    explanation: '侄重孙辈配偶需要补充兄弟姐妹性别、后代性别和配偶性别。'
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

function getSpouseTitleByGender(
  spouseGender: string | undefined,
  maleTitle: string,
  femaleTitle: string,
  neutralTitle: string
): KinshipRuleMatch {
  if (spouseGender === 'female') {
    return { status: 'resolved', primaryTitle: femaleTitle, terminal: true }
  }
  if (spouseGender === 'male') {
    return { status: 'resolved', primaryTitle: maleTitle, terminal: true }
  }
  return {
    status: 'ambiguous',
    primaryTitle: neutralTitle,
    candidates: [neutralTitle, femaleTitle, maleTitle],
    terminal: true,
    explanation: '需要补充配偶性别以区分更具体的称谓。'
  }
}

function matchCousinDescendantSpouse(pathKey: string, relationKey: string): KinshipRuleMatch | null {
  if (relationKey !== 'parent>sibling>child>child>spouse' && relationKey !== 'parent>sibling>child>child>child>spouse') {
    return null
  }

  const side = getCousinSide(pathKey)
  const spouseGender = pathKey.split('>').pop()?.split(':')[1]

  if (relationKey === 'parent>sibling>child>child>spouse') {
    if (side === 'paternal') {
      return getSpouseTitleByGender(spouseGender, '堂侄女婿', '堂侄媳', '堂侄配偶')
    }
    if (side === 'maternalOrCross') {
      return getSpouseTitleByGender(spouseGender, '表侄女婿', '表侄媳', '表侄配偶')
    }
    return {
      status: 'ambiguous',
      primaryTitle: '堂表亲侄辈配偶',
      candidates: ['堂侄媳', '堂侄女婿', '表侄媳', '表侄女婿', '堂表亲侄辈配偶'],
      terminal: true,
      explanation: '需要补充父母性别、旁系性别和配偶性别以区分堂亲或表亲侄辈配偶。'
    }
  }

  if (side === 'paternal') {
    return getSpouseTitleByGender(spouseGender, '堂侄孙女婿', '堂侄孙媳', '堂侄孙辈配偶')
  }
  if (side === 'maternalOrCross') {
    return getSpouseTitleByGender(spouseGender, '表侄孙女婿', '表侄孙媳', '表侄孙辈配偶')
  }
  return {
    status: 'ambiguous',
    primaryTitle: '堂表亲侄孙辈配偶',
    candidates: ['堂侄孙媳', '堂侄孙女婿', '表侄孙媳', '表侄孙女婿', '堂表亲侄孙辈配偶'],
    terminal: true,
    explanation: '需要补充父母性别、旁系性别和配偶性别以区分堂亲或表亲侄孙辈配偶。'
  }
}

export function matchAffinityRules(context: KinshipContext, pathKey: string): KinshipRuleMatch | null {
  if (pathKey in AFFINITY_RULES) {
    return AFFINITY_RULES[pathKey]
  }
  const relationKey = buildRelationKey(context)
  return matchCousinDescendantSpouse(pathKey, relationKey) || GENERIC_TERMINAL_AFFINITY_RULES[relationKey] || null
}

export { AFFINITY_RULES, GENERIC_TERMINAL_AFFINITY_RULES }
