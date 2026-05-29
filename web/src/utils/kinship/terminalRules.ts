import type { KinshipContext, KinshipStep } from './types'
import { getSteps, pathKey } from './helpers'

export const TERMINAL_TITLES = new Set([
  '连襟', '妯娌', '内兄嫂', '内弟媳', '姑姐夫', '姑妹夫',
  '嫂子', '弟媳', '姐夫', '妹夫',
  '侄媳妇', '侄女婿', '外甥媳妇', '外甥女婿',
  '堂嫂', '堂弟媳', '堂姐夫', '堂妹夫',
  '表嫂', '表弟媳', '表姐夫', '表妹夫',
  '儿媳', '女婿', '孙媳', '孙女婿', '外孙媳', '外孙女婿'
])

const TERMINAL_PATH_KEYS = new Set([
  'sibling/spouse',
  'sibling/child/spouse',
  'spouse/sibling/spouse',
  'parent/sibling/child/spouse',
  'parent/parent/sibling/child/spouse',
  'parent/parent/sibling/spouse',
  'child/spouse',
  'child/child/spouse'
])

export function isTerminalTitle(title: string) {
  return TERMINAL_TITLES.has(title)
}

export function terminalReasonByTitle(title: string) {
  if (!isTerminalTitle(title)) return ''
  return '该关系已属于姻亲终点，不建议继续扩展。'
}

export function isTerminalPath(path: KinshipStep[]) {
  return TERMINAL_PATH_KEYS.has(pathKey(path))
}

export function isTerminalContext(context: KinshipContext) {
  return isTerminalPath(getSteps(context))
}
