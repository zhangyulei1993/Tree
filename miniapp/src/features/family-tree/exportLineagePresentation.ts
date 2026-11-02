import type { ExportLineageNode } from './exportLineageTree'

export type ExportLineageScope = 'INTERNAL' | 'SHARE' | 'PUBLIC'

export interface ExportLineageDisplay {
  name: string
  yearText: string
  parentLabels: string[]
  spouseLabels: string[]
}

function maskName(name: string): string {
  const chars = [...name.trim()]
  if (chars.length === 0) return '匿名成员'
  return `${chars[0]}*`
}

function maskRelationLabel(label: string): string {
  const separator = label.indexOf('：')
  if (separator < 0) return maskName(label)
  return `${label.slice(0, separator + 1)}${maskName(label.slice(separator + 1))}`
}

function fullYearText(node: ExportLineageNode['member']): string {
  const birth = node.birthDate?.slice(0, 4)
  const death = node.deathDate?.slice(0, 4)
  if (!death) return ''
  return birth ? `${birth}-${death}` : `卒：${death}`
}

function spouseLabelsForScope(labels: string[], scope: ExportLineageScope): string[] {
  // Former marriages are retained only in the private family archive.
  return scope === 'INTERNAL' ? labels : labels.filter((label) => !/^前[妻夫]：/.test(label))
}

/** 同一条宗支数据按使用范围输出不同的信息密度。 */
export function getExportLineageDisplay(item: ExportLineageNode, scope: ExportLineageScope): ExportLineageDisplay {
  if (scope === 'PUBLIC') {
    return {
      name: maskName(item.member.displayName),
      yearText: '',
      parentLabels: item.parentLabels.map(maskRelationLabel),
      spouseLabels: spouseLabelsForScope(item.spouseLabels, scope).map(maskRelationLabel)
    }
  }
  if (scope === 'SHARE') {
    return {
      name: item.member.displayName,
      yearText: fullYearText(item.member),
      parentLabels: item.parentLabels,
      spouseLabels: spouseLabelsForScope(item.spouseLabels, scope)
    }
  }
  return {
    name: item.member.displayName,
    yearText: fullYearText(item.member),
    parentLabels: item.parentLabels,
    spouseLabels: item.spouseLabels
  }
}
