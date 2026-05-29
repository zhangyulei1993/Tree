import type { AgeOrder, Gender, KinshipContext, KinshipResult, KinshipStep } from './types'
import {
  cleanProcess,
  comparePeople,
  containsOnly,
  endByGender,
  getSteps,
  hasUnknownGender,
  isKnownGender,
  MAX_KINSHIP_LEVEL,
  pathKey,
  pathText,
  titleByGenderAndOrder
} from './helpers'
import { allowedRelations } from './allowedRelations'
import { isTerminalPath, isTerminalTitle, terminalReasonByTitle } from './terminalRules'

function result(
  context: KinshipContext,
  title: string,
  matched: boolean,
  terminal: boolean,
  reason: string,
  normalizedPath: KinshipStep[],
  process: string[]
): KinshipResult {
  const terminalByTitle = isTerminalTitle(title)
  const finalTerminal = terminal || terminalByTitle || isTerminalPath(normalizedPath)
  const finalReason = reason || terminalReasonByTitle(title)

  return {
    title,
    matched,
    terminal: finalTerminal,
    reason: finalReason,
    process: cleanProcess(process),
    allowedNextRelations: allowedRelations({ self: context.self, steps: normalizedPath }),
    displayPath: pathText(normalizedPath),
    normalizedPath
  }
}

function pending(
  context: KinshipContext,
  title: string,
  reasonText: string,
  normalizedPath: KinshipStep[],
  process: string[]
) {
  return result(context, title, false, false, reasonText, normalizedPath, process)
}

function known(gender: Gender) {
  return isKnownGender(gender)
}

function directParent(context: KinshipContext, step: KinshipStep, path: KinshipStep[], process: string[]) {
  if (step.gender === 'male') return result(context, '父亲', true, false, '', path, process)
  if (step.gender === 'female') return result(context, '母亲', true, false, '', path, process)
  return pending(context, '父母', '请选择父母的性别，用于区分父亲或母亲。', path, process)
}

function directChild(context: KinshipContext, step: KinshipStep, path: KinshipStep[], process: string[]) {
  if (step.gender === 'male') return result(context, '儿子', true, false, '', path, process)
  if (step.gender === 'female') return result(context, '女儿', true, false, '', path, process)
  return pending(context, '子女', '请选择子女的性别，用于区分儿子或女儿。', path, process)
}

function directSibling(context: KinshipContext, step: KinshipStep, path: KinshipStep[], process: string[]) {
  const order = comparePeople(step, context.self)
  const title = titleByGenderAndOrder(
    step.gender,
    order,
    '哥哥',
    '弟弟',
    '兄弟',
    '姐姐',
    '妹妹',
    '姐妹',
    '兄弟姐妹'
  )

  if (step.gender === 'unknown') {
    return pending(context, title, '请选择兄弟姐妹的性别。若要区分长幼，请补充出生日期或年龄。', path, process)
  }

  if (order === 'unknown' || order === 'same') {
    return pending(context, title, '请补充本人和该兄弟姐妹的出生日期或年龄，用于区分哥哥、弟弟、姐姐或妹妹。', path, process)
  }

  return result(context, title, true, false, '', path, process)
}

function directSpouse(context: KinshipContext, path: KinshipStep[], process: string[]) {
  if (context.self.gender === 'male') return result(context, '妻子', true, false, '', path, process)
  if (context.self.gender === 'female') return result(context, '丈夫', true, false, '', path, process)
  return pending(context, '配偶', '请先选择本人性别，用于区分丈夫或妻子。', path, process)
}

const ancestorPrefix3: Record<string, string> = {
  'male/male': '曾祖',
  'male/female': '曾外祖',
  'female/male': '外曾祖',
  'female/female': '外曾外祖'
}

const ancestorPrefix4: Record<string, string> = {
  'male/male/male': '高祖',
  'male/male/female': '高外祖',
  'male/female/male': '曾外高祖',
  'male/female/female': '曾外曾外祖',
  'female/male/male': '外高祖',
  'female/male/female': '外高外祖',
  'female/female/male': '外高外祖',
  'female/female/female': '外曾外曾祖'
}

