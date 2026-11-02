import type { CanonicalKinshipResult } from './canonicalTypes'
import type { KinshipResolution } from './types'

/** 将规范结果映射为旧版 KinshipResolution（aliases/candidates 恒为空，仅供过渡）。 */
export function canonicalToKinshipResolution(result: CanonicalKinshipResult): KinshipResolution {
  if (!result.unsupportedCanonicalTitle && result.canonicalTitle) {
    return {
      status: 'resolved',
      primaryTitle: result.canonicalTitle,
      aliases: [],
      pathDescription: result.pathDescription,
      explanation: result.rankLabel
        ? `排行：${result.rankLabel}（与称谓分开展示）`
        : result.explanation
    }
  }

  if (result.incompleteInfo) {
    return {
      status: 'unsupported',
      primaryTitle: undefined,
      aliases: [],
      pathDescription: result.pathDescription,
      explanation: result.explanation || '关系信息不完整',
      incompleteInfo: true
    }
  }

  return {
    status: 'unsupported',
    primaryTitle: undefined,
    aliases: [],
    pathDescription: result.pathDescription,
    explanation: result.explanation || '暂无对应规范称谓',
    incompleteInfo: false
  }
}