function ancestorTitle(path: KinshipStep[]) {
  if (!containsOnly(path, 'parent')) return null
  if (hasUnknownGender(path)) return null

  const last = path[path.length - 1]

  if (path.length === 1) {
    if (last.gender === 'male') return '父亲'
    if (last.gender === 'female') return '母亲'
  }

  if (path.length === 2) {
    const first = path[0].gender

    if (first === 'male' && last.gender === 'male') return '爷爷'
    if (first === 'male' && last.gender === 'female') return '奶奶'
    if (first === 'female' && last.gender === 'male') return '外公'
    if (first === 'female' && last.gender === 'female') return '外婆'
  }

  if (path.length === 3) {
    const prefix = ancestorPrefix3[path.slice(0, 2).map((item) => item.gender).join('/')]
    return prefix ? prefix + endByGender(last.gender) : null
  }

  if (path.length === 4) {
    const prefix = ancestorPrefix4[path.slice(0, 3).map((item) => item.gender).join('/')]
    return prefix ? prefix + endByGender(last.gender) : null
  }

  if (path.length === 5) {
    const purePaternal = path.slice(0, 4).every((item) => item.gender === 'male')
    return (purePaternal ? '五世祖' : '外五世祖') + endByGender(last.gender)
  }

  return null
}

function resolveAncestor(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const title = ancestorTitle(path)
  if (title) return result(context, title, true, path.length >= MAX_KINSHIP_LEVEL, '', path, process)

  const levelName = path.length === 2 ? '祖辈' : path.length === 3 ? '曾祖辈' : path.length === 4 ? '高祖辈' : '五世祖辈'
  return pending(context, levelName, `请补充第 ${path.findIndex((item) => !known(item.gender)) + 1} 层父母的性别，用于精确区分祖辈称谓。`, path, process)
}

function titleBasePrefix(title: string) {
  if (title === '爷爷' || title === '奶奶') return ''
  if (title === '外公' || title === '外婆') return '外'
  return title
    .replace(/祖父$/, '')
    .replace(/祖母$/, '')
    .replace(/父$/, '')
    .replace(/母$/, '')
}

function resolveAncestorSibling(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const sibling = path[path.length - 1]
  const chain = path.slice(0, -1)
  const ancestor = chain[chain.length - 1]
  const chainTitle = ancestorTitle(chain)

  if (!chainTitle || !known(ancestor.gender) || !known(sibling.gender)) {
    return pending(context, '祖辈旁系亲属', '请补充祖辈链条和其兄弟姐妹的性别，用于区分伯祖父、叔祖父、姑祖母、舅祖父或姨祖母。', path, process)
  }

  const prefix = titleBasePrefix(chainTitle)
  const order = comparePeople(sibling, ancestor)

  let title = ''

  if (ancestor.gender === 'male' && sibling.gender === 'male') {
    if (order === 'older') title = prefix + '伯祖父'
    else if (order === 'younger') title = prefix + '叔祖父'
    else title = prefix + '叔伯祖父'
  }

  if (ancestor.gender === 'male' && sibling.gender === 'female') title = prefix + '姑祖母'
  if (ancestor.gender === 'female' && sibling.gender === 'male') title = prefix + '舅祖父'
  if (ancestor.gender === 'female' && sibling.gender === 'female') title = prefix + '姨祖母'

  if (!title) return pending(context, '祖辈旁系亲属', '当前祖辈旁系关系缺少必要信息。', path, process)
  return result(context, title, true, path.length >= MAX_KINSHIP_LEVEL, '', path, process)
}


function resolveAncestorSiblingSpouse(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const siblingPath = path.slice(0, -1)
  const siblingResult = resolveAncestorSibling(context, siblingPath, process)
  const base = siblingResult.title

  const map: Array<[string, string]> = [
    ['伯祖父', '伯祖母'],
    ['叔祖父', '叔祖母'],
    ['姑祖母', '姑祖父'],
    ['舅祖父', '舅祖母'],
    ['姨祖母', '姨祖父']
  ]

  const hit = map.find(([suffix]) => base.endsWith(suffix))
  if (!hit) return pending(context, '祖辈旁系姻亲', '请先补充祖辈旁系的性别和长幼信息。', path, process)

  const title = base.slice(0, base.length - hit[0].length) + hit[1]
  return result(context, title, true, true, '', path, process)
}

function resolveParentSiblingSpouseChild(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const normalized = [path[0], path[1], path[3]]
  const title = cousinTitle(context, normalized)

  if (!title) {
    return pending(context, '堂表亲', '请补充父母、父母兄弟姐妹及其子女的性别。若要区分哥哥、弟弟、姐姐、妹妹，请补充出生日期或年龄。', path, process)
  }

  if (title.endsWith('兄弟') || title.endsWith('姐妹')) {
    return pending(context, title, '请补充本人和该堂表亲的出生日期或年龄，用于区分长幼。', path, process)
  }

  return result(context, title, true, false, '', path, process)
}

function resolveParentSibling(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const parent = path[0]
  const sibling = path[1]

  if (!known(parent.gender) || !known(sibling.gender)) {
    return pending(context, '父母的兄弟姐妹', '请补充父母以及父母兄弟姐妹的性别。', path, process)
  }

  const order = comparePeople(sibling, parent)
  let title = ''

  if (parent.gender === 'male' && sibling.gender === 'male') {
    if (order === 'older') title = '伯父'
    else if (order === 'younger') title = '叔叔'
    else title = '叔伯'
  }

  if (parent.gender === 'male' && sibling.gender === 'female') title = '姑姑'
  if (parent.gender === 'female' && sibling.gender === 'male') title = '舅舅'

  if (parent.gender === 'female' && sibling.gender === 'female') {
    if (order === 'older') title = '姨妈'
    else if (order === 'younger') title = '小姨'
    else title = '姨母'
  }

  if (!title) return pending(context, '父母的兄弟姐妹', '当前路径无法精确匹配父母旁系称谓。', path, process)
  return result(context, title, true, false, '', path, process)
}

function resolveParentSiblingSpouse(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const parent = path[0]
  const sibling = path[1]

  if (!known(parent.gender) || !known(sibling.gender)) {
    return pending(context, '父母旁系姻亲', '请补充父母以及父母兄弟姐妹的性别。', path, process)
  }

  const order = comparePeople(sibling, parent)
  let title = ''

  if (parent.gender === 'male' && sibling.gender === 'male') title = order === 'older' ? '伯母' : order === 'younger' ? '婶婶' : '伯母/婶婶'
  if (parent.gender === 'male' && sibling.gender === 'female') title = '姑父'
  if (parent.gender === 'female' && sibling.gender === 'male') title = '舅妈'
  if (parent.gender === 'female' && sibling.gender === 'female') title = '姨父'

  return result(context, title || '父母旁系姻亲', Boolean(title), false, title ? '' : '当前路径无法精确匹配。', path, process)
}

function cousinStem(parent: KinshipStep, parentSibling: KinshipStep) {
  if (parent.gender === 'male' && parentSibling.gender === 'male') return '堂'
  return '表'
}

function cousinTitle(context: KinshipContext, path: KinshipStep[]) {
  const parent = path[0]
  const parentSibling = path[1]
  const cousin = path[2]

  if (!known(parent.gender) || !known(parentSibling.gender) || !known(cousin.gender)) return null

  const stem = cousinStem(parent, parentSibling)
  const order = comparePeople(cousin, context.self)

  if (cousin.gender === 'male') {
    if (order === 'older') return stem + '哥'
    if (order === 'younger') return stem + '弟'
    return stem + '兄弟'
  }

  if (order === 'older') return stem + '姐'
  if (order === 'younger') return stem + '妹'
  return stem + '姐妹'
}

function resolveCousin(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const title = cousinTitle(context, path)

  if (!title) {
    return pending(context, '堂表亲', '请补充父母、父母兄弟姐妹及其子女的性别。若要区分哥哥、弟弟、姐姐、妹妹，请补充出生日期或年龄。', path, process)
  }

  if (title.endsWith('兄弟') || title.endsWith('姐妹')) {
    return pending(context, title, '请补充本人和该堂表亲的出生日期或年龄，用于区分长幼。', path, process)
  }

  return result(context, title, true, false, '', path, process)
}

function resolveCousinSpouse(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const title = cousinTitle(context, path.slice(0, 3))
  const map: Record<string, string> = {
    堂哥: '堂嫂',
    堂弟: '堂弟媳',
    堂姐: '堂姐夫',
    堂妹: '堂妹夫',
    表哥: '表嫂',
    表弟: '表弟媳',
    表姐: '表姐夫',
    表妹: '表妹夫'
  }

  if (title && map[title]) return result(context, map[title], true, true, '', path, process)
  return pending(context, '堂表亲配偶', '请先补充堂表亲的性别和长幼信息，再判断其配偶称谓。', path, process)
}

function resolveSpouseParent(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const parent = path[1]

  if (context.self.gender === 'male') {
    if (parent.gender === 'male') return result(context, '岳父', true, false, '', path, process)
    if (parent.gender === 'female') return result(context, '岳母', true, false, '', path, process)
    return pending(context, '岳父母', '请选择配偶父母的性别，用于区分岳父或岳母。', path, process)
  }

  if (context.self.gender === 'female') {
    if (parent.gender === 'male') return result(context, '公公', true, false, '', path, process)
    if (parent.gender === 'female') return result(context, '婆婆', true, false, '', path, process)
    return pending(context, '公婆', '请选择配偶父母的性别，用于区分公公或婆婆。', path, process)
  }

  return pending(context, '配偶父母', '请先选择本人性别，用于区分岳父母或公婆。', path, process)
}

function spouseSiblingTitle(context: KinshipContext, spouse: KinshipStep, sibling: KinshipStep) {
  const order = comparePeople(sibling, spouse)

  if (context.self.gender === 'male') {
    if (sibling.gender === 'male') {
      if (order === 'older') return '大舅子'
      if (order === 'younger') return '小舅子'
      return '内兄弟'
    }

    if (sibling.gender === 'female') {
      if (order === 'older') return '大姨子'
      if (order === 'younger') return '小姨子'
      return '姨姐妹'
    }
  }

  if (context.self.gender === 'female') {
    if (sibling.gender === 'male') {
      if (order === 'older') return '大伯哥'
      if (order === 'younger') return '小叔子'
      return '夫兄弟'
    }

    if (sibling.gender === 'female') {
      if (order === 'older') return '大姑姐'
      if (order === 'younger') return '小姑妹'
      return '姑姐妹'
    }
  }

  if (sibling.gender === 'male') return '配偶的兄弟'
  if (sibling.gender === 'female') return '配偶的姐妹'
  return '配偶的兄弟姐妹'
}

function resolveSpouseSibling(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const spouse = path[0]
  const sibling = path[1]

  if (!known(context.self.gender) || !known(sibling.gender)) {
    return pending(context, '配偶的兄弟姐妹', '请补充本人性别、配偶兄弟姐妹的性别，并尽量补充长幼信息。', path, process)
  }

  const title = spouseSiblingTitle(context, spouse, sibling)
  const order = comparePeople(sibling, spouse)

  if (order === 'unknown' || order === 'same') {
    return pending(context, title, '请补充配偶与其兄弟姐妹的出生日期或年龄，用于区分长幼。', path, process)
  }

  return result(context, title, true, false, '', path, process)
}

function resolveSpouseSiblingSpouse(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const spouse = path[0]
  const sibling = path[1]
  const order = comparePeople(sibling, spouse)

  if (!known(context.self.gender) || !known(sibling.gender)) {
    return pending(context, '配偶兄弟姐妹的配偶', '请补充本人性别、配偶兄弟姐妹的性别和长幼信息。', path, process)
  }

  if (context.self.gender === 'male') {
    if (sibling.gender === 'male') {
      if (order === 'older') return result(context, '内兄嫂', true, true, '', path, process)
      if (order === 'younger') return result(context, '内弟媳', true, true, '', path, process)
      return pending(context, '内兄嫂/内弟媳', '请补充配偶与其兄弟的出生日期或年龄，用于区分内兄或内弟。', path, process)
    }

    if (sibling.gender === 'female') return result(context, '连襟', true, true, '', path, process)
  }

  if (context.self.gender === 'female') {
    if (sibling.gender === 'male') return result(context, '妯娌', true, true, '', path, process)

    if (sibling.gender === 'female') {
      if (order === 'older') return result(context, '姑姐夫', true, true, '', path, process)
      if (order === 'younger') return result(context, '姑妹夫', true, true, '', path, process)
      return pending(context, '姑姐夫/姑妹夫', '请补充配偶与其姐妹的出生日期或年龄，用于区分姑姐或姑妹。', path, process)
    }
  }

  return pending(context, '配偶兄弟姐妹的配偶', '当前信息不足以判断配偶兄弟姐妹的配偶称谓。', path, process)
}

function resolveSiblingSpouse(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const sibling = path[0]
  const order = comparePeople(sibling, context.self)

  if (sibling.gender === 'male') {
    if (order === 'older') return result(context, '嫂子', true, true, '', path, process)
    if (order === 'younger') return result(context, '弟媳', true, true, '', path, process)
    return pending(context, '兄弟的配偶', '请补充本人和兄弟的出生日期或年龄，用于区分嫂子或弟媳。', path, process)
  }

  if (sibling.gender === 'female') {
    if (order === 'older') return result(context, '姐夫', true, true, '', path, process)
    if (order === 'younger') return result(context, '妹夫', true, true, '', path, process)
    return pending(context, '姐妹的配偶', '请补充本人和姐妹的出生日期或年龄，用于区分姐夫或妹夫。', path, process)
  }

  return pending(context, '兄弟姐妹的配偶', '请补充兄弟姐妹的性别和长幼信息。', path, process)
}

function resolveSiblingChild(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const sibling = path[0]
  const child = path[1]

  if (!known(sibling.gender) || !known(child.gender)) {
    return pending(context, '侄甥', '请补充兄弟姐妹及其子女的性别。', path, process)
  }

  if (sibling.gender === 'male') {
    return result(context, child.gender === 'male' ? '侄子' : '侄女', true, false, '', path, process)
  }

  return result(context, child.gender === 'male' ? '外甥' : '外甥女', true, false, '', path, process)
}

function resolveSiblingChildSpouse(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const sibling = path[0]
  const child = path[1]

  if (!known(sibling.gender) || !known(child.gender)) {
    return pending(context, '侄甥配偶', '请补充兄弟姐妹及其子女的性别。', path, process)
  }

  if (sibling.gender === 'male' && child.gender === 'male') return result(context, '侄媳妇', true, true, '', path, process)
  if (sibling.gender === 'male' && child.gender === 'female') return result(context, '侄女婿', true, true, '', path, process)
  if (sibling.gender === 'female' && child.gender === 'male') return result(context, '外甥媳妇', true, true, '', path, process)
  return result(context, '外甥女婿', true, true, '', path, process)
}

function resolveDescendant(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const first = path[0]
  const last = path[path.length - 1]

  if (!known(last.gender)) {
    return pending(context, '后代', '请补充后代性别，用于精确区分称谓。', path, process)
  }

  if (path.length === 1) return directChild(context, last, path, process)

  if (path.length === 2) {
    if (!known(first.gender)) return pending(context, '孙辈', '请补充第一层子女的性别，用于区分孙辈或外孙辈。', path, process)
    if (first.gender === 'male') return result(context, last.gender === 'male' ? '孙子' : '孙女', true, false, '', path, process)
    return result(context, last.gender === 'male' ? '外孙' : '外孙女', true, false, '', path, process)
  }

  if (path.length === 3) return result(context, last.gender === 'male' ? '曾孙' : '曾孙女', true, false, '', path, process)
  if (path.length === 4) return result(context, last.gender === 'male' ? '玄孙' : '玄孙女', true, false, '', path, process)
  return result(context, last.gender === 'male' ? '来孙' : '来孙女', true, true, '', path, process)
}

function resolveChildSpouse(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const child = path[0]

  if (child.gender === 'male') return result(context, '儿媳', true, true, '', path, process)
  if (child.gender === 'female') return result(context, '女婿', true, true, '', path, process)
  return pending(context, '子女配偶', '请补充第一层子女的性别，用于区分儿媳或女婿。', path, process)
}

function resolveGrandChildSpouse(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const child = path[0]
  const grandChild = path[1]

  if (!known(child.gender) || !known(grandChild.gender)) {
    return pending(context, '孙辈配偶', '请补充第一层子女和第二层孙辈的性别。', path, process)
  }

  if (child.gender === 'male' && grandChild.gender === 'male') return result(context, '孙媳', true, true, '', path, process)
  if (child.gender === 'male' && grandChild.gender === 'female') return result(context, '孙女婿', true, true, '', path, process)
  if (child.gender === 'female' && grandChild.gender === 'male') return result(context, '外孙媳', true, true, '', path, process)
  return result(context, '外孙女婿', true, true, '', path, process)
}

function fallback(context: KinshipContext, path: KinshipStep[], process: string[]) {
  const key = pathKey(path)
  return pending(context, '暂未覆盖的亲属关系', `当前路径「${key}」尚未配置精确称谓规则。`, path, process)
}

export function resolveKinship(context: KinshipContext): KinshipResult {
  const raw = Array.isArray(context.steps) ? context.steps : context.path
  const rawLength = Array.isArray(raw) ? raw.length : 0
  const path = getSteps(context)
  const key = pathKey(path)
  const process = [
    `输入路径：${pathText(path)}`,
    `基础关系序列：${key || 'self'}`,
    `五层限制：当前 ${path.length}/${MAX_KINSHIP_LEVEL} 层`
  ]

  if (rawLength > MAX_KINSHIP_LEVEL) {
    return result(context, '超出五层关系', false, true, '最多只允许推导 5 层关系，请删除多余层级。', path, process)
  }

  if (!path.length) {
    return result(context, '本人', true, false, '', path, process)
  }

  if (path.length === 1) {
    const step = path[0]
    if (step.relation === 'parent') return directParent(context, step, path, process)
    if (step.relation === 'child') return directChild(context, step, path, process)
    if (step.relation === 'sibling') return directSibling(context, step, path, process)
    if (step.relation === 'spouse') return directSpouse(context, path, process)
  }

  if (containsOnly(path, 'parent')) return resolveAncestor(context, path, process)
  if (containsOnly(path, 'child')) return resolveDescendant(context, path, process)

  if (key === 'parent/sibling') return resolveParentSibling(context, path, process)
  if (key === 'parent/sibling/spouse') return resolveParentSiblingSpouse(context, path, process)
  if (key === 'parent/sibling/spouse/child') return resolveParentSiblingSpouseChild(context, path, process)
  if (key === 'parent/sibling/child') return resolveCousin(context, path, process)
  if (key === 'parent/sibling/child/spouse') return resolveCousinSpouse(context, path, process)

  if (/^(parent\/)+parent\/sibling$/.test(key)) return resolveAncestorSibling(context, path, process)
  if (/^(parent\/)+parent\/sibling\/spouse$/.test(key)) return resolveAncestorSiblingSpouse(context, path, process)

  if (key === 'spouse/parent') return resolveSpouseParent(context, path, process)
  if (key === 'spouse/sibling') return resolveSpouseSibling(context, path, process)
  if (key === 'spouse/sibling/spouse') return resolveSpouseSiblingSpouse(context, path, process)

  if (key === 'sibling/spouse') return resolveSiblingSpouse(context, path, process)
  if (key === 'sibling/child') return resolveSiblingChild(context, path, process)
  if (key === 'sibling/child/spouse') return resolveSiblingChildSpouse(context, path, process)

  if (key === 'child/spouse') return resolveChildSpouse(context, path, process)
  if (key === 'child/child/spouse') return resolveGrandChildSpouse(context, path, process)

  return fallback(context, path, process)
}
